// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Drone Non-Commercial License
// that can be found in the LICENSE file.

//go:build !oss
// +build !oss

package template

import (
	"bytes"
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

func TestHandleUpdate(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	template := mock.NewMockTemplateStore(controller)
	template.EXPECT().FindName(gomock.Any(), dummyTemplate.Name, dummyTemplate.Namespace).Return(dummyTemplate, nil)
	template.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

	c := new(chi.Context)
	c.URLParams.Add("name", "my_template.yml")
	c.URLParams.Add("namespace", "my_org")

	in := new(bytes.Buffer)
	json.NewEncoder(in).Encode(dummyTemplate)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", in)
	r = r.WithContext(
		context.WithValue(context.Background(), chi.RouteCtxKey, c),
	)

	HandleUpdate(template).ServeHTTP(w, r)
	if got, want := w.Code, http.StatusOK; want != got {
		t.Errorf("Want response code %d, got %d", want, got)
	}
}

func TestHandleUpdate_ValidationErrorData(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	template := mock.NewMockTemplateStore(controller)
	template.EXPECT().FindName(gomock.Any(), dummyTemplate.Name, dummyTemplate.Namespace).Return(&core.Template{Name: "my_template.yml"}, nil)

	c := new(chi.Context)
	c.URLParams.Add("name", "my_template.yml")
	c.URLParams.Add("namespace", "my_org")

	in := new(bytes.Buffer)
	json.NewEncoder(in).Encode(&core.Secret{Data: ""})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", in)
	r = r.WithContext(
		context.WithValue(context.Background(), chi.RouteCtxKey, c),
	)

	HandleUpdate(template).ServeHTTP(w, r)
	if got, want := w.Code, http.StatusBadRequest; want != got {
		t.Errorf("Want response code %d, got %d", want, got)
	}

	got, want := new(errors.Error), &errors.Error{Message: "No Template Data Provided"}
	json.NewDecoder(w.Body).Decode(got)
	if diff := cmp.Diff(got, want); len(diff) != 0 {
		t.Errorf(diff)
	}
}

func TestHandleUpdate_TemplateNotFound(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	template := mock.NewMockTemplateStore(controller)
	template.EXPECT().FindName(gomock.Any(), dummyTemplate.Name, dummyTemplate.Namespace).Return(nil, errors.ErrNotFound)

	c := new(chi.Context)
	c.URLParams.Add("name", "my_template.yml")
	c.URLParams.Add("namespace", "my_org")

	in := new(bytes.Buffer)
	json.NewEncoder(in).Encode(&core.Secret{})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", in)
	r = r.WithContext(
		context.WithValue(context.Background(), chi.RouteCtxKey, c),
	)

	HandleUpdate(template).ServeHTTP(w, r)
	if got, want := w.Code, http.StatusNotFound; want != got {
		t.Errorf("Want response code %d, got %d", want, got)
	}

	got, want := new(errors.Error), errors.ErrNotFound
	json.NewDecoder(w.Body).Decode(got)
	if diff := cmp.Diff(got, want); len(diff) != 0 {
		t.Errorf(diff)
	}
}

func TestHandleUpdate_UpdateError(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	template := mock.NewMockTemplateStore(controller)
	template.EXPECT().FindName(gomock.Any(), dummyTemplate.Name, dummyTemplate.Namespace).Return(&core.Template{Name: "my_template.yml"}, nil)
	template.EXPECT().Update(gomock.Any(), gomock.Any()).Return(errors.ErrNotFound)

	c := new(chi.Context)
	c.URLParams.Add("name", "my_template.yml")
	c.URLParams.Add("namespace", "my_org")

	in := new(bytes.Buffer)
	json.NewEncoder(in).Encode(&core.Template{Data: "my_data"})

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", in)
	r = r.WithContext(
		context.WithValue(context.Background(), chi.RouteCtxKey, c),
	)

	HandleUpdate(template).ServeHTTP(w, r)
	if got, want := w.Code, http.StatusInternalServerError; want != got {
		t.Errorf("Want response code %d, got %d", want, got)
	}

	got, want := new(errors.Error), errors.ErrNotFound
	json.NewDecoder(w.Body).Decode(got)
	if diff := cmp.Diff(got, want); len(diff) != 0 {
		t.Errorf(diff)
	}
}
