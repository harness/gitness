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

package events

import (
	"errors"
	"fmt"
)

var (
	errDiscardEvent = &discardEventError{}
)

// discardEventError is an error which, if returned by the event handler,
// causes the source event to be discarded despite any errors.
type discardEventError struct {
	inner error
}

func NewDiscardEventError(inner error) error {
	return &discardEventError{
		inner: inner,
	}
}

func NewDiscardEventErrorf(format string, args ...interface{}) error {
	return &discardEventError{
		inner: fmt.Errorf(format, args...),
	}
}

func (e *discardEventError) Error() string {
	return fmt.Sprintf("discarding requested due to: %s", e.inner)
}

func (e *discardEventError) Unwrap() error {
	return e.inner
}

func (e *discardEventError) Is(target error) bool {
	// NOTE: it's an internal event and we only ever check with the singleton instance
	return errors.Is(target, errDiscardEvent)
}
