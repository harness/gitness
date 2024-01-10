// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Drone Non-Commercial License
// that can be found in the LICENSE file.

//go:build !oss
// +build !oss

package card

import (
	"strconv"

	"github.com/drone/drone/core"
	"github.com/drone/drone/handler/api/render"

	"net/http"

	"github.com/go-chi/chi/v5"
)

// HandleDelete returns an http.HandlerFunc that processes http
// requests to delete a card.
func HandleDelete(
	buildStore core.BuildStore,
	cardStore core.CardStore,
	stageStore core.StageStore,
	stepStore core.StepStore,
	repoStore core.RepositoryStore,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			namespace = chi.URLParam(r, "owner")
			name      = chi.URLParam(r, "name")
		)

		buildNumber, err := strconv.ParseInt(chi.URLParam(r, "build"), 10, 64)
		if err != nil {
			render.BadRequest(w, err)
			return
		}

		stageNumber, err := strconv.Atoi(chi.URLParam(r, "stage"))
		if err != nil {
			render.BadRequest(w, err)
			return
		}

		stepNumber, err := strconv.Atoi(chi.URLParam(r, "step"))
		if err != nil {
			render.BadRequest(w, err)
			return
		}

		repo, err := repoStore.FindName(r.Context(), namespace, name)
		if err != nil {
			render.NotFound(w, err)
			return
		}
		build, err := buildStore.FindNumber(r.Context(), repo.ID, buildNumber)
		if err != nil {
			render.NotFound(w, err)
			return
		}
		stage, err := stageStore.FindNumber(r.Context(), build.ID, stageNumber)
		if err != nil {
			render.NotFound(w, err)
			return
		}
		step, err := stepStore.FindNumber(r.Context(), stage.ID, stepNumber)
		if err != nil {
			render.NotFound(w, err)
			return
		}

		_, err = cardStore.Find(r.Context(), step.ID)
		if err != nil {
			render.NotFound(w, err)
			return
		}
		err = cardStore.Delete(r.Context(), step.ID)
		if err != nil {
			render.InternalError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
