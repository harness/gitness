// Copyright 2019 Drone IO, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package acl

import (
	"net/http"
	"time"

	"github.com/drone/drone/core"
	"github.com/drone/drone/handler/api/errors"
	"github.com/drone/drone/handler/api/render"
	"github.com/drone/drone/handler/api/request"
	"github.com/drone/drone/logger"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

// InjectRepository returns an http.Handler middleware that injects
// the repository and repository permissions into the context.
func InjectRepository(
	repoz core.RepositoryService,
	repos core.RepositoryStore,
	perms core.PermStore,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var (
				ctx   = r.Context()
				owner = chi.URLParam(r, "owner")
				name  = chi.URLParam(r, "name")
			)

			log := logger.FromRequest(r).WithFields(
				logrus.Fields{
					"namespace": owner,
					"name":      name,
				},
			)

			// the user is stored in the context and is
			// provided by a an ancestor middleware in the
			// chain.
			user, sessionExists := request.UserFrom(ctx)

			repo, err := repos.FindName(ctx, owner, name)
			if err != nil {
				if sessionExists {
					render.NotFound(w, errors.ErrNotFound)
				} else {
					render.Unauthorized(w, errors.ErrUnauthorized)
				}
				log.WithError(err).Debugln("api: repository not found")
				return
			}

			// the repository is stored in the request context
			// and can be accessed by subsequent handlers in the
			// request chain.
			ctx = request.WithRepo(ctx, repo)

			// if the user does not exist in the request context,
			// this is a guest session, and there are no repository
			// permissions to lookup.
			if !sessionExists {
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			// else get the cached permissions from the database
			// for the user and repository.
			perm, err := perms.Find(ctx, repo.UID, user.ID)
			if err != nil {
				// if the permissions are not found we forward
				// the request to the next handler in the chain
				// with no permissions in the context.
				//
				// It is the responsibility to downstream
				// middleware and handlers to decide if the
				// request should be rejected.
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			log = log.WithFields(
				logrus.Fields{
					"read":  perm.Read,
					"write": perm.Write,
					"admin": perm.Admin,
				},
			)

			// because the permissions are synced with the remote
			// system (e.g. github) they may be stale. If the permissions
			// are stale they are refreshed below.
			if perm.Synced == 0 || time.Unix(perm.Synced, 0).Add(time.Hour).Before(time.Now()) {
				log.Debugln("api: sync repository permissions")

				permv, err := repoz.FindPerm(ctx, user, repo.Slug)
				if err != nil {
					render.NotFound(w, errors.ErrNotFound)
					log.WithError(err).
						Warnln("api: cannot sync repository permissions")
					return
				}

				perm.Synced = time.Now().Unix()
				perm.Read = permv.Read
				perm.Write = permv.Write
				perm.Admin = permv.Admin

				err = perms.Update(ctx, perm)
				if err != nil {
					log.WithError(err).Debugln("api: cannot cache repository permissions")
				} else {
					log.Debugln("api: repository permissions synchronized")
				}
			}

			ctx = request.WithPerm(ctx, perm)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
