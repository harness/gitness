// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Drone Non-Commercial License
// that can be found in the LICENSE file.

//go:build !oss
// +build !oss

package template

import (
	"encoding/json"
	"net/http"
	"path/filepath"

	"github.com/drone/drone/core"
	"github.com/drone/drone/handler/api/errors"
	"github.com/drone/drone/handler/api/render"

	"github.com/go-chi/chi/v5"
)

var (
	errTemplateExtensionInvalid = errors.New("Template extension invalid. Must be yaml, starlark or jsonnet")
)

type templateInput struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

// HandleCreate returns an http.HandlerFunc that processes http
// requests to create a new template.
func HandleCreate(templateStore core.TemplateStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		namespace := chi.URLParam(r, "namespace")
		in := new(templateInput)
		err := json.NewDecoder(r.Body).Decode(in)
		if err != nil {
			render.BadRequest(w, err)
			return
		}

		// check valid template extension type
		switch filepath.Ext(in.Name) {
		case ".yml", ".yaml":
		case ".star", ".starlark", ".script":
		case ".jsonnet":
		default:
			render.BadRequest(w, errTemplateExtensionInvalid)
			return
		}

		t := &core.Template{
			Name:      in.Name,
			Data:      in.Data,
			Namespace: namespace,
		}

		err = t.Validate()
		if err != nil {
			render.BadRequest(w, err)
			return
		}

		err = templateStore.Create(r.Context(), t)
		if err != nil {
			render.InternalError(w, err)
			return
		}

		render.JSON(w, t, 200)
	}
}
