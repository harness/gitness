// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Drone Non-Commercial License
// that can be found in the LICENSE file.

//go:build !oss
// +build !oss

package card

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/drone/drone/handler/api/errors"
	"github.com/drone/drone/mock"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
)

func TestHandleFind(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	repos := mock.NewMockRepositoryStore(controller)
	repos.EXPECT().FindName(gomock.Any(), "octocat", "hello-world").Return(dummyRepo, nil)

	build := mock.NewMockBuildStore(controller)
	build.EXPECT().FindNumber(gomock.Any(), dummyBuild.ID, gomock.Any()).Return(dummyBuild, nil)

	stage := mock.NewMockStageStore(controller)
	stage.EXPECT().FindNumber(gomock.Any(), dummyBuild.ID, gomock.Any()).Return(dummyStage, nil)

	step := mock.NewMockStepStore(controller)
	step.EXPECT().FindNumber(gomock.Any(), dummyStage.ID, gomock.Any()).Return(dummyStep, nil)

	card := mock.NewMockCardStore(controller)
	card.EXPECT().Find(gomock.Any(), dummyStep.ID).Return(ioutil.NopCloser(
		bytes.NewBuffer(dummyCard.Data),
	), nil)

	c := new(chi.Context)
	c.URLParams.Add("owner", "octocat")
	c.URLParams.Add("name", "hello-world")
	c.URLParams.Add("build", "1")
	c.URLParams.Add("stage", "1")
	c.URLParams.Add("step", "1")

	in := new(bytes.Buffer)
	json.NewEncoder(in).Encode(dummyCard)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", in)
	r = r.WithContext(
		context.WithValue(context.Background(), chi.RouteCtxKey, c),
	)

	HandleFind(build, card, stage, step, repos).ServeHTTP(w, r)
	if got, want := w.Code, http.StatusOK; want != got {
		t.Errorf("Want response code %d, got %d", want, got)
	}
}

func TestHandleFind_CardNotFound(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	repos := mock.NewMockRepositoryStore(controller)
	repos.EXPECT().FindName(gomock.Any(), "octocat", "hello-world").Return(dummyRepo, nil)

	build := mock.NewMockBuildStore(controller)
	build.EXPECT().FindNumber(gomock.Any(), dummyBuild.ID, gomock.Any()).Return(dummyBuild, nil)

	stage := mock.NewMockStageStore(controller)
	stage.EXPECT().FindNumber(gomock.Any(), dummyBuild.ID, gomock.Any()).Return(dummyStage, nil)

	step := mock.NewMockStepStore(controller)
	step.EXPECT().FindNumber(gomock.Any(), dummyStage.ID, gomock.Any()).Return(dummyStep, nil)

	card := mock.NewMockCardStore(controller)
	card.EXPECT().Find(gomock.Any(), dummyStep.ID).Return(nil, errors.ErrNotFound)

	c := new(chi.Context)
	c.URLParams.Add("owner", "octocat")
	c.URLParams.Add("name", "hello-world")
	c.URLParams.Add("build", "1")
	c.URLParams.Add("stage", "1")
	c.URLParams.Add("step", "1")

	in := new(bytes.Buffer)
	json.NewEncoder(in).Encode(dummyCard)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/", in)
	r = r.WithContext(
		context.WithValue(context.Background(), chi.RouteCtxKey, c),
	)

	HandleFind(build, card, stage, step, repos).ServeHTTP(w, r)
	if got, want := w.Code, http.StatusNotFound; want != got {
		t.Errorf("Want response code %d, got %d", want, got)
	}

	got, want := new(errors.Error), errors.ErrNotFound
	json.NewDecoder(w.Body).Decode(got)
	if diff := cmp.Diff(got, want); len(diff) != 0 {
		t.Errorf(diff)
	}
}
