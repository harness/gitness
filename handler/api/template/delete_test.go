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

	"github.com/drone/drone/handler/api/errors"
	"github.com/drone/drone/mock"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
)

func TestHandleDelete(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	template := mock.NewMockTemplateStore(controller)
	template.EXPECT().FindName(gomock.Any(), dummyTemplate.Name, dummyTemplate.Namespace).Return(dummyTemplate, nil)
	template.EXPECT().Delete(gomock.Any(), dummyTemplate).Return(nil)

	c := new(chi.Context)
	c.URLParams.Add("name", "my_template.yml")
	c.URLParams.Add("namespace", "my_org")

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	r = r.WithContext(
		context.WithValue(context.Background(), chi.RouteCtxKey, c),
	)

	HandleDelete(template).ServeHTTP(w, r)
	if got, want := w.Code, http.StatusNoContent; want != got {
		t.Errorf("Want response code %d, got %d", want, got)
	}
}

func TestHandleDelete_TemplateNotFound(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	template := mock.NewMockTemplateStore(controller)
	template.EXPECT().FindName(gomock.Any(), dummyTemplate.Name, dummyTemplate.Namespace).Return(nil, errors.ErrNotFound)

	c := new(chi.Context)
	c.URLParams.Add("name", "my_template.yml")
	c.URLParams.Add("namespace", "my_org")

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	r = r.WithContext(
		context.WithValue(context.Background(), chi.RouteCtxKey, c),
	)

	HandleDelete(template).ServeHTTP(w, r)
	if got, want := w.Code, http.StatusNotFound; want != got {
		t.Errorf("Want response code %d, got %d", want, got)
	}

	got, want := new(errors.Error), errors.ErrNotFound
	json.NewDecoder(w.Body).Decode(got)
	if diff := cmp.Diff(got, want); len(diff) != 0 {
		t.Errorf(diff)
	}
}

func TestHandleDelete_DeleteError(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	template := mock.NewMockTemplateStore(controller)
	template.EXPECT().FindName(gomock.Any(), dummyTemplate.Name, dummyTemplate.Namespace).Return(dummyTemplate, nil)
	template.EXPECT().Delete(gomock.Any(), dummyTemplate).Return(errors.ErrNotFound)

	c := new(chi.Context)
	c.URLParams.Add("name", "my_template.yml")
	c.URLParams.Add("namespace", "my_org")

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", nil)
	r = r.WithContext(
		context.WithValue(context.Background(), chi.RouteCtxKey, c),
	)

	HandleDelete(template).ServeHTTP(w, r)
	if got, want := w.Code, http.StatusInternalServerError; want != got {
		t.Errorf("Want response code %d, got %d", want, got)
	}

	got, want := new(errors.Error), errors.ErrNotFound
	json.NewDecoder(w.Body).Decode(got)
	if diff := cmp.Diff(got, want); len(diff) != 0 {
		t.Errorf(diff)
	}
}
