// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Drone Non-Commercial License
// that can be found in the LICENSE file.

package users

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/drone/drone/core"
	"github.com/drone/drone/handler/api/errors"
	"github.com/drone/drone/mock"
	"github.com/go-chi/chi/v5"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// The purpose of this test is to make sure admins can rotate someone
// else's token.
func TestTokenRotate(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()
	startingHash := "MjAxOC0wOC0xMVQxNTo1ODowN1o"
	mockUser := &core.User{
		ID:    1,
		Login: "octocat",
		Hash:  startingHash,
	}

	c := new(chi.Context)
	c.URLParams.Add("user", "octocat")

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/", nil)
	r = r.WithContext(
		context.WithValue(context.Background(), chi.RouteCtxKey, c),
	)

	users := mock.NewMockUserStore(controller)
	users.EXPECT().FindLogin(gomock.Any(), mockUser.Login).Return(mockUser, nil)
	users.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

	HandleTokenRotation(users)(w, r)
	if got, want := w.Code, 200; want != got {
		t.Errorf("Want response code %d, got %d", want, got)
	}

	got, want := &userWithMessage{}, mockUser
	json.NewDecoder(w.Body).Decode(got)

	ignore := cmpopts.IgnoreFields(core.User{}, "Hash")
	if diff := cmp.Diff(got.User, want, ignore); len(diff) != 0 {
		t.Errorf(diff)
	}
	if got.Message == "" {
		t.Errorf("Expect Message returned")
	}
	if got, want := mockUser.Hash, startingHash; got == want {
		t.Errorf("Expect user hash updated")
	}
}

// the purpose of this unit test is to verify we fail safely when a non existing user is provided
func TestToken_UserNotFound(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	startingHash := "MjAxOC0wOC0xMVQxNTo1ODowN1o"
	mockUser := &core.User{
		ID:    1,
		Login: "octocat",
		Hash:  startingHash,
	}
	c := new(chi.Context)
	c.URLParams.Add("user", "octocat")

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/?rotate=true", nil)
	r = r.WithContext(
		context.WithValue(context.Background(), chi.RouteCtxKey, c),
	)

	users := mock.NewMockUserStore(controller)
	users.EXPECT().FindLogin(gomock.Any(), mockUser.Login).Return(mockUser, nil)
	users.EXPECT().Update(gomock.Any(), gomock.Any()).Return(errors.ErrNotFound)

	HandleTokenRotation(users)(w, r)
	if got, want := w.Code, 500; want != got {
		t.Errorf("Want response code %d, got %d", want, got)
	}

	got, want := new(errors.Error), errors.ErrNotFound
	json.NewDecoder(w.Body).Decode(got)
	if diff := cmp.Diff(got, want); len(diff) != 0 {
		t.Errorf(diff)
	}
}

// the purpose of this unit test is to verify we fail safely when a non existing user is provided
func TestToken_UpdateError(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	c := new(chi.Context)
	c.URLParams.Add("user", "octocat")

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/?rotate=true", nil)
	r = r.WithContext(
		context.WithValue(context.Background(), chi.RouteCtxKey, c),
	)

	users := mock.NewMockUserStore(controller)
	users.EXPECT().FindLogin(gomock.Any(), mockUser.Login).Return(nil, errors.ErrNotFound)

	HandleTokenRotation(users)(w, r)
	if got, want := w.Code, 404; want != got {
		t.Errorf("Want response code %d, got %d", want, got)
	}

	got, want := new(errors.Error), errors.ErrNotFound
	json.NewDecoder(w.Body).Decode(got)
	if diff := cmp.Diff(got, want); len(diff) != 0 {
		t.Errorf(diff)
	}
}
