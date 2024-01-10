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

package users

import (
	"net/http"

	"github.com/dchest/uniuri"
	"github.com/drone/drone/core"
	"github.com/drone/drone/handler/api/render"
	"github.com/drone/drone/logger"
	"github.com/go-chi/chi/v5"
)

type userWithMessage struct {
	*core.User
	Message string `json:"message"`
}

// HandleToken returns an http.HandlerFunc that writes json-encoded
// account information to the http response body with the user token.
func HandleTokenRotation(users core.UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		login := chi.URLParam(r, "user")
		user, err := users.FindLogin(r.Context(), login)
		if err != nil {
			render.NotFound(w, err)
			logger.FromRequest(r).WithError(err).
				Debugln("api: cannot find user")
			return
		}
		user.Hash = uniuri.NewLen(32)
		if err := users.Update(r.Context(), user); err != nil {
			render.InternalError(w, err)
			return
		}
		render.JSON(w, &userWithMessage{user, "Token rotated successfully."}, 200)
	}
}
