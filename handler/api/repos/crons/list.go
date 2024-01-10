// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Drone Non-Commercial License
// that can be found in the LICENSE file.

//go:build !oss
// +build !oss

package crons

import (
	"net/http"

	"github.com/drone/drone/core"
	"github.com/drone/drone/handler/api/render"

	"github.com/go-chi/chi/v5"
)

// HandleList returns an http.HandlerFunc that writes a json-encoded
// list of cron jobs to the response body.
func HandleList(
	repos core.RepositoryStore,
	crons core.CronStore,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			namespace = chi.URLParam(r, "owner")
			name      = chi.URLParam(r, "name")
		)
		repo, err := repos.FindName(r.Context(), namespace, name)
		if err != nil {
			render.NotFound(w, err)
			return
		}
		list, err := crons.List(r.Context(), repo.ID)
		if err != nil {
			render.NotFound(w, err)
			return
		}
		render.JSON(w, list, 200)
	}
}
