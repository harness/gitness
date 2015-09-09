package server

import (
	"net/http"
	"time"

	"github.com/drone/drone/Godeps/_workspace/src/github.com/gin-gonic/gin"

	"github.com/drone/drone/pkg/bus"
	"github.com/drone/drone/pkg/queue"
	"github.com/drone/drone/pkg/remote"
	"github.com/drone/drone/pkg/runner"
	"github.com/drone/drone/pkg/store"
	"github.com/drone/drone/pkg/token"
	common "github.com/drone/drone/pkg/types"
)

func SetQueue(q queue.Queue) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("queue", q)
		c.Next()
	}
}

func ToQueue(c *gin.Context) queue.Queue {
	v, ok := c.Get("queue")
	if !ok {
		return nil
	}
	return v.(queue.Queue)
}

func SetBus(r bus.Bus) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("bus", r)
		c.Next()
	}
}

func ToBus(c *gin.Context) bus.Bus {
	v, ok := c.Get("bus")
	if !ok {
		return nil
	}
	return v.(bus.Bus)
}

func ToRemote(c *gin.Context) remote.Remote {
	v, ok := c.Get("remote")
	if !ok {
		return nil
	}
	return v.(remote.Remote)
}

func SetRemote(r remote.Remote) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("remote", r)
		c.Next()
	}
}

func ToRunner(c *gin.Context) runner.Runner {
	v, ok := c.Get("runner")
	if !ok {
		return nil
	}
	return v.(runner.Runner)
}

func SetRunner(r runner.Runner) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("runner", r)
		c.Next()
	}
}

func ToPerm(c *gin.Context) *common.Perm {
	v, ok := c.Get("perm")
	if !ok {
		return nil
	}
	return v.(*common.Perm)
}

func ToUser(c *gin.Context) *common.User {
	v, ok := c.Get("user")
	if !ok {
		return nil
	}
	return v.(*common.User)
}

func ToRepo(c *gin.Context) *common.Repo {
	v, ok := c.Get("repo")
	if !ok {
		return nil
	}
	return v.(*common.Repo)
}

func ToDatastore(c *gin.Context) store.Store {
	return c.MustGet("datastore").(store.Store)
}

func SetDatastore(ds store.Store) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("datastore", ds)
		c.Next()
	}
}

func SetUser() gin.HandlerFunc {
	return func(c *gin.Context) {

		var store = ToDatastore(c)
		var user *common.User

		_, err := token.ParseRequest(c.Request, func(t *token.Token) (string, error) {
			var err error
			user, err = store.UserLogin(t.Text)
			if err != nil {
				return "", err
			}
			return user.Hash, nil
		})

		if err == nil && user != nil && user.ID != 0 {
			c.Set("user", user)
		}
		c.Next()
	}
}

func SetRepo() gin.HandlerFunc {
	return func(c *gin.Context) {
		ds := ToDatastore(c)
		u := ToUser(c)
		owner := c.Params.ByName("owner")
		name := c.Params.ByName("name")
		r, err := ds.RepoName(owner, name)
		switch {
		case err != nil && u != nil:
			c.Fail(404, err)
			return
		case err != nil && u == nil:
			c.Fail(401, err)
			return
		}
		c.Set("repo", r)
		c.Next()
	}
}

func SetPerm() gin.HandlerFunc {
	return func(c *gin.Context) {
		remote := ToRemote(c)
		user := ToUser(c)
		repo := ToRepo(c)
		perm := perms(remote, user, repo)
		c.Set("perm", perm)
		c.Next()
	}
}

func MustUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		u := ToUser(c)
		if u == nil {
			c.AbortWithStatus(401)
		} else {
			c.Set("user", u)
			c.Next()
		}
	}
}

func MustAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		u := ToUser(c)
		if u == nil {
			c.AbortWithStatus(401)
		} else if !u.Admin {
			c.AbortWithStatus(403)
		} else {
			c.Set("user", u)
			c.Next()
		}
	}
}

func CheckPull() gin.HandlerFunc {
	return func(c *gin.Context) {
		u := ToUser(c)
		m := ToPerm(c)

		switch {
		case u == nil && m == nil:
			c.AbortWithStatus(401)
		case u == nil && m.Pull == false:
			c.AbortWithStatus(401)
		case u != nil && m.Pull == false:
			c.AbortWithStatus(404)
		default:
			c.Next()
		}
	}
}

func CheckPush() gin.HandlerFunc {
	return func(c *gin.Context) {
		switch c.Request.Method {
		case "GET", "OPTIONS":
			c.Next()
			return
		}

		u := ToUser(c)
		m := ToPerm(c)

		switch {
		case u == nil && m.Push == false:
			c.AbortWithStatus(401)
		case u != nil && m.Push == false:
			c.AbortWithStatus(404)
		default:
			c.Next()
		}
	}
}

func SetHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Writer.Header().Add("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Add("X-Frame-Options", "DENY")
		c.Writer.Header().Add("X-Content-Type-Options", "nosniff")
		c.Writer.Header().Add("X-XSS-Protection", "1; mode=block")
		c.Writer.Header().Add("Cache-Control", "no-cache")
		c.Writer.Header().Add("Cache-Control", "no-store")
		c.Writer.Header().Add("Cache-Control", "max-age=0")
		c.Writer.Header().Add("Cache-Control", "must-revalidate")
		c.Writer.Header().Add("Cache-Control", "value")
		c.Writer.Header().Set("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
		c.Writer.Header().Set("Expires", "Thu, 01 Jan 1970 00:00:00 GMT")
		if c.Request.TLS != nil {
			c.Writer.Header().Add("Strict-Transport-Security", "max-age=31536000")
		}

		if c.Request.Method == "OPTIONS" {
			c.Writer.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,OPTIONS")
			c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization")
			c.Writer.Header().Set("Allow", "HEAD,GET,POST,PUT,PATCH,DELETE,OPTIONS")
			c.Writer.Header().Set("Content-Type", "application/json")
			c.Writer.WriteHeader(200)
			return
		}

		c.Next()
	}
}
