// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Drone Non-Commercial License
// that can be found in the LICENSE file.

//go:build !oss
// +build !oss

package template

import (
	"encoding/json"
	"net/http"

	"github.com/drone/drone/core"
	"github.com/drone/drone/handler/api/render"

	"github.com/go-chi/chi/v5"
)

type templateUpdate struct {
	Data      *string `json:"data"`
	Namespace *string `json:"namespace"`
}

// HandleUpdate returns an http.HandlerFunc that processes http
// requests to update a template.
func HandleUpdate(templateStore core.TemplateStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			name      = chi.URLParam(r, "name")
			namespace = chi.URLParam(r, "namespace")
		)

		in := new(templateUpdate)
		err := json.NewDecoder(r.Body).Decode(in)
		if err != nil {
			render.BadRequest(w, err)
			return
		}

		s, err := templateStore.FindName(r.Context(), name, namespace)
		if err != nil {
			render.NotFound(w, err)
			return
		}

		if in.Data != nil {
			s.Data = *in.Data
		}
		if in.Namespace != nil {
			s.Namespace = *in.Namespace
		}

		err = s.Validate()
		if err != nil {
			render.BadRequest(w, err)
			return
		}

		err = templateStore.Update(r.Context(), s)
		if err != nil {
			render.InternalError(w, err)
			return
		}

		render.JSON(w, s, 200)
	}
}
