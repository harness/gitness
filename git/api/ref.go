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

package api

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"

	"github.com/harness/gitness/errors"
	"github.com/harness/gitness/git/api/foreachref"
	"github.com/harness/gitness/git/command"
	"github.com/harness/gitness/git/sha"
)

// GitReferenceField represents the different fields available When listing references.
// For the full list, see https://git-scm.com/docs/git-for-each-ref#_field_names
type GitReferenceField string

const (
	GitReferenceFieldRefName     GitReferenceField = "refname"
	GitReferenceFieldObjectType  GitReferenceField = "objecttype"
	GitReferenceFieldObjectName  GitReferenceField = "objectname"
	GitReferenceFieldCreatorDate GitReferenceField = "creatordate"
)

func ParseGitReferenceField(f string) (GitReferenceField, error) {
	switch f {
	case string(GitReferenceFieldCreatorDate):
		return GitReferenceFieldCreatorDate, nil
	case string(GitReferenceFieldRefName):
		return GitReferenceFieldRefName, nil
	case string(GitReferenceFieldObjectName):
		return GitReferenceFieldObjectName, nil
	case string(GitReferenceFieldObjectType):
		return GitReferenceFieldObjectType, nil
	default:
		return GitReferenceFieldRefName, fmt.Errorf("unknown git reference field '%s'", f)
	}
}

type WalkInstruction int

const (
	WalkInstructionStop WalkInstruction = iota
	WalkInstructionHandle
	WalkInstructionSkip
)

type WalkReferencesEntry map[GitReferenceField]string

// TODO: can be generic (so other walk methods can use the same)
type WalkReferencesInstructor func(WalkReferencesEntry) (WalkInstruction, error)

// TODO: can be generic (so other walk methods can use the same)
type WalkReferencesHandler func(WalkReferencesEntry) error

type WalkReferencesOptions struct {
	// Patterns are the patterns used to pre-filter the references of the repo.
	// OPTIONAL. By default all references are walked.
	Patterns []string

	// Fields indicates the fields that are passed to the instructor & handler
	// OPTIONAL. Default fields are:
	// - GitReferenceFieldRefName
	// - GitReferenceFieldObjectName
	Fields []GitReferenceField

	// Instructor indicates on how to handle the reference.
	// OPTIONAL. By default all references are handled.
	// NOTE: once walkInstructionStop is returned, the walking stops.
	Instructor WalkReferencesInstructor

	// Sort indicates the field by which the references should be sorted.
	// OPTIONAL. By default GitReferenceFieldRefName is used.
	Sort GitReferenceField

	// Order indicates the Order (asc or desc) of the sorted output
	Order SortOrder

	// MaxWalkDistance is the maximum number of nodes that are iterated over before the walking stops.
	// OPTIONAL. A value of <= 0 will walk all references.
	// WARNING: Skipped elements count towards the walking distance
	MaxWalkDistance int32
}

func DefaultInstructor(
	_ WalkReferencesEntry,
) (WalkInstruction, error) {
	return WalkInstructionHandle, nil
}

// WalkReferences uses the provided options to filter the available references of the repo,
// and calls the handle function for every matching node.
// The instructor & handler are called with a map that contains the matching value for every field provided in fields.
// TODO: walkReferences related code should be moved to separate file.
func (g *Git) WalkReferences(
	ctx context.Context,
	repoPath string,
	handler WalkReferencesHandler,
	opts *WalkReferencesOptions,
) error {
	if repoPath == "" {
		return ErrRepositoryPathEmpty
	}
	// backfil optional options
	if opts.Instructor == nil {
		opts.Instructor = DefaultInstructor
	}
	if len(opts.Fields) == 0 {
		opts.Fields = []GitReferenceField{GitReferenceFieldRefName, GitReferenceFieldObjectName}
	}
	if opts.MaxWalkDistance <= 0 {
		opts.MaxWalkDistance = math.MaxInt32
	}
	if opts.Patterns == nil {
		opts.Patterns = []string{}
	}
	if string(opts.Sort) == "" {
		opts.Sort = GitReferenceFieldRefName
	}

	// prepare for-each-ref input
	sortArg := mapToReferenceSortingArgument(opts.Sort, opts.Order)
	rawFields := make([]string, len(opts.Fields))
	for i := range opts.Fields {
		rawFields[i] = string(opts.Fields[i])
	}
	format := foreachref.NewFormat(rawFields...)

	// initializer pipeline for output processing
	pipeOut, pipeIn := io.Pipe()
	defer pipeOut.Close()

	go func() {
		cmd := command.New("for-each-ref",
			command.WithFlag("--format", format.Flag()),
			command.WithFlag("--sort", sortArg),
			command.WithFlag("--count", strconv.Itoa(int(opts.MaxWalkDistance))),
			command.WithFlag("--ignore-case"),
		)
		cmd.Add(command.WithArg(opts.Patterns...))
		err := cmd.Run(ctx,
			command.WithDir(repoPath),
			command.WithStdout(pipeIn),
		)
		if err != nil {
			_ = pipeIn.CloseWithError(err)
		} else {
			_ = pipeIn.Close()
		}
	}()

	// TODO: return error from git command!!!!

	parser := format.Parser(pipeOut)
	return walkReferenceParser(parser, handler, opts)
}

func walkReferenceParser(
	parser *foreachref.Parser,
	handler WalkReferencesHandler,
	opts *WalkReferencesOptions,
) error {
	for i := int32(0); i < opts.MaxWalkDistance; i++ {
		// parse next line - nil if end of output reached or an error occurred.
		rawRef := parser.Next()
		if rawRef == nil {
			break
		}

		// convert to correct map.
		ref, err := mapRawRef(rawRef)
		if err != nil {
			return err
		}

		// check with the instructor on the next instruction.
		instruction, err := opts.Instructor(ref)
		if err != nil {
			return fmt.Errorf("error getting instruction: %w", err)
		}

		if instruction == WalkInstructionSkip {
			continue
		}
		if instruction == WalkInstructionStop {
			break
		}

		// otherwise handle the reference.
		err = handler(ref)
		if err != nil {
			return fmt.Errorf("error handling reference: %w", err)
		}
	}

	if err := parser.Err(); err != nil {
		return processGitErrorf(err, "failed to parse reference walk output")
	}

	return nil
}

// GetRef get's the target of a reference
// IMPORTANT provide full reference name to limit risk of collisions across reference types
// (e.g `refs/heads/main` instead of `main`).
func (g *Git) GetRef(
	ctx context.Context,
	repoPath string,
	ref string,
) (sha.SHA, error) {
	if repoPath == "" {
		return sha.None, ErrRepositoryPathEmpty
	}
	cmd := command.New("show-ref",
		command.WithFlag("--verify"),
		command.WithFlag("-s"),
		command.WithArg(ref),
	)
	output := &bytes.Buffer{}
	err := cmd.Run(ctx, command.WithDir(repoPath), command.WithStdout(output))
	if err != nil {
		if command.AsError(err).IsExitCode(128) && strings.Contains(err.Error(), "not a valid ref") {
			return sha.None, errors.NotFound("reference %q not found", ref)
		}
		return sha.None, err
	}

	return sha.New(output.String())
}

// GetReferenceFromBranchName assumes the provided value is the branch name (not the ref!)
// and first sanitizes the branch name (remove any spaces or 'refs/heads/' prefix)
// It then returns the full form of the branch reference.
func GetReferenceFromBranchName(branchName string) string {
	// remove spaces
	branchName = strings.TrimSpace(branchName)
	// remove `refs/heads/` prefix (shouldn't be there, but if it is remove it to try to avoid complications)
	// NOTE: This is used to reduce misconfigurations via api
	// TODO: block via CLI, too
	branchName = strings.TrimPrefix(branchName, gitReferenceNamePrefixBranch)

	// return reference
	return gitReferenceNamePrefixBranch + branchName
}

func GetReferenceFromTagName(tagName string) string {
	// remove spaces
	tagName = strings.TrimSpace(tagName)
	// remove `refs/heads/` prefix (shouldn't be there, but if it is remove it to try to avoid complications)
	// NOTE: This is used to reduce misconfigurations via api
	// TODO: block via CLI, too
	tagName = strings.TrimPrefix(tagName, gitReferenceNamePrefixTag)

	// return reference
	return gitReferenceNamePrefixTag + tagName
}
