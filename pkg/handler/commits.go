package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/drone/drone/pkg/channel"
	"github.com/drone/drone/pkg/database"
	. "github.com/drone/drone/pkg/model"
)

// Display a specific Commit.
func CommitShow(w http.ResponseWriter, r *http.Request, u *User, repo *Repo) error {
	branch := r.FormValue("branch")
	if branch == "" {
		branch = "master"
	}

	hash := r.FormValue(":commit")
	labl := r.FormValue(":label")

	// get the commit from the database
	commit, err := database.GetCommitBranchHash(branch, hash, repo.ID)
	if err != nil {
		return err
	}

	// get the builds from the database. a commit can have
	// multiple sub-builds (or matrix builds)
	builds, err := database.ListBuilds(commit.ID)
	if err != nil {
		return err
	}

	data := struct {
		User   *User
		Repo   *Repo
		Commit *Commit
		Build  *Build
		Builds []*Build
		Token  string
	}{u, repo, commit, builds[0], builds, ""}

	// get the specific build requested by the user. instead
	// of a database round trip, we can just loop through the
	// list and extract the requested build.
	for _, b := range builds {
		if b.Slug == labl {
			data.Build = b
			break
		}
	}

	// generate a token to connect with the websocket
	// handler and stream output, if the build is running.
	data.Token = channel.Token(fmt.Sprintf(
		"%s/%s/%s/commit/%s/%s/builds/%s", repo.Host, repo.Owner, repo.Name, commit.Branch, commit.Hash, builds[0].Slug))

	// render the repository template.
	return RenderTemplate(w, "repo_commit.html", &data)
}

// Helper method for saving a failed build or commit in the case where it never starts to build.
// This can happen if the yaml is bad or doesn't exist.
func saveFailedBuild(commit *Commit, msg string) error {

	// Set the commit to failed
	commit.Status = "Failure"
	commit.Created = time.Now().UTC()
	commit.Finished = commit.Created
	commit.Duration = 0
	if err := database.SaveCommit(commit); err != nil {
		return err
	}

	// save the build to the database
	build := &Build{}
	build.Slug = "1" // TODO: This should not be hardcoded
	build.CommitID = commit.ID
	build.Created = time.Now().UTC()
	build.Finished = build.Created
	commit.Duration = 0
	build.Status = "Failure"
	build.Stdout = msg
	if err := database.SaveBuild(build); err != nil {
		return err
	}

	// TODO: Should the status be Error instead of Failure?

	// TODO: Do we need to update the branch table too?

	return nil

}
