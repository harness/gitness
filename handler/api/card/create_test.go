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

type card struct {
	Id   int64  `json:"id,omitempty"`
	Data []byte `json:"card_data"`
}

var (
	dummyRepo = &core.Repository{
		ID:     1,
		UserID: 1,
		Slug:   "octocat/hello-world",
	}
	dummyBuild = &core.Build{
		ID:     1,
		RepoID: 1,
		Number: 1,
	}
	dummyStage = &core.Stage{
		ID:      1,
		BuildID: 1,
	}
	dummyStep = &core.Step{
		ID:      1,
		StageID: 1,
		Schema:  "https://myschema.com",
	}
	dummyCard = &card{
		Id:   dummyStep.ID,
		Data: []byte("{\"type\": \"AdaptiveCard\"}"),
	}
)

func TestHandleCreate(t *testing.T) {
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
	step.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

	card := mock.NewMockCardStore(controller)
	card.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	c := new(chi.Context)
	c.URLParams.Add("owner", "octocat")
	c.URLParams.Add("name", "hello-world")
	c.URLParams.Add("build", "1")
	c.URLParams.Add("stage", "1")
	c.URLParams.Add("step", "1")

	in := new(bytes.Buffer)
	json.NewEncoder(in).Encode(dummyCard)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", in)
	r = r.WithContext(
		context.WithValue(context.Background(), chi.RouteCtxKey, c),
	)

	HandleCreate(build, card, stage, step, repos).ServeHTTP(w, r)
	if got, want := w.Code, http.StatusOK; want != got {
		t.Errorf("Want response code %d, got %d", want, got)
	}
}

func TestHandleCreate_BadRequest(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	c := new(chi.Context)
	c.URLParams.Add("owner", "octocat")
	c.URLParams.Add("name", "hello-world")
	c.URLParams.Add("build", "1")
	c.URLParams.Add("stage", "1")
	c.URLParams.Add("step", "1")

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", nil)
	r = r.WithContext(
		context.WithValue(context.Background(), chi.RouteCtxKey, c),
	)

	HandleCreate(nil, nil, nil, nil, nil).ServeHTTP(w, r)
	got, want := &errors.Error{}, &errors.Error{Message: "EOF"}
	json.NewDecoder(w.Body).Decode(got)
	if diff := cmp.Diff(got, want); len(diff) != 0 {
		t.Errorf(diff)
	}
}

func TestHandleCreate_CreateError(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	repos := mock.NewMockRepositoryStore(controller)
	repos.EXPECT().FindName(gomock.Any(), "octocat", "hello-world").Return(dummyRepo, nil)

	build := mock.NewMockBuildStore(controller)
	build.EXPECT().FindNumber(gomock.Any(), dummyBuild.ID, gomock.Any()).Return(dummyBuild, nil)

	stage := mock.NewMockStageStore(controller)
	stage.EXPECT().FindNumber(gomock.Any(), dummyBuild.ID, gomock.Any()).Return(dummyStage, nil)

	card := mock.NewMockCardStore(controller)
	card.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.ErrNotFound)

	step := mock.NewMockStepStore(controller)
	step.EXPECT().FindNumber(gomock.Any(), dummyStage.ID, gomock.Any()).Return(dummyStep, nil)

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

	HandleCreate(build, card, stage, step, repos).ServeHTTP(w, r)
	if got, want := w.Code, http.StatusInternalServerError; want != got {
		t.Errorf("Want response code %d, got %d", want, got)
	}

	got, want := new(errors.Error), errors.ErrNotFound
	json.NewDecoder(w.Body).Decode(got)
	if diff := cmp.Diff(got, want); len(diff) != 0 {
		t.Errorf(diff)
	}
}
