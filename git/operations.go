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
	"bytes"
	"context"
	"fmt"
	"path"
	"strings"
	"time"

	"github.com/harness/gitness/errors"
	"github.com/harness/gitness/git/api"
	"github.com/harness/gitness/git/sha"

	"github.com/rs/zerolog/log"
	"golang.org/x/exp/slices"
)

const (
	defaultFilePermission = "100644" // 0o644 default file permission
)

type FileAction string

const (
	CreateAction FileAction = "CREATE"
	UpdateAction FileAction = "UPDATE"
	DeleteAction            = "DELETE"
	MoveAction              = "MOVE"
)

func (FileAction) Enum() []interface{} {
	return []interface{}{CreateAction, UpdateAction, DeleteAction, MoveAction}
}

// CommitFileAction holds file operation data.
type CommitFileAction struct {
	Action  FileAction
	Path    string
	Payload []byte
	SHA     string
}

// CommitFilesParams holds the data for file operations.
type CommitFilesParams struct {
	WriteParams
	Title     string
	Message   string
	Branch    string
	NewBranch string
	Actions   []CommitFileAction

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

func (p *CommitFilesParams) Validate() error {
	return p.WriteParams.Validate()
}

type CommitFilesResponse struct {
	CommitID sha.SHA
}

//nolint:gocognit
func (s *Service) CommitFiles(ctx context.Context, params *CommitFilesParams) (CommitFilesResponse, error) {
	if err := params.Validate(); err != nil {
		return CommitFilesResponse{}, err
	}

	log := log.Ctx(ctx).With().Str("repo_uid", params.RepoUID).Logger()

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

	repoPath := getFullPathForRepo(s.reposRoot, params.RepoUID)

	log.Debug().Msg("check if empty")

	// check if repo is empty
	// IMPORTANT: we don't use gitea's repo.IsEmpty() as that only checks whether the default branch exists (in HEAD).
	// This can be an issue in case someone created a branch already in the repo (just default branch is missing).
	// In that case the user can accidentally create separate git histories (which most likely is unintended).
	// If the user wants to actually build a disconnected commit graph they can use the cli.
	isEmpty, err := s.git.HasBranches(ctx, repoPath)
	if err != nil {
		return CommitFilesResponse{}, fmt.Errorf("CommitFiles: failed to determine if repository is empty: %w", err)
	}

	log.Debug().Msg("validate and prepare input")

	// ensure input data is valid
	// the commit will be nil for empty repositories
	commit, err := s.validateAndPrepareHeader(ctx, repoPath, isEmpty, params)
	if err != nil {
		return CommitFilesResponse{}, err
	}

	var oldCommitSHA sha.SHA
	if commit != nil {
		oldCommitSHA = commit.SHA
	}

	log.Debug().Msg("create shared repo")

	newCommitSHA, err := func() (sha.SHA, error) {
		// Create a directory for the temporary shared repository.
		shared, err := api.NewSharedRepo(s.git, s.tmpDir, params.RepoUID, repoPath)
		if err != nil {
			return sha.None, fmt.Errorf("failed to create shared repository: %w", err)
		}
		defer shared.Close(ctx)

		// Create bare repository with alternates pointing to the original repository.
		err = shared.InitAsShared(ctx)
		if err != nil {
			return sha.None, fmt.Errorf("failed to create temp repo with alternates: %w", err)
		}

		log.Debug().Msgf("prepare tree (empty: %t)", isEmpty)

		// handle empty repo separately (as branch doesn't exist, no commit exists, ...)
		if isEmpty {
			err = s.prepareTreeEmptyRepo(ctx, shared, params.Actions)
		} else {
			err = shared.SetIndex(ctx, oldCommitSHA.String())
			if err != nil {
				return sha.None, fmt.Errorf("failed to set index to temp repo: %w", err)
			}

			err = s.prepareTree(ctx, shared, params.Actions, commit)
		}
		if err != nil {
			return sha.None, fmt.Errorf("failed to prepare tree: %w", err)
		}

		log.Debug().Msg("write tree")

		// Now write the tree
		treeHash, err := shared.WriteTree(ctx)
		if err != nil {
			return sha.None, fmt.Errorf("failed to write tree object: %w", err)
		}

		message := strings.TrimSpace(params.Title)
		if len(params.Message) > 0 {
			message += "\n\n" + strings.TrimSpace(params.Message)
		}

		log.Debug().Msg("commit tree")

		// Now commit the tree
		commitSHA, err := shared.CommitTreeWithDate(
			ctx,
			oldCommitSHA,
			&api.Identity{
				Name:  author.Name,
				Email: author.Email,
			},
			&api.Identity{
				Name:  committer.Name,
				Email: committer.Email,
			},
			treeHash,
			message,
			false,
			authorDate,
			committerDate,
		)
		if err != nil {
			return sha.None, fmt.Errorf("failed to commit the tree: %w", err)
		}

		err = shared.MoveObjects(ctx)
		if err != nil {
			return sha.None, fmt.Errorf("failed to move git objects: %w", err)
		}

		return commitSHA, nil
	}()
	if err != nil {
		return CommitFilesResponse{}, fmt.Errorf("CommitFiles: failed to create commit in shared repository: %w", err)
	}

	log.Debug().Msg("update ref")

	branchRef := api.GetReferenceFromBranchName(params.Branch)
	if params.Branch != params.NewBranch {
		// we are creating a new branch, rather than updating the existing one
		oldCommitSHA = sha.Nil
		branchRef = api.GetReferenceFromBranchName(params.NewBranch)
	}
	err = s.git.UpdateRef(
		ctx,
		params.EnvVars,
		repoPath,
		branchRef,
		oldCommitSHA,
		newCommitSHA,
	)
	if err != nil {
		return CommitFilesResponse{}, fmt.Errorf("CommitFiles: failed to update ref %s: %w", branchRef, err)
	}

	log.Debug().Msg("get commit")

	commit, err = s.git.GetCommit(ctx, repoPath, newCommitSHA.String())
	if err != nil {
		return CommitFilesResponse{}, fmt.Errorf("failed to get commit for SHA %s: %w",
			newCommitSHA.String(), err)
	}

	log.Debug().Msg("done")

	return CommitFilesResponse{
		CommitID: commit.SHA,
	}, nil
}

func (s *Service) prepareTree(
	ctx context.Context,
	shared *api.SharedRepo,
	actions []CommitFileAction,
	commit *api.Commit,
) error {
	// execute all actions
	for i := range actions {
		if err := s.processAction(ctx, shared, &actions[i], commit); err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) prepareTreeEmptyRepo(
	ctx context.Context,
	shared *api.SharedRepo,
	actions []CommitFileAction,
) error {
	for _, action := range actions {
		if action.Action != CreateAction {
			return errors.PreconditionFailed("action not allowed on empty repository")
		}

		filePath := api.CleanUploadFileName(action.Path)
		if filePath == "" {
			return errors.InvalidArgument("invalid path")
		}

		if err := createFile(ctx, shared, nil, filePath, defaultFilePermission, action.Payload); err != nil {
			return errors.Internal(err, "failed to create file '%s'", action.Path)
		}
	}

	return nil
}

func (s *Service) validateAndPrepareHeader(
	ctx context.Context,
	repoPath string,
	isEmpty bool,
	params *CommitFilesParams,
) (*api.Commit, error) {
	if params.Branch == "" {
		defaultBranchRef, err := s.git.GetDefaultBranch(ctx, repoPath)
		if err != nil {
			return nil, fmt.Errorf("failed to get default branch: %w", err)
		}
		params.Branch = defaultBranchRef
	}

	if params.NewBranch == "" {
		params.NewBranch = params.Branch
	}

	// trim refs/heads/ prefixes to avoid issues when calling gitea API
	params.Branch = strings.TrimPrefix(strings.TrimSpace(params.Branch), gitReferenceNamePrefixBranch)
	params.NewBranch = strings.TrimPrefix(strings.TrimSpace(params.NewBranch), gitReferenceNamePrefixBranch)

	// if the repo is empty then we can skip branch existence checks
	if isEmpty {
		return nil, nil //nolint:nilnil // an empty repository has no commit and there's no error
	}

	// ensure source branch exists
	branch, err := s.git.GetBranch(ctx, repoPath, params.Branch)
	if err != nil {
		return nil, fmt.Errorf("failed to get source branch '%s': %w", params.Branch, err)
	}

	// ensure new branch doesn't exist yet (if new branch creation was requested)
	if params.Branch != params.NewBranch {
		existingBranch, err := s.git.GetBranch(ctx, repoPath, params.NewBranch)
		if existingBranch != nil {
			return nil, errors.Conflict("branch %s already exists", existingBranch.Name)
		}
		if err != nil && !errors.IsNotFound(err) {
			return nil, fmt.Errorf("failed to create new branch '%s': %w", params.NewBranch, err)
		}
	}

	return branch.Commit, nil
}

func (s *Service) processAction(
	ctx context.Context,
	shared *api.SharedRepo,
	action *CommitFileAction,
	commit *api.Commit,
) (err error) {
	filePath := api.CleanUploadFileName(action.Path)
	if filePath == "" {
		return errors.InvalidArgument("path cannot be empty")
	}

	switch action.Action {
	case CreateAction:
		err = createFile(ctx, shared, commit, filePath, defaultFilePermission, action.Payload)
	case UpdateAction:
		err = updateFile(ctx, shared, commit, filePath, action.SHA, defaultFilePermission, action.Payload)
	case MoveAction:
		err = moveFile(ctx, shared, commit, filePath, action.SHA, defaultFilePermission, action.Payload)
	case DeleteAction:
		err = deleteFile(ctx, shared, filePath)
	}

	return err
}

func createFile(ctx context.Context, repo *api.SharedRepo, commit *api.Commit,
	filePath, mode string, payload []byte) error {
	// only check path availability if a source commit is available (empty repo won't have such a commit)
	if commit != nil {
		if err := checkPathAvailability(ctx, repo, commit, filePath, true); err != nil {
			return err
		}
	}

	hash, err := repo.WriteGitObject(ctx, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("createFile: error hashing object: %w", err)
	}

	// Add the object to the index
	if err = repo.AddObjectToIndex(ctx, mode, hash.String(), filePath); err != nil {
		return fmt.Errorf("createFile: error creating object: %w", err)
	}
	return nil
}

func updateFile(
	ctx context.Context,
	repo *api.SharedRepo,
	commit *api.Commit,
	filePath string,
	sha string,
	mode string,
	payload []byte,
) error {
	// get file mode from existing file (default unless executable)
	entry, err := getFileEntry(ctx, repo, commit, sha, filePath)
	if err != nil {
		return err
	}
	if entry.IsExecutable() {
		mode = "100755"
	}

	hash, err := repo.WriteGitObject(ctx, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("updateFile: error hashing object: %w", err)
	}

	if err = repo.AddObjectToIndex(ctx, mode, hash.String(), filePath); err != nil {
		return fmt.Errorf("updateFile: error updating object: %w", err)
	}
	return nil
}

func moveFile(
	ctx context.Context,
	repo *api.SharedRepo,
	commit *api.Commit,
	filePath string,
	sha string,
	mode string,
	payload []byte,
) error {
	newPath, newContent, err := parseMovePayload(payload)
	if err != nil {
		return err
	}

	// ensure file exists and matches SHA
	entry, err := getFileEntry(ctx, repo, commit, sha, filePath)
	if err != nil {
		return err
	}

	// ensure new path is available
	if err = checkPathAvailability(ctx, repo, commit, newPath, false); err != nil {
		return err
	}

	var fileHash string
	var fileMode string
	if newContent != nil {
		hash, err := repo.WriteGitObject(ctx, bytes.NewReader(newContent))
		if err != nil {
			return fmt.Errorf("moveFile: error hashing object: %w", err)
		}

		fileHash = hash.String()
		fileMode = mode
		if entry.IsExecutable() {
			fileMode = "100755"
		}
	} else {
		fileHash = entry.SHA.String()
		fileMode = entry.Mode.String()
	}

	if err = repo.AddObjectToIndex(ctx, fileMode, fileHash, newPath); err != nil {
		return fmt.Errorf("moveFile: add object error: %w", err)
	}

	if err = repo.RemoveFilesFromIndex(ctx, filePath); err != nil {
		return fmt.Errorf("moveFile: remove object error: %w", err)
	}
	return nil
}

func deleteFile(ctx context.Context, repo *api.SharedRepo, filePath string) error {
	filesInIndex, err := repo.LsFiles(ctx, filePath)
	if err != nil {
		return fmt.Errorf("deleteFile: listing files error: %w", err)
	}
	if !slices.Contains(filesInIndex, filePath) {
		return errors.NotFound("file path %s not found", filePath)
	}

	if err = repo.RemoveFilesFromIndex(ctx, filePath); err != nil {
		return fmt.Errorf("deleteFile: remove object error: %w", err)
	}
	return nil
}

func getFileEntry(
	ctx context.Context,
	repo *api.SharedRepo,
	commit *api.Commit,
	sha string,
	path string,
) (*api.TreeNode, error) {
	entry, err := repo.GetTreeNode(ctx, commit.SHA.String(), path)
	if errors.IsNotFound(err) {
		return nil, errors.NotFound("path %s not found", path)
	}
	if err != nil {
		return nil, fmt.Errorf("getFileEntry: failed to get tree for path %s: %w", path, err)
	}

	// If a SHA was given and the SHA given doesn't match the SHA of the fromTreePath, throw error
	if sha != "" && sha != entry.SHA.String() {
		return nil, errors.InvalidArgument("sha does not match for path %s [given: %s, expected: %s]",
			path, sha, entry.SHA)
	}

	return entry, nil
}

// checkPathAvailability ensures that the path is available for the requested operation.
// For the path where this file will be created/updated, we need to make
// sure no parts of the path are existing files or links except for the last
// item in the path which is the file name, and that shouldn't exist IF it is
// a new file OR is being moved to a new path.
func checkPathAvailability(
	ctx context.Context,
	repo *api.SharedRepo,
	commit *api.Commit,
	filePath string,
	isNewFile bool,
) error {
	parts := strings.Split(filePath, "/")
	subTreePath := ""
	for index, part := range parts {
		subTreePath = path.Join(subTreePath, part)
		entry, err := repo.GetTreeNode(ctx, commit.SHA.String(), subTreePath)
		if err != nil {
			if errors.IsNotFound(err) {
				// Means there is no item with that name, so we're good
				break
			}
			return fmt.Errorf("checkPathAvailability: failed to get tree entry for path %s: %w", subTreePath, err)
		}
		switch {
		case index < len(parts)-1:
			if !entry.IsDir() {
				return errors.Conflict("a file already exists where you're trying to create a subdirectory [path: %s]",
					subTreePath)
			}
		case entry.IsLink():
			return errors.Conflict("a symbolic link already exist where you're trying to create a subdirectory [path: %s]",
				subTreePath)
		case entry.IsDir():
			return errors.Conflict("a directory already exists where you're trying to create a subdirectory [path: %s]",
				subTreePath)
		case filePath != "" || isNewFile:
			return errors.Conflict("file path %s already exists", filePath)
		}
	}
	return nil
}

func parseMovePayload(payload []byte) (string, []byte, error) {
	var newContent []byte
	var newPath string
	filePathEnd := bytes.IndexByte(payload, 0)
	if filePathEnd < 0 {
		newPath = string(payload)
		newContent = nil
	} else {
		newPath = string(payload[:filePathEnd])
		newContent = payload[filePathEnd+1:]
	}

	newPath = api.CleanUploadFileName(newPath)
	if newPath == "" {
		return "", nil, api.ErrInvalidPath
	}

	return newPath, newContent, nil
}
