package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/drone/drone/server/database"
	"github.com/drone/drone/server/session"
	"github.com/drone/drone/shared/httputil"
	"github.com/drone/drone/shared/model"
	"github.com/gorilla/pat"
)

type CommitHandler struct {
	users   database.UserManager
	perms   database.PermManager
	repos   database.RepoManager
	commits database.CommitManager
	builds  database.BuildManager
	sess    session.Session
	queue   chan *model.Request
}

func NewCommitHandler(users database.UserManager, repos database.RepoManager, commits database.CommitManager, builds database.BuildManager, perms database.PermManager, sess session.Session, queue chan *model.Request) *CommitHandler {
	return &CommitHandler{users, perms, repos, commits, builds, sess, queue}
}

// GetFeed gets recent commits for the repository and branch
// GET /v1/repos/{host}/{owner}/{name}/branches/{branch}/commits
func (h *CommitHandler) GetFeed(w http.ResponseWriter, r *http.Request) error {
	var host, owner, name = parseRepo(r)
	var branch = r.FormValue(":branch")

	// get the user form the session.
	user := h.sess.User(r)

	// get the repository from the database.
	repo, err := h.repos.FindName(host, owner, name)
	switch {
	case err != nil && user == nil:
		return notAuthorized{}
	case err != nil && user != nil:
		return notFound{}
	}

	// user must have read access to the repository.
	ok, _ := h.perms.Read(user, repo)
	switch {
	case ok == false && user == nil:
		return notAuthorized{}
	case ok == false && user != nil:
		return notFound{}
	}

	commits, err := h.commits.ListBranch(repo.ID, branch)
	if err != nil {
		return notFound{err}
	}

	return json.NewEncoder(w).Encode(commits)
}

// GetCommit gets the commit for the repository, branch and sha.
// GET /v1/repos/{host}/{owner}/{name}/branches/{branch}/commits/{commit}
func (h *CommitHandler) GetCommit(w http.ResponseWriter, r *http.Request) error {
	var host, owner, name = parseRepo(r)
	var branch = r.FormValue(":branch")
	var sha = r.FormValue(":commit")

	// get the user form the session.
	user := h.sess.User(r)

	// get the repository from the database.
	repo, err := h.repos.FindName(host, owner, name)
	switch {
	case err != nil && user == nil:
		return notAuthorized{}
	case err != nil && user != nil:
		return notFound{}
	}

	// user must have read access to the repository.
	ok, _ := h.perms.Read(user, repo)
	switch {
	case ok == false && user == nil:
		return notAuthorized{}
	case ok == false && user != nil:
		return notFound{}
	}

	commit, err := h.commits.FindSha(repo.ID, branch, sha)
	if err != nil {
		return notFound{err}
	}

	return json.NewEncoder(w).Encode(commit)
}

// PostCommit gets the commit for the repository and schedules to re-build.
// GET /v1/repos/{host}/{owner}/{name}/branches/{branch}/commits/{commit}
func (h *CommitHandler) PostCommit(w http.ResponseWriter, r *http.Request) error {
	var host, owner, name = parseRepo(r)
	var branch = r.FormValue(":branch")
	var sha = r.FormValue(":commit")

	// get the user form the session.
	user := h.sess.User(r)
	if user == nil {
		return notAuthorized{}
	}

	// get the repo from the database
	repo, err := h.repos.FindName(host, owner, name)
	switch {
	case err != nil && user == nil:
		return notAuthorized{}
	case err != nil && user != nil:
		return notFound{}
	}

	// user must have admin access to the repository.
	if ok, _ := h.perms.Admin(user, repo); !ok {
		return notFound{err}
	}

	c, err := h.commits.FindSha(repo.ID, branch, sha)
	if err != nil {
		return notFound{err}
	}

	// we can't start an already started build
	if c.Status == model.StatusStarted || c.Status == model.StatusEnqueue {
		return badRequest{errors.New("This commit already builds")}
	}

	c.Status = model.StatusEnqueue
	c.Started = 0
	c.Finished = 0
	c.Duration = 0
	if err := h.commits.Update(c); err != nil {
		return internalServerError{err}
	}

	repoOwner, err := h.users.Find(repo.UserID)
	if err != nil {
		return badRequest{err}
	}

	builds, err := h.builds.FindCommit(c.ID)
	if err != nil {
		return notFound{err}
	}

	for _, build := range builds {
		if build.Status == model.StatusStarted || build.Status == model.StatusEnqueue {
			return badRequest{errors.New("This build already builds")}
		} else {
			build.Status = model.StatusEnqueue
			build.Started = 0
			build.Finished = 0
			build.Duration = 0
			h.builds.Update(build)
		}
	}

	// drop the items on the queue
	go func() {
		h.queue <- &model.Request{
			User:   repoOwner,
			Host:   httputil.GetURL(r),
			Repo:   repo,
			Commit: c,
			Builds: builds,
		}
	}()

	w.WriteHeader(http.StatusOK)
	return nil
}

func (h *CommitHandler) Register(r *pat.Router) {
	r.Get("/v1/repos/{host}/{owner}/{name}/branches/{branch}/commits/{commit}", errorHandler(h.GetCommit))
	r.Post("/v1/repos/{host}/{owner}/{name}/branches/{branch}/commits/{commit}", errorHandler(h.PostCommit)).Queries("action", "rebuild")
	r.Get("/v1/repos/{host}/{owner}/{name}/branches/{branch}/commits", errorHandler(h.GetFeed))
}
