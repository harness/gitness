// Copyright 2023 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package git

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"runtime/debug"
	"time"

	"github.com/harness/gitness/errors"
	"github.com/harness/gitness/git/api"
	"github.com/harness/gitness/git/check"
	"github.com/harness/gitness/git/hash"

	gonanoid "github.com/matoous/go-nanoid/v2"
	"github.com/rs/zerolog/log"
)

const (
	// repoGitUIDLength is the length of the generated repo uid.
	repoGitUIDLength = 42

	// repoGitUIDAlphabet is the alphabet used for generating repo uids
	// NOTE: keep it lowercase and alphanumerical to avoid issues with case insensitive filesystems.
	repoGitUIDAlphabet = "abcdefghijklmnopqrstuvwxyz0123456789"

	gitReferenceNamePrefixBranch = "refs/heads/"
	gitReferenceNamePrefixTag    = "refs/tags/"

	gitHooksDir = "hooks"

	fileMode700 = 0o700
)

var (
	gitServerHookNames = []string{
		"pre-receive",
		// "update", // update is disabled for performance reasons (called once for every ref)
		"post-receive",
	}
)

type CreateRepositoryParams struct {
	// Create operation is different from all (from user side), as UID doesn't exist yet.
	// Only take actor and envars as input and create WriteParams manually
	RepoUID string
	Actor   Identity
	EnvVars map[string]string

	DefaultBranch string
	Files         []File

	// Committer overwrites the git committer used for committing the files
	// (optional, default: actor)
	Committer *Identity
	// CommitterDate overwrites the git committer date used for committing the files
	// (optional, default: current time on server)
	CommitterDate *time.Time
	// Author overwrites the git author used for committing the files
	// (optional, default: committer)
	Author *Identity
	// AuthorDate overwrites the git author date used for committing the files
	// (optional, default: committer date)
	AuthorDate *time.Time
}

func (p *CreateRepositoryParams) Validate() error {
	return p.Actor.Validate()
}

type CreateRepositoryOutput struct {
	UID string
}

type DeleteRepositoryParams struct {
	WriteParams
}

type GetRepositorySizeParams struct {
	ReadParams
}

type GetRepositorySizeOutput struct {
	// Total size of the repository in KiB.
	Size int64
}

type SyncRepositoryParams struct {
	WriteParams
	Source            string
	CreateIfNotExists bool

	// RefSpecs [OPTIONAL] allows to override the refspecs that are being synced from the remote repository.
	// By default all references present on the remote repository will be fetched (including scm internal ones).
	RefSpecs []string
}

type SyncRepositoryOutput struct {
	DefaultBranch string
}

type HashRepositoryParams struct {
	ReadParams
	HashType        hash.Type
	AggregationType hash.AggregationType
}

func (p *HashRepositoryParams) Validate() error {
	return p.ReadParams.Validate()
}

type HashRepositoryOutput struct {
	Hash []byte
}
type UpdateDefaultBranchParams struct {
	WriteParams
	// BranchName is the name of the branch
	BranchName string
}

func (s *Service) CreateRepository(
	ctx context.Context,
	params *CreateRepositoryParams,
) (*CreateRepositoryOutput, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}

	log := log.Ctx(ctx)

	if params.RepoUID == "" {
		uid, err := NewRepositoryUID()
		if err != nil {
			return nil, fmt.Errorf("failed to create new uid: %w", err)
		}
		params.RepoUID = uid
	}
	log.Info().
		Msgf("Create new git repository with uid '%s' and default branch '%s'", params.RepoUID, params.DefaultBranch)

	writeParams := WriteParams{
		RepoUID: params.RepoUID,
		Actor:   params.Actor,
		EnvVars: params.EnvVars,
	}

	committer := params.Actor
	if params.Committer != nil {
		committer = *params.Committer
	}
	committerDate := time.Now().UTC()
	if params.CommitterDate != nil {
		committerDate = *params.CommitterDate
	}

	author := committer
	if params.Author != nil {
		author = *params.Author
	}
	authorDate := committerDate
	if params.AuthorDate != nil {
		authorDate = *params.AuthorDate
	}

	err := s.createRepositoryInternal(
		ctx,
		&writeParams,
		params.DefaultBranch,
		params.Files,
		&committer,
		committerDate,
		&author,
		authorDate,
	)
	if err != nil {
		return nil, err
	}

	return &CreateRepositoryOutput{
		UID: params.RepoUID,
	}, nil
}

func NewRepositoryUID() (string, error) {
	return gonanoid.Generate(repoGitUIDAlphabet, repoGitUIDLength)
}

func (s *Service) DeleteRepository(ctx context.Context, params *DeleteRepositoryParams) error {
	if err := params.Validate(); err != nil {
		return err
	}
	repoPath := getFullPathForRepo(s.reposRoot, params.RepoUID)

	if _, err := os.Stat(repoPath); err != nil && os.IsNotExist(err) {
		return errors.NotFound("repository path not found")
	} else if err != nil {
		return fmt.Errorf("failed to check the status of the repository %v: %w", repoPath, err)
	}

	return s.DeleteRepositoryBestEffort(ctx, params.RepoUID)
}

func (s *Service) DeleteRepositoryBestEffort(ctx context.Context, repoUID string) error {
	repoPath := getFullPathForRepo(s.reposRoot, repoUID)
	tempPath := path.Join(s.reposGraveyard, repoUID)

	// delete should not fail if repoGraveyard dir does not exist.
	if _, err := os.Stat(s.reposGraveyard); os.IsNotExist(err) {
		if errdir := os.MkdirAll(s.reposGraveyard, fileMode700); errdir != nil {
			return fmt.Errorf("clean up dir '%s' doesn't exist and can't be created: %w", s.reposGraveyard, errdir)
		}
	}
	// move current dir to a temp dir (prevent partial deletion)
	if err := os.Rename(repoPath, tempPath); err != nil {
		return fmt.Errorf("couldn't move dir %s to %s : %w", repoPath, tempPath, err)
	}

	if err := os.RemoveAll(tempPath); err != nil {
		log.Ctx(ctx).Warn().Err(err).Msgf("failed to delete dir %s from graveyard", tempPath)
	}
	return nil
}

func (s *Service) SyncRepository(
	ctx context.Context,
	params *SyncRepositoryParams,
) (*SyncRepositoryOutput, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}

	repoPath := getFullPathForRepo(s.reposRoot, params.RepoUID)

	// create repo if requested
	_, err := os.Stat(repoPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, errors.Internal(err, "failed to create repository")
	}

	if os.IsNotExist(err) {
		if !params.CreateIfNotExists {
			return nil, errors.NotFound("repository not found")
		}

		// the default branch doesn't matter for a sync,
		// we create an empty repo and the head will by updated as part of the Sync.
		const syncDefaultBranch = "main"
		if err = s.createRepositoryInternal(
			ctx,
			&params.WriteParams,
			syncDefaultBranch,
			nil,
			nil,
			time.Time{},
			nil,
			time.Time{},
		); err != nil {
			return nil, err
		}
	}

	// sync repo content
	err = s.git.Sync(ctx, repoPath, params.Source, params.RefSpecs)
	if err != nil {
		return nil, fmt.Errorf("SyncRepository: failed to sync git repo: %w", err)
	}

	// get remote default branch
	defaultBranch, err := s.git.GetRemoteDefaultBranch(ctx, params.Source)
	if errors.Is(err, api.ErrNoDefaultBranch) {
		return &SyncRepositoryOutput{
			DefaultBranch: "",
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("SyncRepository: failed to get default branch from repo: %w", err)
	}

	// set default branch
	err = s.git.SetDefaultBranch(ctx, repoPath, defaultBranch, true)
	if err != nil {
		return nil, fmt.Errorf("SyncRepository: failed to set default branch of repo: %w", err)
	}

	return &SyncRepositoryOutput{
		DefaultBranch: defaultBranch,
	}, nil
}

func (s *Service) HashRepository(ctx context.Context, params *HashRepositoryParams) (*HashRepositoryOutput, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}

	repoPath := getFullPathForRepo(s.reposRoot, params.RepoUID)

	// add all references of the repo to the channel in a separate go routine, to allow streamed processing.
	// Ensure we cancel the go routine in case we exit the func early.
	goCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	hashChan := make(chan hash.SourceNext)

	go func() {
		// always close channel last before leaving go routine
		defer close(hashChan)
		defer func() {
			if r := recover(); r != nil {
				hashChan <- hash.SourceNext{
					Err: fmt.Errorf("panic received while filling data source: %s", debug.Stack()),
				}
			}
		}()

		// add default branch to hash
		defaultBranch, err := s.git.GetDefaultBranch(goCtx, repoPath)
		if err != nil {
			hashChan <- hash.SourceNext{
				Err: fmt.Errorf("HashRepository: failed to get default branch: %w", err),
			}
			return
		}

		hashChan <- hash.SourceNext{
			Data: hash.SerializeHead(defaultBranch),
		}

		err = s.git.WalkReferences(goCtx, repoPath, func(wre api.WalkReferencesEntry) error {
			ref, ok := wre[api.GitReferenceFieldRefName]
			if !ok {
				return errors.New("ref entry didn't contain the ref name")
			}
			sha, ok := wre[api.GitReferenceFieldObjectName]
			if !ok {
				return errors.New("ref entry didn't contain the ref object sha")
			}

			hashChan <- hash.SourceNext{
				Data: hash.SerializeReference(ref, sha),
			}

			return nil
		}, &api.WalkReferencesOptions{})
		if err != nil {
			hashChan <- hash.SourceNext{
				Err: fmt.Errorf("HashRepository: failed to walk references: %w", err),
			}
		}
	}()

	hasher, err := hash.New(params.HashType, params.AggregationType)
	if err != nil {
		return nil, fmt.Errorf("HashRepository: failed to get new reference hasher: %w", err)
	}
	source := hash.SourceFromChannel(ctx, hashChan)

	res, err := hasher.Hash(source)
	if err != nil {
		return nil, fmt.Errorf("HashRepository: failed to hash repository: %w", err)
	}

	return &HashRepositoryOutput{
		Hash: res,
	}, nil
}

//nolint:gocognit // refactor if needed
func (s *Service) createRepositoryInternal(
	ctx context.Context,
	base *WriteParams,
	defaultBranch string,
	files []File,
	committer *Identity,
	committerDate time.Time,
	author *Identity,
	authorDate time.Time,
) error {
	log := log.Ctx(ctx)
	repoPath := getFullPathForRepo(s.reposRoot, base.RepoUID)
	if _, err := os.Stat(repoPath); !os.IsNotExist(err) {
		return errors.Conflict("repository already exists at path %q", repoPath)
	}

	// create repository in repos folder
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	err := s.git.InitRepository(ctx, repoPath, true)
	// delete repo dir on error
	defer func() {
		if err != nil {
			cleanuperr := s.DeleteRepositoryBestEffort(ctx, base.RepoUID)
			if cleanuperr != nil {
				log.Warn().Err(cleanuperr).Msg("failed to cleanup repo dir")
			}
		}
	}()

	if err != nil {
		return fmt.Errorf("createRepositoryInternal: failed to initialize the repository: %w", err)
	}

	// update default branch (currently set to non-existent branch)
	err = s.git.SetDefaultBranch(ctx, repoPath, defaultBranch, true)
	if err != nil {
		return fmt.Errorf("createRepositoryInternal: error updating default branch for repo '%s': %w",
			base.RepoUID, err)
	}

	// only execute file creation logic if files are provided
	//nolint: nestif

	// we need temp dir for cloning
	tempDir, err := os.MkdirTemp("", "*-"+base.RepoUID)
	if err != nil {
		return fmt.Errorf("error creating temp dir for repo %s: %w", base.RepoUID, err)
	}
	defer func(path string) {
		// when repo is successfully created remove temp dir
		errRm := os.RemoveAll(path)
		if errRm != nil {
			log.Err(errRm).Msg("failed to cleanup temporary dir.")
		}
	}(tempDir)

	// Clone repository to temp dir
	if err = s.git.Clone(ctx, repoPath, tempDir, api.CloneRepoOptions{}); err != nil {
		return fmt.Errorf("createRepositoryInternal: failed to clone repo: %w", err)
	}

	// logic for receiving files
	filePaths := make([]string, 0, 16)
	for _, file := range files {
		var filePath string
		filePath, err = s.handleFileUploadIfAvailable(ctx, tempDir, file)
		if errors.Is(err, io.EOF) {
			log.Info().Msg("received stream EOF")
			break
		}
		if err != nil {
			return errors.Internal(err, "failed to receive file %s", file)
		}

		filePaths = append(filePaths, filePath)
	}

	if len(filePaths) > 0 {
		if committer == nil {
			committer = &base.Actor
		}
		if author == nil {
			author = committer
		}
		// NOTE: This creates the branch in origin repo (as it doesn't exist as of now)
		// TODO: this should at least be a constant and not hardcoded?
		if err = s.addFilesAndPush(ctx,
			tempDir,
			filePaths, "HEAD:"+defaultBranch,
			nil,
			author,
			authorDate,
			committer,
			committerDate,
			"origin",
			"initial commit",
		); err != nil {
			return err
		}
	}

	// setup server hook symlinks pointing to configured server hook binary
	// IMPORTANT: Setup hooks after repo creation to avoid issues with externally dependent services.
	for _, hook := range gitServerHookNames {
		hookPath := path.Join(repoPath, gitHooksDir, hook)
		err = os.Symlink(s.gitHookPath, hookPath)
		if err != nil {
			return errors.Internal(err, "failed to setup symlink for hook '%s' ('%s' -> '%s')",
				hook, hookPath, s.gitHookPath)
		}
	}

	log.Info().Msgf("repository created. Path: %s", repoPath)
	return nil
}

// GetRepositorySize accumulates the sizes of counted Git objects.
func (s *Service) GetRepositorySize(
	ctx context.Context,
	params *GetRepositorySizeParams,
) (*GetRepositorySizeOutput, error) {
	repoPath := getFullPathForRepo(s.reposRoot, params.RepoUID)
	count, err := s.git.CountObjects(ctx, repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to count objects for repo: %w", err)
	}

	return &GetRepositorySizeOutput{
		Size: count.Size + count.SizePack,
	}, nil
}

// UpdateDefaultBranch updates the default branch of the repo.
func (s *Service) UpdateDefaultBranch(
	ctx context.Context,
	params *UpdateDefaultBranchParams,
) error {
	if err := params.Validate(); err != nil {
		return err
	}
	if err := check.BranchName(params.BranchName); err != nil {
		return errors.InvalidArgument(err.Error())
	}

	repoPath := getFullPathForRepo(s.reposRoot, params.RepoUID)

	err := s.git.SetDefaultBranch(ctx, repoPath, params.BranchName, false)
	if err != nil {
		return fmt.Errorf("UpdateDefaultBranch: failed to update repo default branch %q: %w",
			params.BranchName, err)
	}
	return nil
}

type ArchiveParams struct {
	ReadParams
	api.ArchiveParams
}

func (p *ArchiveParams) Validate() error {
	if err := p.ReadParams.Validate(); err != nil {
		return err
	}
	//nolint:revive
	if err := p.ArchiveParams.Validate(); err != nil {
		return err
	}
	return nil
}

func (s *Service) Archive(ctx context.Context, params ArchiveParams, w io.Writer) error {
	if err := params.Validate(); err != nil {
		return err
	}
	repoPath := getFullPathForRepo(s.reposRoot, params.RepoUID)
	err := s.git.Archive(ctx, repoPath, params.ArchiveParams, w)
	if err != nil {
		return fmt.Errorf("failed to run git archive: %w", err)
	}

	return nil
}
