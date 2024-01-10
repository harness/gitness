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

// HandleList returns an http.HandlerFunc that writes a json-encoded
// list of templates to the response body by namespace
func HandleList(templateStore core.TemplateStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		namespace := chi.URLParam(r, "namespace")
		list, err := templateStore.List(r.Context(), namespace)
		if err != nil {
			render.NotFound(w, err)
			return
		}
		render.JSON(w, list, 200)
	}
}
