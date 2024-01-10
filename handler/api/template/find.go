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

// HandleFind returns an http.HandlerFunc that writes json-encoded
// template details to the response body.
func HandleFind(templateStore core.TemplateStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var (
			name      = chi.URLParam(r, "name")
			namespace = chi.URLParam(r, "namespace")
		)
		template, err := templateStore.FindName(r.Context(), name, namespace)
		if err != nil {
			render.NotFound(w, err)
			return
		}
		render.JSON(w, template, 200)
	}
}
