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
	"fmt"
	"strings"

	"github.com/harness/gitness/errors"
	"github.com/harness/gitness/git/enum"

	"github.com/rs/zerolog/log"
)

var (
	ErrInvalidPath         = errors.New("path is invalid")
	ErrRepositoryPathEmpty = errors.InvalidArgument("repository path cannot be empty")
	ErrBranchNameEmpty     = errors.InvalidArgument("branch name cannot be empty")
	ErrParseDiffHunkHeader = errors.Internal(nil, "failed to parse diff hunk header")
	ErrNoDefaultBranch     = errors.New("no default branch")
	ErrInvalidSignature    = errors.New("invalid signature")
)

// PushOutOfDateError represents an error if merging fails due to unrelated histories.
type PushOutOfDateError struct {
	StdOut string
	StdErr string
	Err    error
}

func (err *PushOutOfDateError) Error() string {
	return fmt.Sprintf("PushOutOfDate Error: %v: %s\n%s", err.Err, err.StdErr, err.StdOut)
}

// Unwrap unwraps the underlying error.
func (err *PushOutOfDateError) Unwrap() error {
	return fmt.Errorf("%w - %s", err.Err, err.StdErr)
}

// PushRejectedError represents an error if merging fails due to rejection from a hook.
type PushRejectedError struct {
	Message string
	StdOut  string
	StdErr  string
	Err     error
}

// IsErrPushRejected checks if an error is a PushRejectedError.
func IsErrPushRejected(err error) bool {
	var errPushRejected *PushRejectedError
	return errors.As(err, &errPushRejected)
}

func (err *PushRejectedError) Error() string {
	return fmt.Sprintf("PushRejected Error: %v: %s\n%s", err.Err, err.StdErr, err.StdOut)
}

// Unwrap unwraps the underlying error.
func (err *PushRejectedError) Unwrap() error {
	return fmt.Errorf("%w - %s", err.Err, err.StdErr)
}

// GenerateMessage generates the remote message from the stderr.
func (err *PushRejectedError) GenerateMessage() {
	messageBuilder := &strings.Builder{}
	i := strings.Index(err.StdErr, "remote: ")
	if i < 0 {
		err.Message = ""
		return
	}
	for {
		if len(err.StdErr) <= i+8 {
			break
		}
		if err.StdErr[i:i+8] != "remote: " {
			break
		}
		i += 8
		nl := strings.IndexByte(err.StdErr[i:], '\n')
		if nl >= 0 {
			messageBuilder.WriteString(err.StdErr[i : i+nl+1])
			i = i + nl + 1
		} else {
			messageBuilder.WriteString(err.StdErr[i:])
			i = len(err.StdErr)
		}
	}
	err.Message = strings.TrimSpace(messageBuilder.String())
}

// MoreThanOneError represents an error when there are more
// than one sources (branch, tag) with the same name.
type MoreThanOneError struct {
	StdOut string
	StdErr string
	Err    error
}

// IsErrMoreThanOne checks if an error is a MoreThanOneError.
func IsErrMoreThanOne(err error) bool {
	var errMoreThanOne *MoreThanOneError
	return errors.As(err, &errMoreThanOne)
}

func (err *MoreThanOneError) Error() string {
	return fmt.Sprintf("MoreThanOneError Error: %v: %s\n%s", err.Err, err.StdErr, err.StdOut)
}

// Logs the error and message, returns either the provided message or a git equivalent if possible.
// Always logs the full message with error as warning.
// Note: git errors should be processed in command package, this will probably be removed in the future.
func processGitErrorf(err error, format string, args ...interface{}) error {
	// create fallback error returned if we can't map it
	fallbackErr := errors.Internal(err, format, args...)

	// always log internal error together with message.
	log.Warn().Msgf("%v: [GIT] %v", fallbackErr, err)

	switch {
	case err.Error() == "no such file or directory":
		return errors.NotFound("repository not found")
	case strings.Contains(err.Error(), "reference already exists"):
		return errors.Conflict("reference already exists")
	default:
		return fallbackErr
	}
}

// MergeUnrelatedHistoriesError represents an error if merging fails due to unrelated histories.
type MergeUnrelatedHistoriesError struct {
	Method enum.MergeMethod
	StdOut string
	StdErr string
	Err    error
}

func IsMergeUnrelatedHistoriesError(err error) bool {
	return errors.Is(err, &MergeUnrelatedHistoriesError{})
}

func (e *MergeUnrelatedHistoriesError) Error() string {
	return fmt.Sprintf("Merge UnrelatedHistories Error: %v: %s\n%s", e.Err, e.StdErr, e.StdOut)
}

func (e *MergeUnrelatedHistoriesError) Unwrap() error {
	return e.Err
}

//nolint:errorlint // the purpose of this method is to check whether the target itself if of this type.
func (e *MergeUnrelatedHistoriesError) Is(target error) bool {
	_, ok := target.(*MergeUnrelatedHistoriesError)
	return ok
}
