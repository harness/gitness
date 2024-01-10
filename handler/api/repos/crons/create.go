// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Drone Non-Commercial License
// that can be found in the LICENSE file.

//go:build !oss
// +build !oss

package crons

import (
	"encoding/json"
	"net/http"

	"github.com/drone/drone/core"
	"github.com/drone/drone/handler/api/render"

	"github.com/go-chi/chi/v5"
)

// HandleCreate returns an http.HandlerFunc that processes http
// requests to create a new cronjob.
func HandleCreate(
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
		in := new(core.Cron)
		err = json.NewDecoder(r.Body).Decode(in)
		if err != nil {
			render.BadRequest(w, err)
			return
		}
		cronjob := new(core.Cron)
		cronjob.Event = core.EventPush
		cronjob.Branch = in.Branch
		cronjob.RepoID = repo.ID
		cronjob.SetName(in.Name)
		err = cronjob.SetExpr(in.Expr)
		if err != nil {
			render.BadRequest(w, err)
			return
		}

		err = cronjob.Validate()
		if err != nil {
			render.BadRequest(w, err)
			return
		}

		err = crons.Create(r.Context(), cronjob)
		if err != nil {
			render.InternalError(w, err)
			return
		}
		render.JSON(w, cronjob, 200)
	}
}
