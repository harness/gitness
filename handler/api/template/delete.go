// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Drone Non-Commercial License
// that can be found in the LICENSE file.

//go:build !oss
// +build !oss

package template

import (
	"net/http"

	"github.com/drone/drone/core"
	"github.com/drone/drone/handler/api/render"

	"github.com/go-chi/chi/v5"
)

// HandleDelete returns an http.HandlerFunc that processes http
// requests to delete a template.
func HandleDelete(template core.TemplateStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			name      = chi.URLParam(r, "name")
			namespace = chi.URLParam(r, "namespace")
		)
		s, err := template.FindName(r.Context(), name, namespace)
		if err != nil {
			render.NotFound(w, err)
			return
		}
		err = template.Delete(r.Context(), s)
		if err != nil {
			render.InternalError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
