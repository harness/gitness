// Copyright 2019 Drone.IO Inc. All rights reserved.
// Use of this source code is governed by the Drone Non-Commercial License
// that can be found in the LICENSE file.

//go:build !oss
// +build !oss

/*

/rpc/v2/stage                       POST  (request)
/rpc/v2/stage/{stage}?machine=      POST  (accept, details)
/rpc/v2/stage/{stage}               PUT   (beforeAll, afterAll)
/rpc/v2/stage/{stage}/steps/{step}  PUT   (before, after)
/rpc/v2/build/{build}/watch         POST  (watch)
/rpc/v2/stage/{stage}/logs/batch    POST  (batch)
/rpc/v2/stage/{stage}/logs/upload   POST  (upload)

*/

package rpc2

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/drone/drone/core"
	"github.com/drone/drone/handler/api/render"
	"github.com/drone/drone/operator/manager"
	"github.com/drone/drone/store/shared/db"
)

// default http request timeout
var defaultTimeout = time.Second * 30

var noContext = context.Background()

// HandleJoin returns an http.HandlerFunc that makes an
// http.Request to join the cluster.
//
// POST /rpc/v2/nodes/:machine
func HandleJoin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeOK(w) // this is a no-op
	}
}

// HandleLeave returns an http.HandlerFunc that makes an
// http.Request to leave the cluster.
//
// DELETE /rpc/v2/nodes/:machine
func HandleLeave() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeOK(w) // this is a no-op
	}
}

// HandlePing returns an http.HandlerFunc that makes an
// http.Request to ping the server and confirm connectivity.
//
// GET /rpc/v2/ping
func HandlePing() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeOK(w) // this is a no-op
	}
}

// HandleRequest returns an http.HandlerFunc that processes an
// http.Request to request a stage from the queue for execution.
//
// POST /rpc/v2/stage
func HandleRequest(m manager.BuildManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
		defer cancel()

		req := new(manager.Request)
		err := json.NewDecoder(r.Body).Decode(req)
		if err != nil {
			writeError(w, err)
			return
		}
		stage, err := m.Request(ctx, req)
		if err != nil {
			writeError(w, err)
		} else {
			writeJSON(w, stage)
		}
	}
}

// HandleAccept returns an http.HandlerFunc that processes an
// http.Request to accept ownership of the stage.
//
// POST /rpc/v2/stage/{stage}?machine=
func HandleAccept(m manager.BuildManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stage, _ := strconv.ParseInt(
			chi.URLParam(r, "stage"), 10, 64)

		out, err := m.Accept(noContext, stage, r.FormValue("machine"))
		if err != nil {
			writeError(w, err)
		} else {
			writeJSON(w, out)
		}
	}
}

// HandleInfo returns an http.HandlerFunc that processes an
// http.Request to get the build details.
//
// POST /rpc/v2/build/{build}
func HandleInfo(m manager.BuildManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		stage, _ := strconv.ParseInt(
			chi.URLParam(r, "stage"), 10, 64)

		res, err := m.Details(noContext, stage)
		if err != nil {
			writeError(w, err)
			return
		}

		netrc, err := m.Netrc(noContext, res.Repo.ID)
		if err != nil {
			writeError(w, err)
			return
		}

		writeJSON(w, &details{
			Context: res,
			Netrc:   netrc,
			Repo: &repository{
				Repository: res.Repo,
				Secret:     res.Repo.Secret,
			},
		})
	}
}

// HandleUpdateStage returns an http.HandlerFunc that processes
// an http.Request to update a stage.
//
// PUT /rpc/v2/stage/{stage}
func HandleUpdateStage(m manager.BuildManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dst := new(core.Stage)
		err := json.NewDecoder(r.Body).Decode(dst)
		if err != nil {
			writeError(w, err)
			return
		}

		if dst.Status == core.StatusPending ||
			dst.Status == core.StatusRunning {
			err = m.BeforeAll(noContext, dst)
		} else {
			err = m.AfterAll(noContext, dst)
		}

		if err != nil {
			writeError(w, err)
		} else {
			writeJSON(w, dst)
		}
	}
}

// HandleUpdateStep returns an http.HandlerFunc that processes
// an http.Request to update a step.
//
// POST /rpc/v2/step/{step}
func HandleUpdateStep(m manager.BuildManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dst := new(core.Step)
		err := json.NewDecoder(r.Body).Decode(dst)
		if err != nil {
			writeError(w, err)
			return
		}

		if dst.Status == core.StatusPending ||
			dst.Status == core.StatusRunning {
			err = m.Before(noContext, dst)
		} else {
			err = m.After(noContext, dst)
		}

		if err != nil {
			writeError(w, err)
		} else {
			writeJSON(w, dst)
		}
	}
}

// HandleWatch returns an http.HandlerFunc that accepts a
// blocking http.Request that watches a build for cancellation
// events.
//
// GET /rpc/v2/build/{build}/watch
func HandleWatch(m manager.BuildManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx, cancel := context.WithTimeout(ctx, defaultTimeout)
		defer cancel()

		build, _ := strconv.ParseInt(
			chi.URLParamFromCtx(ctx, "build"), 10, 64)

		_, err := m.Watch(ctx, build)
		if err != nil {
			writeError(w, err)
		} else {
			writeOK(w)
		}
	}
}

// HandleLogBatch returns an http.HandlerFunc that accepts an
// http.Request to submit a stream of logs to the system.
//
// POST /rpc/v2/step/{step}/logs/batch
func HandleLogBatch(m manager.BuildManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		step, _ := strconv.ParseInt(
			chi.URLParam(r, "step"), 10, 64)

		lines := []*core.Line{}
		err := json.NewDecoder(r.Body).Decode(&lines)
		if err != nil {
			writeError(w, err)
			return
		}

		// TODO(bradrydzewski) modify the write function to
		// accept a slice of lines.
		for _, line := range lines {
			err := m.Write(noContext, step, line)
			if err != nil {
				writeError(w, err)
				return
			}
		}

		writeOK(w)
	}
}

// HandleLogUpload returns an http.HandlerFunc that accepts an
// http.Request to upload and persist logs for a pipeline stage.
//
// POST /rpc/v2/step/{step}/logs/upload
func HandleLogUpload(m manager.BuildManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		step, _ := strconv.ParseInt(
			chi.URLParam(r, "step"), 10, 64)

		err := m.Upload(noContext, step, r.Body)
		if err != nil {
			writeError(w, err)
		} else {
			writeOK(w)
		}
	}
}

// HandleCardUpload returns an http.HandlerFunc that accepts an
// http.Request to upload and persist a card for a pipeline step.
//
// POST /rpc/v2/step/{step}/card
func HandleCardUpload(m manager.BuildManager) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		step, _ := strconv.ParseInt(
			chi.URLParam(r, "step"), 10, 64)

		in := new(core.CardInput)
		err := json.NewDecoder(r.Body).Decode(in)
		if err != nil {
			render.BadRequest(w, err)
			return
		}

		err = m.UploadCard(noContext, step, in)
		if err != nil {
			writeError(w, err)
		} else {
			writeOK(w)
		}
	}
}

// write a 200 Status OK to the response body.
func writeJSON(w http.ResponseWriter, v interface{}) {
	json.NewEncoder(w).Encode(v)
}

// write a 200 Status OK to the response body.
func writeOK(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
}

// write an error message to the response body.
func writeError(w http.ResponseWriter, err error) {
	if err == context.DeadlineExceeded {
		w.WriteHeader(204) // should retry
	} else if err == context.Canceled {
		w.WriteHeader(204) // should retry
	} else if err == db.ErrOptimisticLock {
		w.WriteHeader(409) // should abort
	} else {
		w.WriteHeader(500) // should fail
	}
	io.WriteString(w, err.Error())
}
