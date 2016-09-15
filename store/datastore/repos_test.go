package datastore

import (
	"fmt"
	"testing"

	"github.com/drone/drone/model"
	"github.com/franela/goblin"
)

func TestRepos(t *testing.T) {
	db := openTest()
	defer db.Close()

	s := From(db)
	g := goblin.Goblin(t)
	g.Describe("Repo", func() {

		// before each test be sure to purge the package
		// table data from the database.
		g.BeforeEach(func() {
			db.Exec("DELETE FROM builds")
			db.Exec("DELETE FROM repos")
			db.Exec("DELETE FROM users")
		})

		g.It("Should Set a Repo", func() {
			repo := model.Repo{
				UserID:   1,
				FullName: "bradrydzewski/drone",
				Owner:    "bradrydzewski",
				Name:     "drone",
			}
			err1 := s.CreateRepo(&repo)
			err2 := s.UpdateRepo(&repo)
			getrepo, err3 := s.GetRepo(repo.ID)

			g.Assert(err1 == nil).IsTrue()
			g.Assert(err2 == nil).IsTrue()
			g.Assert(err3 == nil).IsTrue()
			g.Assert(repo.ID).Equal(getrepo.ID)
		})

		g.It("Should Add a Repo", func() {
			repo := model.Repo{
				UserID:   1,
				FullName: "bradrydzewski/drone",
				Owner:    "bradrydzewski",
				Name:     "drone",
			}
			err := s.CreateRepo(&repo)
			g.Assert(err == nil).IsTrue()
			g.Assert(repo.ID != 0).IsTrue()
		})

		g.It("Should Get a Repo by ID", func() {
			repo := model.Repo{
				UserID:   1,
				FullName: "bradrydzewski/drone",
				Owner:    "bradrydzewski",
				Name:     "drone",
			}
			s.CreateRepo(&repo)
			getrepo, err := s.GetRepo(repo.ID)
			g.Assert(err == nil).IsTrue()
			g.Assert(repo.ID).Equal(getrepo.ID)
			g.Assert(repo.UserID).Equal(getrepo.UserID)
			g.Assert(repo.Owner).Equal(getrepo.Owner)
			g.Assert(repo.Name).Equal(getrepo.Name)
		})

		g.It("Should Get a Repo by Name", func() {
			repo := model.Repo{
				UserID:   1,
				FullName: "bradrydzewski/drone",
				Owner:    "bradrydzewski",
				Name:     "drone",
			}
			s.CreateRepo(&repo)
			getrepo, err := s.GetRepoName(repo.FullName)
			g.Assert(err == nil).IsTrue()
			g.Assert(repo.ID).Equal(getrepo.ID)
			g.Assert(repo.UserID).Equal(getrepo.UserID)
			g.Assert(repo.Owner).Equal(getrepo.Owner)
			g.Assert(repo.Name).Equal(getrepo.Name)
		})

		g.It("Should Get a Repo List", func() {
			repo1 := &model.Repo{
				UserID:   1,
				Owner:    "bradrydzewski",
				Name:     "drone",
				FullName: "bradrydzewski/drone",
			}
			repo2 := &model.Repo{
				UserID:   2,
				Owner:    "drone",
				Name:     "drone",
				FullName: "drone/drone",
			}
			repo3 := &model.Repo{
				UserID:   2,
				Owner:    "octocat",
				Name:     "hello-world",
				FullName: "octocat/hello-world",
			}
			s.CreateRepo(repo1)
			s.CreateRepo(repo2)
			s.CreateRepo(repo3)

			repos, err := s.GetRepoListOf([]*model.RepoLite{
				{FullName: "bradrydzewski/drone"},
				{FullName: "drone/drone"},
			})
			g.Assert(err == nil).IsTrue()
			g.Assert(len(repos)).Equal(2)
			g.Assert(repos[0].ID).Equal(repo1.ID)
			g.Assert(repos[1].ID).Equal(repo2.ID)
		})

		g.It("Should Get a Repo List", func() {
			repo1 := &model.Repo{
				UserID:   1,
				Owner:    "bradrydzewski",
				Name:     "drone",
				FullName: "bradrydzewski/drone",
			}
			repo2 := &model.Repo{
				UserID:   2,
				Owner:    "drone",
				Name:     "drone",
				FullName: "drone/drone",
			}
			s.CreateRepo(repo1)
			s.CreateRepo(repo2)

			count, err := s.GetRepoCount()
			g.Assert(err == nil).IsTrue()
			g.Assert(count).Equal(2)
		})

		g.It("Should Get a Repo List Paginated", func() {
			_repos := []*model.RepoLite{
				{FullName: "bradrydzewski/drone"},
				{FullName: "drone/drone"},
			}
			_repo := []*model.Repo{
				{
					UserID:   1,
					Owner:    "bradrydzewski",
					Name:     "drone",
					FullName: "bradrydzewski/drone",
				},
				{
					UserID:   2,
					Owner:    "drone",
					Name:     "drone",
					FullName: "drone/drone",
				},
			}
			s.CreateRepo(_repo[0])
			s.CreateRepo(_repo[1])

			for i := 0; i < 3000; i++ {
				// find by 3000 repos
				_repos = append(_repos, &model.RepoLite{FullName: "octocat/hello-world" + fmt.Sprint(i)})
				// but create only 2000 repos
				if i >= 1999 && i < 2999 {
					continue
				}
				repo := &model.Repo{
					UserID:   2,
					Owner:    "octocat",
					Name:     "hello-world" + fmt.Sprint(i),
					FullName: "octocat/hello-world" + fmt.Sprint(i),
				}
				_repo = append(_repo, repo)
				s.CreateRepo(repo)
			}
			repos, err := s.GetRepoListOf(_repos)
			g.Assert(err == nil).IsTrue()
			g.Assert(len(repos)).Equal(2002)
			g.Assert(repos[0].ID).Equal(_repo[0].ID)
			g.Assert(repos[1].ID).Equal(_repo[1].ID)
			g.Assert(repos[2001].ID).Equal(_repo[2001].ID)
		})

		g.It("Should Delete a Repo", func() {
			repo := model.Repo{
				UserID:   1,
				FullName: "bradrydzewski/drone",
				Owner:    "bradrydzewski",
				Name:     "drone",
			}
			s.CreateRepo(&repo)
			_, err1 := s.GetRepo(repo.ID)
			err2 := s.DeleteRepo(&repo)
			_, err3 := s.GetRepo(repo.ID)
			g.Assert(err1 == nil).IsTrue()
			g.Assert(err2 == nil).IsTrue()
			g.Assert(err3 == nil).IsFalse()
		})

		g.It("Should Enforce Unique Repo Name", func() {
			repo1 := model.Repo{
				UserID:   1,
				FullName: "bradrydzewski/drone",
				Owner:    "bradrydzewski",
				Name:     "drone",
			}
			repo2 := model.Repo{
				UserID:   2,
				FullName: "bradrydzewski/drone",
				Owner:    "bradrydzewski",
				Name:     "drone",
			}
			err1 := s.CreateRepo(&repo1)
			err2 := s.CreateRepo(&repo2)
			g.Assert(err1 == nil).IsTrue()
			g.Assert(err2 == nil).IsFalse()
		})
	})
}
