// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Drone Non-Commercial License
// that can be found in the LICENSE file.

//go:build !oss
// +build !oss

package template

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/drone/drone/core"
	"github.com/drone/drone/handler/api/errors"
	"github.com/drone/drone/mock"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
)

var (
	dummyTemplate = &core.Template{
		Name:      "my_template.yml",
		Data:      "my_data",
		Created:   1,
		Updated:   2,
		Namespace: "my_org",
	}
	dummyTemplateList = []*core.Template{
		dummyTemplate,
	}
)

func TestHandleAll(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	templates := mock.NewMockTemplateStore(controller)
	templates.EXPECT().ListAll(gomock.Any()).Return(dummyTemplateList, nil)

	c := new(chi.Context)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	r = r.WithContext(
		context.WithValue(context.Background(), chi.RouteCtxKey, c),
	)

	HandleListAll(templates).ServeHTTP(w, r)
	if got, want := w.Code, http.StatusOK; want != got {
		t.Errorf("Want response code %d, got %d", want, got)
	}
}

func TestHandleAll_TemplateListErr(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	templates := mock.NewMockTemplateStore(controller)
	templates.EXPECT().ListAll(gomock.Any()).Return(nil, errors.ErrNotFound)

	c := new(chi.Context)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	r = r.WithContext(
		context.WithValue(context.Background(), chi.RouteCtxKey, c),
	)

	HandleListAll(templates).ServeHTTP(w, r)
	if got, want := w.Code, http.StatusNotFound; want != got {
		t.Errorf("Want response code %d, got %d", want, got)
	}

	got, want := new(errors.Error), errors.ErrNotFound
	json.NewDecoder(w.Body).Decode(got)
	if diff := cmp.Diff(got, want); len(diff) != 0 {
		t.Errorf(diff)
	}
}
