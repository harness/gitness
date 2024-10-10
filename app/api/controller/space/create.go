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

package space

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	apiauth "github.com/harness/gitness/app/api/auth"
	"github.com/harness/gitness/app/api/usererror"
	"github.com/harness/gitness/app/auth"
	"github.com/harness/gitness/app/bootstrap"
	"github.com/harness/gitness/app/paths"
	"github.com/harness/gitness/types"
	"github.com/harness/gitness/types/check"
	"github.com/harness/gitness/types/enum"
)

var (
	errParentIDNegative = usererror.BadRequest(
		"Parent ID has to be either zero for a root space or greater than zero for a child space.")
)

type CreateInput struct {
	ParentRef string `json:"parent_ref"`
	// TODO [CODE-1363]: remove after identifier migration.
	UID         string `json:"uid" deprecated:"true"`
	Identifier  string `json:"identifier"`
	Description string `json:"description"`
	IsPublic    bool   `json:"is_public"`
}

// Create creates a new space.
//
//nolint:gocognit // refactor if required
func (c *Controller) Create(
	ctx context.Context,
	session *auth.Session,
	in *CreateInput,
) (*SpaceOutput, error) {
	if err := c.sanitizeCreateInput(in); err != nil {
		return nil, fmt.Errorf("failed to sanitize input: %w", err)
	}

	parentSpace, err := c.getSpaceCheckAuthSpaceCreation(ctx, session, in.ParentRef)
	if err != nil {
		return nil, err
	}

	isPublicAccessSupported, err := c.publicAccess.IsPublicAccessSupported(ctx, parentSpace.Path)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to check if public access is supported for parent space %q: %w",
			parentSpace.Path,
			err,
		)
	}
	if in.IsPublic && !isPublicAccessSupported {
		return nil, errPublicSpaceCreationDisabled
	}

	var space *types.Space
	err = c.tx.WithTx(ctx, func(ctx context.Context) error {
		space, err = c.createSpaceInnerInTX(ctx, session, parentSpace.ID, in)
		return err
	})
	if err != nil {
		return nil, err
	}

	err = c.publicAccess.Set(ctx, enum.PublicResourceTypeSpace, space.Path, in.IsPublic)
	if err != nil {
		if dErr := c.publicAccess.Delete(ctx, enum.PublicResourceTypeSpace, space.Path); dErr != nil {
			return nil, fmt.Errorf("failed to set space public access (and public access cleanup: %w): %w", dErr, err)
		}

		// only cleanup space itself if cleanup of public access succeeded or we risk leaking public access
		if dErr := c.PurgeNoAuth(ctx, session, space); dErr != nil {
			return nil, fmt.Errorf("failed to set space public access (and space purge: %w): %w", dErr, err)
		}

		return nil, fmt.Errorf("failed to set space public access (successful cleanup): %w", err)
	}

	return GetSpaceOutput(ctx, c.publicAccess, space)
}

func (c *Controller) createSpaceInnerInTX(
	ctx context.Context,
	session *auth.Session,
	parentID int64,
	in *CreateInput,
) (*types.Space, error) {
	spacePath := in.Identifier
	if parentID > 0 {
		// (re-)read parent path in transaction to ensure correctness
		parentPath, err := c.spacePathStore.FindPrimaryBySpaceID(ctx, parentID)
		if err != nil {
			return nil, fmt.Errorf("failed to find primary path for parent '%d': %w", parentID, err)
		}
		spacePath = paths.Concatenate(parentPath.Value, in.Identifier)

		// ensure path is within accepted depth!
		err = check.PathDepth(spacePath, true)
		if err != nil {
			return nil, fmt.Errorf("path is invalid: %w", err)
		}
	}

	now := time.Now().UnixMilli()
	space := &types.Space{
		Version:     0,
		ParentID:    parentID,
		Identifier:  in.Identifier,
		Description: in.Description,
		Path:        spacePath,
		CreatedBy:   session.Principal.ID,
		Created:     now,
		Updated:     now,
	}
	err := c.spaceStore.Create(ctx, space)
	if err != nil {
		return nil, fmt.Errorf("space creation failed: %w", err)
	}

	pathSegment := &types.SpacePathSegment{
		Identifier: space.Identifier,
		IsPrimary:  true,
		SpaceID:    space.ID,
		ParentID:   parentID,
		CreatedBy:  space.CreatedBy,
		Created:    now,
		Updated:    now,
	}
	err = c.spacePathStore.InsertSegment(ctx, pathSegment)
	if err != nil {
		return nil, fmt.Errorf("failed to insert primary path segment: %w", err)
	}

	// add space membership to top level space only (as the user doesn't have inherited permissions already)
	if parentID == 0 {
		membership := &types.Membership{
			MembershipKey: types.MembershipKey{
				SpaceID:     space.ID,
				PrincipalID: session.Principal.ID,
			},
			Role: enum.MembershipRoleSpaceOwner,

			// membership has been created by the system
			CreatedBy: bootstrap.NewSystemServiceSession().Principal.ID,
			Created:   now,
			Updated:   now,
		}
		err = c.membershipStore.Create(ctx, membership)
		if err != nil {
			return nil, fmt.Errorf("failed to make user owner of the space: %w", err)
		}
	}

	return space, nil
}

func (c *Controller) getSpaceCheckAuthSpaceCreation(
	ctx context.Context,
	session *auth.Session,
	parentRef string,
) (*types.Space, error) {
	parentRefAsID, err := strconv.ParseInt(parentRef, 10, 64)
	if (parentRefAsID <= 0 && err == nil) || (len(strings.TrimSpace(parentRef)) == 0) {
		// TODO: Restrict top level space creation - should be move to authorizer?
		if auth.IsAnonymousSession(session) {
			return nil, fmt.Errorf("anonymous user not allowed to create top level spaces: %w", usererror.ErrUnauthorized)
		}

		return &types.Space{}, nil
	}

	parentSpace, err := c.spaceStore.FindByRef(ctx, parentRef)
	if err != nil {
		return nil, fmt.Errorf("failed to get parent space: %w", err)
	}

	if err = apiauth.CheckSpaceScope(
		ctx,
		c.authorizer,
		session,
		parentSpace,
		enum.ResourceTypeSpace,
		enum.PermissionSpaceEdit,
	); err != nil {
		return nil, fmt.Errorf("authorization failed: %w", err)
	}

	return parentSpace, nil
}

func (c *Controller) sanitizeCreateInput(in *CreateInput) error {
	// TODO [CODE-1363]: remove after identifier migration.
	if in.Identifier == "" {
		in.Identifier = in.UID
	}

	if len(in.ParentRef) > 0 && !c.nestedSpacesEnabled {
		// TODO (Nested Spaces): Remove once support is added
		return errNestedSpacesNotSupported
	}

	parentRefAsID, err := strconv.ParseInt(in.ParentRef, 10, 64)
	if err == nil && parentRefAsID < 0 {
		return errParentIDNegative
	}

	isRoot := false
	if (err == nil && parentRefAsID == 0) || (len(strings.TrimSpace(in.ParentRef)) == 0) {
		isRoot = true
	}

	if err := c.identifierCheck(in.Identifier, isRoot); err != nil {
		return err
	}

	in.Description = strings.TrimSpace(in.Description)
	if err := check.Description(in.Description); err != nil { //nolint:revive
		return err
	}

	return nil
}
