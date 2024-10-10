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

package check

import (
	"github.com/harness/gitness/types/enum"
)

var (
	ErrServiceAccountParentTypeIsInvalid = &ValidationError{
		"Provided parent type is invalid.",
	}
	ErrServiceAccountParentIDInvalid = &ValidationError{
		"ParentID required - Global service accounts are not supported.",
	}
)

// ServiceAccountParent verifies the remaining fields of a service account
// that aren't inherited from principal.
func ServiceAccountParent(parentType enum.ParentResourceType, parentID int64) error {
	if parentType != enum.ParentResourceTypeRepo && parentType != enum.ParentResourceTypeSpace {
		return ErrServiceAccountParentTypeIsInvalid
	}

	// validate service account belongs to sth
	if parentID <= 0 {
		return ErrServiceAccountParentIDInvalid
	}

	return nil
}
