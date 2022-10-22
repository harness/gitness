// Copyright 2022 Harness Inc. All rights reserved.
// Use of this source code is governed by the Polyform Free Trial License
// that can be found in the LICENSE.md file for this repository.

package router

import (
	"context"
	"fmt"
	"net/http"

	"github.com/harness/gitness/internal/api/handler/resource"

	"github.com/harness/gitness/internal/api/controller/repo"
	"github.com/harness/gitness/internal/api/controller/serviceaccount"
	"github.com/harness/gitness/internal/api/controller/space"
	"github.com/harness/gitness/internal/api/controller/user"
	"github.com/harness/gitness/internal/api/handler/account"
	handlerrepo "github.com/harness/gitness/internal/api/handler/repo"
	handlerserviceaccount "github.com/harness/gitness/internal/api/handler/serviceaccount"
	handlerspace "github.com/harness/gitness/internal/api/handler/space"
	"github.com/harness/gitness/internal/api/handler/system"
	handleruser "github.com/harness/gitness/internal/api/handler/user"
	"github.com/harness/gitness/internal/api/middleware/accesslog"
	middlewareauthn "github.com/harness/gitness/internal/api/middleware/authn"
	"github.com/harness/gitness/internal/api/middleware/encode"
	"github.com/harness/gitness/internal/api/middleware/principal"
	"github.com/harness/gitness/types/enum"

	"github.com/harness/gitness/internal/api/request"
	"github.com/harness/gitness/internal/auth/authn"
	"github.com/harness/gitness/internal/store"
	"github.com/harness/gitness/types"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog/hlog"
)

// APIHandler is an abstraction of an http handler that handles API calls.
type APIHandler interface {
	http.Handler
}

var (
	// terminatedPathPrefixesAPI is the list of prefixes that will require resolving terminated paths.
	terminatedPathPrefixesAPI = []string{"/v1/spaces/", "/v1/repos/"}
)

// NewAPIHandler returns a new APIHandler.
func NewAPIHandler(
	systemStore store.SystemStore,
	authenticator authn.Authenticator,
	repoCtrl *repo.Controller,
	spaceCtrl *space.Controller,
	saCtrl *serviceaccount.Controller,
	userCtrl *user.Controller) APIHandler {
	config := systemStore.Config(context.Background())

	// Use go-chi router for inner routing.
	r := chi.NewRouter()

	// Apply common api middleware.
	r.Use(middleware.NoCache)
	r.Use(middleware.Recoverer)

	// configure logging middleware.
	r.Use(hlog.URLHandler("url"))
	r.Use(hlog.MethodHandler("method"))
	r.Use(hlog.RequestIDHandler("request", "Request-Id"))
	r.Use(accesslog.HlogHandler())

	// configure cors middleware
	r.Use(corsHandler(config))

	// for now always attempt auth - enforced per operation.
	r.Use(middlewareauthn.Attempt(authenticator))

	r.Route("/v1", func(r chi.Router) {
		setupRoutesV1(r, repoCtrl, spaceCtrl, saCtrl, userCtrl)
	})

	// wrap router in terminatedPath encoder.
	return encode.TerminatedPathBefore(terminatedPathPrefixesAPI, r)
}

func corsHandler(config *types.Config) func(http.Handler) http.Handler {
	return cors.New(
		cors.Options{
			AllowedOrigins:   config.Cors.AllowedOrigins,
			AllowedMethods:   config.Cors.AllowedMethods,
			AllowedHeaders:   config.Cors.AllowedHeaders,
			ExposedHeaders:   config.Cors.ExposedHeaders,
			AllowCredentials: config.Cors.AllowCredentials,
			MaxAge:           config.Cors.MaxAge,
		},
	).Handler
}

func setupRoutesV1(r chi.Router, repoCtrl *repo.Controller, spaceCtrl *space.Controller,
	saCtrl *serviceaccount.Controller, userCtrl *user.Controller) {
	setupSpaces(r, spaceCtrl)
	setupRepos(r, repoCtrl)
	setupUsers(r, userCtrl)
	setupServiceAccounts(r, saCtrl)
	setupAdmin(r, userCtrl)
	setupAccount(r, userCtrl)
	setupSystem(r)
	setupResources(r)
}

func setupSpaces(r chi.Router, spaceCtrl *space.Controller) {
	r.Route("/spaces", func(r chi.Router) {
		// Create takes path and parentId via body, not uri
		r.Post("/", handlerspace.HandleCreate(spaceCtrl))

		r.Route(fmt.Sprintf("/{%s}", request.PathParamSpaceRef), func(r chi.Router) {
			// space operations
			r.Get("/", handlerspace.HandleFind(spaceCtrl))
			r.Put("/", handlerspace.HandleUpdate(spaceCtrl))
			r.Delete("/", handlerspace.HandleDelete(spaceCtrl))

			r.Post("/move", handlerspace.HandleMove(spaceCtrl))
			r.Get("/spaces", handlerspace.HandleListSpaces(spaceCtrl))
			r.Get("/repos", handlerspace.HandleListRepos(spaceCtrl))
			r.Get("/serviceAccounts", handlerspace.HandleListServiceAccounts(spaceCtrl))

			// Child collections
			r.Route("/paths", func(r chi.Router) {
				r.Get("/", handlerspace.HandleListPaths(spaceCtrl))
				r.Post("/", handlerspace.HandleCreatePath(spaceCtrl))

				// per path operations
				r.Route(fmt.Sprintf("/{%s}", request.PathParamPathID), func(r chi.Router) {
					r.Delete("/", handlerspace.HandleDeletePath(spaceCtrl))
				})
			})
		})
	})
}

func setupRepos(r chi.Router, repoCtrl *repo.Controller) {
	r.Route("/repos", func(r chi.Router) {
		// Create takes path and parentId via body, not uri
		r.Post("/", handlerrepo.HandleCreate(repoCtrl))
		r.Route(fmt.Sprintf("/{%s}", request.PathParamRepoRef), func(r chi.Router) {
			// repo level operations
			r.Get("/", handlerrepo.HandleFind(repoCtrl))
			r.Put("/", handlerrepo.HandleUpdate(repoCtrl))
			r.Delete("/", handlerrepo.HandleDelete(repoCtrl))

			r.Post("/move", handlerrepo.HandleMove(repoCtrl))
			r.Get("/serviceAccounts", handlerrepo.HandleListServiceAccounts(repoCtrl))

			// content operations
			// NOTE: this allows /content and /content/ to both be valid (without any other tricks.)
			// We don't expect there to be any other operations in that route (as that could overlap with file names)
			r.Route("/content", func(r chi.Router) {
				r.Get("/*", handlerrepo.HandleGetContent(repoCtrl))
			})

			// commit operations
			r.Route("/commits", func(r chi.Router) {
				r.Get("/", handlerrepo.HandleListCommits(repoCtrl))
			})

			// branch operations
			r.Route("/branches", func(r chi.Router) {
				r.Get("/", handlerrepo.HandleListBranches(repoCtrl))
			})

			// repo path operations
			r.Route("/paths", func(r chi.Router) {
				r.Get("/", handlerrepo.HandleListPaths(repoCtrl))
				r.Post("/", handlerrepo.HandleCreatePath(repoCtrl))

				// per path operations
				r.Route(fmt.Sprintf("/{%s}", request.PathParamPathID), func(r chi.Router) {
					r.Delete("/", handlerrepo.HandleDeletePath(repoCtrl))
				})
			})
		})
	})
}

func setupUsers(r chi.Router, userCtrl *user.Controller) {
	r.Route("/user", func(r chi.Router) {
		// enforce principial authenticated and it's a user
		r.Use(principal.RestrictTo(enum.PrincipalTypeUser))

		r.Get("/", handleruser.HandleFind(userCtrl))
		r.Patch("/", handleruser.HandleUpdate(userCtrl))

		// PAT
		r.Route("/tokens", func(r chi.Router) {
			r.Get("/", handleruser.HandleListTokens(userCtrl, enum.TokenTypePAT))
			r.Post("/", handleruser.HandleCreateAccessToken(userCtrl))

			// per token operations
			r.Route(fmt.Sprintf("/{%s}", request.PathParamTokenID), func(r chi.Router) {
				r.Delete("/", handleruser.HandleDeleteToken(userCtrl, enum.TokenTypePAT))
			})
		})

		// SESSION TOKENS
		r.Route("/sessions", func(r chi.Router) {
			r.Get("/", handleruser.HandleListTokens(userCtrl, enum.TokenTypeSession))

			// per token operations
			r.Route(fmt.Sprintf("/{%s}", request.PathParamTokenID), func(r chi.Router) {
				r.Delete("/", handleruser.HandleDeleteToken(userCtrl, enum.TokenTypeSession))
			})
		})
	})
}

func setupServiceAccounts(r chi.Router, saCtrl *serviceaccount.Controller) {
	r.Route("/serviceAccounts", func(r chi.Router) {
		// create takes parent information via body
		r.Post("/", handlerserviceaccount.HandleCreate(saCtrl))

		r.Route(fmt.Sprintf("/{%s}", request.PathParamServiceAccountUID), func(r chi.Router) {
			r.Get("/", handlerserviceaccount.HandleFind(saCtrl))
			r.Delete("/", handlerserviceaccount.HandleDelete(saCtrl))

			// SAT
			r.Route("/tokens", func(r chi.Router) {
				r.Get("/", handlerserviceaccount.HandleListTokens(saCtrl))
				r.Post("/", handlerserviceaccount.HandleCreateToken(saCtrl))

				// per token operations
				r.Route(fmt.Sprintf("/{%s}", request.PathParamTokenID), func(r chi.Router) {
					r.Delete("/", handlerserviceaccount.HandleDeleteToken(saCtrl))
				})
			})
		})
	})
}

func setupSystem(r chi.Router) {
	r.Route("/system", func(r chi.Router) {
		r.Get("/health", system.HandleHealth)
		r.Get("/version", system.HandleVersion)
	})
}

func setupResources(r chi.Router) {
	r.Route("/resources", func(r chi.Router) {
		r.Get("/gitignore", resource.HandleGitIgnore())
		r.Get("/license", resource.HandleLicence())
	})
}

func setupAdmin(r chi.Router, _ *user.Controller) {
	r.Route("/users", func(r chi.Router) {
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(fmt.Sprintf("Create user '%s'", chi.URLParam(r, "rref"))))
		})

		r.Route(fmt.Sprintf("/{%s}", request.PathParamUserUID), func(r chi.Router) {
			r.Get("/", func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(fmt.Sprintf("Get user '%s'", chi.URLParam(r, "rref"))))
			})
			r.Put("/", func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(fmt.Sprintf("Update user '%s'", chi.URLParam(r, "rref"))))
			})
			r.Delete("/", func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(fmt.Sprintf("Delete user '%s'", chi.URLParam(r, "rref"))))
			})
		})
	})
}

func setupAccount(r chi.Router, userCtrl *user.Controller) {
	r.Post("/login", account.HandleLogin(userCtrl))
	r.Post("/register", account.HandleRegister(userCtrl))
}