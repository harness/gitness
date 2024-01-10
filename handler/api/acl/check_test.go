// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Drone Non-Commercial License
// that can be found in the LICENSE file.

package acl

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/drone/drone/core"
	"github.com/drone/drone/handler/api/errors"
	"github.com/drone/drone/handler/api/request"
	"github.com/google/go-cmp/cmp"

	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
)

var noContext = context.Background()

// this test verifies that a 401 unauthorized error is written to
// the response if the client is not authenticated and repository
// visibility is internal or private.
func TestCheckAccess_Guest_Unauthorized(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/repos/octocat/hello-world", nil)
	r = r.WithContext(
		request.WithRepo(noContext, mockRepo),
	)

	router := chi.NewRouter()
	router.Route("/api/repos/{owner}/{name}", func(router chi.Router) {
		router.Use(CheckReadAccess())
		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			t.Errorf("Must not invoke next handler in middleware chain")
		})
	})

	router.ServeHTTP(w, r)

	if got, want := w.Code, http.StatusUnauthorized; got != want {
		t.Errorf("Want status code %d, got %d", want, got)
	}

	got, want := new(errors.Error), errors.ErrUnauthorized
	json.NewDecoder(w.Body).Decode(got)
	if diff := cmp.Diff(got, want); len(diff) != 0 {
		t.Errorf(diff)
	}
}

// this test verifies the the next handler in the middleware
// chain is processed if the user is not authenticated BUT
// the repository is publicly visible.
func TestCheckAccess_Guest_PublicVisibility(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	mockRepo := *mockRepo
	mockRepo.Visibility = core.VisibilityPublic

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/repos/octocat/hello-world", nil)
	r = r.WithContext(
		request.WithRepo(noContext, &mockRepo),
	)

	router := chi.NewRouter()
	router.Route("/api/repos/{owner}/{name}", func(router chi.Router) {
		router.Use(CheckReadAccess())
		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})
	})

	router.ServeHTTP(w, r)

	if got, want := w.Code, http.StatusTeapot; got != want {
		t.Errorf("Want status code %d, got %d", want, got)
	}
}

// this test verifies that a 401 unauthorized error is written to
// the response if the repository visibility is internal, and the
// client is not authenticated.
func TestCheckAccess_Guest_InternalVisibility(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	mockRepo := *mockRepo
	mockRepo.Visibility = core.VisibilityInternal

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/repos/octocat/hello-world", nil)
	r = r.WithContext(
		request.WithRepo(noContext, &mockRepo),
	)

	router := chi.NewRouter()
	router.Route("/api/repos/{owner}/{name}", func(router chi.Router) {
		router.Use(CheckReadAccess())
		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			t.Errorf("Must not invoke next handler in middleware chain")
		})
	})

	router.ServeHTTP(w, r)

	if got, want := w.Code, http.StatusUnauthorized; got != want {
		t.Errorf("Want status code %d, got %d", want, got)
	}
}

// this test verifies the the next handler in the middleware
// chain is processed if the user is authenticated AND
// the repository is publicly visible.
func TestCheckAccess_Authenticated_PublicVisibility(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	mockRepo := *mockRepo
	mockRepo.Visibility = core.VisibilityPublic

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/repos/octocat/hello-world", nil)
	r = r.WithContext(
		request.WithUser(
			request.WithRepo(noContext, &mockRepo), mockUser),
	)

	router := chi.NewRouter()
	router.Route("/api/repos/{owner}/{name}", func(router chi.Router) {
		router.Use(CheckReadAccess())
		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})
	})

	router.ServeHTTP(w, r)

	if got, want := w.Code, http.StatusTeapot; got != want {
		t.Errorf("Want status code %d, got %d", want, got)
	}
}

// this test verifies the the next handler in the middleware
// chain is processed if the user is authenticated AND
// the repository has internal visible.
func TestCheckAccess_Authenticated_InternalVisibility(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	mockRepo := *mockRepo
	mockRepo.Visibility = core.VisibilityInternal

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/repos/octocat/hello-world", nil)
	r = r.WithContext(
		request.WithUser(
			request.WithRepo(noContext, &mockRepo), mockUser),
	)

	router := chi.NewRouter()
	router.Route("/api/repos/{owner}/{name}", func(router chi.Router) {
		router.Use(CheckReadAccess())
		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})
	})

	router.ServeHTTP(w, r)

	if got, want := w.Code, http.StatusTeapot; got != want {
		t.Errorf("Want status code %d, got %d", want, got)
	}
}

// this test verifies that a 404 not found error is written to
// the response if the repository is not found AND the user is
// authenticated.
func TestCheckAccess_Authenticated_RepositoryNotFound(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/repos/octocat/hello-world", nil)

	router := chi.NewRouter()
	router.Route("/api/repos/{owner}/{name}", func(router chi.Router) {
		router.Use(CheckReadAccess())
		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			t.Errorf("Must not invoke next handler in middleware chain")
		})
	})

	router.ServeHTTP(w, r)

	if got, want := w.Code, http.StatusNotFound; got != want {
		t.Errorf("Want status code %d, got %d", want, got)
	}

	got, want := new(errors.Error), errors.ErrNotFound
	json.NewDecoder(w.Body).Decode(got)
	if diff := cmp.Diff(got, want); len(diff) != 0 {
		t.Errorf(diff)
	}
}

// this test verifies that a 404 not found error is written to
// the response if the user does not have permissions to access
// the repository.
func TestCheckAccess_Permission_NotFound(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/repos/octocat/hello-world", nil)
	r = r.WithContext(
		request.WithUser(
			request.WithRepo(noContext, mockRepo), mockUser),
	)

	router := chi.NewRouter()
	router.Route("/api/repos/{owner}/{name}", func(router chi.Router) {
		router.Use(CheckReadAccess())
		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			t.Errorf("Must not invoke next handler in middleware chain")
		})
	})

	router.ServeHTTP(w, r)

	if got, want := w.Code, http.StatusNotFound; got != want {
		t.Errorf("Want status code %d, got %d", want, got)
	}

	got, want := new(errors.Error), errors.ErrNotFound
	json.NewDecoder(w.Body).Decode(got)
	if diff := cmp.Diff(got, want); len(diff) != 0 {
		t.Errorf(diff)
	}
}

// this test verifies the the next handler in the middleware
// chain is processed if the user has read access to the
// repository.
func TestCheckReadAccess(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	readAccess := &core.Perm{
		Synced: time.Now().Unix(),
		Read:   true,
		Write:  false,
		Admin:  false,
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/repos/octocat/hello-world", nil)
	r = r.WithContext(
		request.WithUser(r.Context(), mockUser),
	)
	r = r.WithContext(
		request.WithPerm(
			request.WithUser(
				request.WithRepo(noContext, mockRepo),
				mockUser,
			),
			readAccess,
		),
	)

	router := chi.NewRouter()
	router.Route("/api/repos/{owner}/{name}", func(router chi.Router) {
		router.Use(CheckReadAccess())
		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})
	})

	router.ServeHTTP(w, r)

	if got, want := w.Code, http.StatusTeapot; got != want {
		t.Errorf("Want status code %d, got %d", want, got)
	}
}

// this test verifies that a 404 not found error is written to
// the response if the user lacks read access to the repository.
func TestCheckReadAccess_InsufficientPermissions(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	noAccess := &core.Perm{
		Synced: time.Now().Unix(),
		Read:   false,
		Write:  false,
		Admin:  false,
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/repos/octocat/hello-world", nil)
	r = r.WithContext(
		request.WithPerm(
			request.WithUser(
				request.WithRepo(noContext, mockRepo),
				mockUser,
			),
			noAccess,
		),
	)

	router := chi.NewRouter()
	router.Route("/api/repos/{owner}/{name}", func(router chi.Router) {
		router.Use(CheckReadAccess())
		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			t.Errorf("Must not invoke next handler in middleware chain")
		})
	})

	router.ServeHTTP(w, r)

	if got, want := w.Code, http.StatusNotFound; got != want {
		t.Errorf("Want status code %d, got %d", want, got)
	}

	got, want := new(errors.Error), errors.ErrNotFound
	json.NewDecoder(w.Body).Decode(got)
	if diff := cmp.Diff(got, want); len(diff) != 0 {
		t.Errorf(diff)
	}
}

// this test verifies the the next handler in the middleware
// chain is processed if the user has write access to the
// repository.
func TestCheckWriteAccess(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	writeAccess := &core.Perm{
		Synced: time.Now().Unix(),
		Read:   true,
		Write:  true,
		Admin:  false,
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/repos/octocat/hello-world", nil)
	r = r.WithContext(
		request.WithPerm(
			request.WithUser(
				request.WithRepo(noContext, mockRepo),
				mockUser,
			),
			writeAccess,
		),
	)

	router := chi.NewRouter()
	router.Route("/api/repos/{owner}/{name}", func(router chi.Router) {
		router.Use(CheckWriteAccess())
		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})
	})

	router.ServeHTTP(w, r)

	if got, want := w.Code, http.StatusTeapot; got != want {
		t.Errorf("Want status code %d, got %d", want, got)
	}
}

// this test verifies the the next handler in the middleware
// chain is not processed if the user has write access BUT
// has been inactivated (e.g. blocked).
func TestCheckWriteAccess_InactiveUser(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	writeAccess := &core.Perm{
		Synced: time.Now().Unix(),
		Read:   true,
		Write:  true,
		Admin:  false,
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/repos/octocat/hello-world", nil)
	r = r.WithContext(
		request.WithPerm(
			request.WithUser(
				request.WithRepo(noContext, mockRepo),
				mockUserInactive,
			),
			writeAccess,
		),
	)

	router := chi.NewRouter()
	router.Route("/api/repos/{owner}/{name}", func(router chi.Router) {
		router.Use(CheckWriteAccess())
		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			t.Error("should not invoke handler")
		})
	})

	router.ServeHTTP(w, r)

	if got, want := w.Code, http.StatusForbidden; got != want {
		t.Errorf("Want status code %d, got %d", want, got)
	}
}

// this test verifies that a 404 not found error is written to
// the response if the user lacks write access to the repository.
//
// TODO(bradrydzewski) we should consider returning a 403 forbidden
// if the user has read access.
func TestCheckWriteAccess_InsufficientPermissions(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	noAccess := &core.Perm{
		Synced: time.Now().Unix(),
		Read:   true,
		Write:  false,
		Admin:  false,
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/repos/octocat/hello-world", nil)
	r = r.WithContext(
		request.WithPerm(
			request.WithUser(
				request.WithRepo(noContext, mockRepo),
				mockUser,
			),
			noAccess,
		),
	)

	router := chi.NewRouter()
	router.Route("/api/repos/{owner}/{name}", func(router chi.Router) {
		router.Use(CheckWriteAccess())
		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			t.Errorf("Must not invoke next handler in middleware chain")
		})
	})

	router.ServeHTTP(w, r)

	if got, want := w.Code, http.StatusNotFound; got != want {
		t.Errorf("Want status code %d, got %d", want, got)
	}

	got, want := new(errors.Error), errors.ErrNotFound
	json.NewDecoder(w.Body).Decode(got)
	if diff := cmp.Diff(got, want); len(diff) != 0 {
		t.Errorf(diff)
	}
}

// this test verifies the the next handler in the middleware
// chain is processed if the user has admin access to the
// repository.
func TestCheckAdminAccess(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	noAccess := &core.Perm{
		Synced: time.Now().Unix(),
		Read:   true,
		Write:  true,
		Admin:  true,
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/repos/octocat/hello-world", nil)
	r = r.WithContext(
		request.WithPerm(
			request.WithUser(
				request.WithRepo(noContext, mockRepo),
				mockUser,
			),
			noAccess,
		),
	)

	router := chi.NewRouter()
	router.Route("/api/repos/{owner}/{name}", func(router chi.Router) {
		router.Use(CheckAdminAccess())
		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})
	})

	router.ServeHTTP(w, r)

	if got, want := w.Code, http.StatusTeapot; got != want {
		t.Errorf("Want status code %d, got %d", want, got)
	}
}

// this test verifies that a 404 not found error is written to
// the response if the user lacks admin access to the repository.
//
// TODO(bradrydzewski) we should consider returning a 403 forbidden
// if the user has read access.
func TestCheckAdminAccess_InsufficientPermissions(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	noAccess := &core.Perm{
		Synced: time.Now().Unix(),
		Read:   true,
		Write:  true,
		Admin:  false,
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/repos/octocat/hello-world", nil)
	r = r.WithContext(
		request.WithPerm(
			request.WithUser(
				request.WithRepo(noContext, mockRepo),
				mockUser,
			),
			noAccess,
		),
	)

	router := chi.NewRouter()
	router.Route("/api/repos/{owner}/{name}", func(router chi.Router) {
		router.Use(CheckAdminAccess())
		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			t.Errorf("Must not invoke next handler in middleware chain")
		})
	})

	router.ServeHTTP(w, r)

	if got, want := w.Code, http.StatusNotFound; got != want {
		t.Errorf("Want status code %d, got %d", want, got)
	}

	got, want := new(errors.Error), errors.ErrNotFound
	json.NewDecoder(w.Body).Decode(got)
	if diff := cmp.Diff(got, want); len(diff) != 0 {
		t.Errorf(diff)
	}
}

// this test verifies the the next handler in the middleware
// chain is processed if the authenticated user is a system
// administrator.
func TestCheckAdminAccess_SystemAdmin(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	user := &core.User{ID: 1, Admin: true, Active: true}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/repos/octocat/hello-world", nil)
	r = r.WithContext(
		request.WithUser(r.Context(), user),
	)

	router := chi.NewRouter()
	router.Route("/api/repos/{owner}/{name}", func(router chi.Router) {
		router.Use(CheckAdminAccess())
		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTeapot)
		})
	})

	router.ServeHTTP(w, r)

	if got, want := w.Code, http.StatusTeapot; got != want {
		t.Errorf("Want status code %d, got %d", want, got)
	}
}

// this test verifies that a 401 unauthorized error is written to
// the response if the client is not authenticated and write
// access is required.
func TestCheckAccess_Guest_Write(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/repos/octocat/hello-world", nil)
	r = r.WithContext(
		request.WithRepo(noContext, mockRepo),
	)

	router := chi.NewRouter()
	router.Route("/api/repos/{owner}/{name}", func(router chi.Router) {
		router.Use(CheckAccess(true, true, false))
		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			t.Errorf("Must not invoke next handler in middleware chain")
		})
	})
	router.ServeHTTP(w, r)

	if got, want := w.Code, http.StatusUnauthorized; got != want {
		t.Errorf("Want status code %d, got %d", want, got)
	}

	got, want := new(errors.Error), errors.ErrUnauthorized
	json.NewDecoder(w.Body).Decode(got)
	if diff := cmp.Diff(got, want); len(diff) != 0 {
		t.Errorf(diff)
	}
}

// this test verifies that a 401 unauthorized error is written to
// the response if the client is not authenticated and admin
// access is required.
func TestCheckAccess_Guest_Admin(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/repos/octocat/hello-world", nil)
	r = r.WithContext(
		request.WithRepo(noContext, mockRepo),
	)

	router := chi.NewRouter()
	router.Route("/api/repos/{owner}/{name}", func(router chi.Router) {
		router.Use(CheckAccess(true, false, true))
		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			t.Errorf("Must not invoke next handler in middleware chain")
		})
	})
	router.ServeHTTP(w, r)

	if got, want := w.Code, http.StatusUnauthorized; got != want {
		t.Errorf("Want status code %d, got %d", want, got)
	}

	got, want := new(errors.Error), errors.ErrUnauthorized
	json.NewDecoder(w.Body).Decode(got)
	if diff := cmp.Diff(got, want); len(diff) != 0 {
		t.Errorf(diff)
	}
}

// // this test verifies the the next handler in the middleware
// // chain is processed if the authenticated has read permissions
// // that are successfully synchronized with the source.
// func TestCheckAccess_RefreshPerms(t *testing.T) {
// 	controller := gomock.NewController(t)
// 	defer controller.Finish()

// 	expiredAccess := &core.Perm{
// 		Synced: 0,
// 		Read:   false,
// 		Write:  false,
// 		Admin:  false,
// 	}

// 	updatedAccess := &core.Perm{
// 		Read:  true,
// 		Write: true,
// 		Admin: true,
// 	}

// 	checkPermUpdate := func(ctx context.Context, perm *core.Perm) {
// 		if perm.Synced == 0 {
// 			t.Errorf("Expect synced timestamp updated")
// 		}
// 		if perm.Read == false {
// 			t.Errorf("Expect Read flag updated")
// 		}
// 		if perm.Write == false {
// 			t.Errorf("Expect Write flag updated")
// 		}
// 		if perm.Admin == false {
// 			t.Errorf("Expect Admin flag updated")
// 		}
// 	}

// 	repos := mock.NewMockRepositoryStore(controller)
// 	repos.EXPECT().FindName(gomock.Any(), "octocat", "hello-world").Return(mockRepo, nil)

// 	perms := mock.NewMockPermStore(controller)
// 	perms.EXPECT().Find(gomock.Any(), mockRepo.UID, mockUser.ID).Return(expiredAccess, nil)
// 	perms.EXPECT().Update(gomock.Any(), expiredAccess).Return(nil).Do(checkPermUpdate)

// 	service := mock.NewMockRepositoryService(controller)
// 	service.EXPECT().FindPerm(gomock.Any(), "octocat/hello-world").Return(updatedAccess, nil)

// 	factory := mock.NewMockRepositoryServiceFactory(controller)
// 	factory.EXPECT().Create(mockUser).Return(service)

// 	w := httptest.NewRecorder()
// 	r := httptest.NewRequest("GET", "/api/repos/octocat/hello-world", nil)
// 	r = r.WithContext(
// 		request.WithUser(r.Context(), mockUser),
// 	)

// 	router := chi.NewRouter()
// 	router.Route("/api/repos/{owner}/{name}", func(router chi.Router) {
// 		router.Use(CheckReadAccess(factory, repos, perms))
// 		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
// 			w.WriteHeader(http.StatusTeapot)
// 		})
// 	})

// 	router.ServeHTTP(w, r)

// 	if got, want := w.Code, http.StatusTeapot; got != want {
// 		t.Errorf("Want status code %d, got %d", want, got)
// 	}
// }

// // this test verifies that a 404 not found error is written to
// // the response if the user permissions are expired and the
// // updated permissions cannot be fetched.
// func TestCheckAccess_RefreshPerms_Error(t *testing.T) {
// 	controller := gomock.NewController(t)
// 	defer controller.Finish()

// 	expiredAccess := &core.Perm{
// 		Synced: 0,
// 		Read:   false,
// 		Write:  false,
// 		Admin:  false,
// 	}

// 	repos := mock.NewMockRepositoryStore(controller)
// 	repos.EXPECT().FindName(gomock.Any(), "octocat", "hello-world").Return(mockRepo, nil)

// 	perms := mock.NewMockPermStore(controller)
// 	perms.EXPECT().Find(gomock.Any(), mockRepo.UID, mockUser.ID).Return(expiredAccess, nil)

// 	service := mock.NewMockRepositoryService(controller)
// 	service.EXPECT().FindPerm(gomock.Any(), "octocat/hello-world").Return(nil, io.EOF)

// 	factory := mock.NewMockRepositoryServiceFactory(controller)
// 	factory.EXPECT().Create(mockUser).Return(service)

// 	w := httptest.NewRecorder()
// 	r := httptest.NewRequest("GET", "/api/repos/octocat/hello-world", nil)
// 	r = r.WithContext(
// 		request.WithUser(r.Context(), mockUser),
// 	)

// 	router := chi.NewRouter()
// 	router.Route("/api/repos/{owner}/{name}", func(router chi.Router) {
// 		router.Use(CheckReadAccess(factory, repos, perms))
// 		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
// 			w.WriteHeader(http.StatusTeapot)
// 		})
// 	})

// 	router.ServeHTTP(w, r)
// 	if got, want := w.Code, 404; got != want {
// 		t.Errorf("Want status code %d, got %d", want, got)
// 	}
// }

// // this test verifies the the next handler in the middleware
// // chain is processed if the user permissions are expired,
// // updated permissions are fetched, but fail the changes fail
// // to persist to the database. We know the user has access,
// // so we allow them to proceed even in the event of a failure.
// func TestCheckAccess_RefreshPerms_CannotSave(t *testing.T) {
// 	controller := gomock.NewController(t)
// 	defer controller.Finish()

// 	expiredAccess := &core.Perm{
// 		Synced: 0,
// 		Read:   false,
// 		Write:  false,
// 		Admin:  false,
// 	}

// 	updatedAccess := &core.Perm{
// 		Read:  true,
// 		Write: true,
// 		Admin: true,
// 	}

// 	service := mock.NewMockRepositoryService(controller)
// 	service.EXPECT().FindPerm(gomock.Any(), "octocat/hello-world").Return(updatedAccess, nil)

// 	factory := mock.NewMockRepositoryServiceFactory(controller)
// 	factory.EXPECT().Create(mockUser).Return(service)

// 	repos := mock.NewMockRepositoryStore(controller)
// 	repos.EXPECT().FindName(gomock.Any(), "octocat", "hello-world").Return(mockRepo, nil)

// 	perms := mock.NewMockPermStore(controller)
// 	perms.EXPECT().Find(gomock.Any(), mockRepo.UID, mockUser.ID).Return(expiredAccess, nil)
// 	perms.EXPECT().Update(gomock.Any(), expiredAccess).Return(io.EOF)

// 	w := httptest.NewRecorder()
// 	r := httptest.NewRequest("GET", "/api/repos/octocat/hello-world", nil)
// 	r = r.WithContext(
// 		request.WithUser(r.Context(), mockUser),
// 	)

// 	router := chi.NewRouter()
// 	router.Route("/api/repos/{owner}/{name}", func(router chi.Router) {
// 		router.Use(CheckReadAccess(factory, repos, perms))
// 		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
// 			w.WriteHeader(http.StatusTeapot)
// 		})
// 	})

// 	router.ServeHTTP(w, r)
// 	if got, want := w.Code, http.StatusTeapot; got != want {
// 		t.Errorf("Want status code %d, got %d", want, got)
// 	}
// }
