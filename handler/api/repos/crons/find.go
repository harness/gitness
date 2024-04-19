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

// HandleFind returns an http.HandlerFunc that writes json-encoded
// cronjob details to the response body.
func HandleFind(
	repos core.RepositoryStore,
	crons core.CronStore,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			namespace = chi.URLParam(r, "owner")
			name      = chi.URLParam(r, "name")
			cron      = chi.URLParam(r, "cron")
		)
		repo, err := repos.FindName(r.Context(), namespace, name)
		if err != nil {
			render.NotFound(w, err)
			return
		}
		cronjob, err := crons.FindName(r.Context(), repo.ID, cron)
		if err != nil {
			render.NotFound(w, err)
			return
		}
		render.JSON(w, cronjob, 200)
	}
}
