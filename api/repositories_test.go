package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"time"

	. "gopkg.in/check.v1"

	"github.com/hoffie/larasync/repository"
)

type RepoListTests struct {
	server *Server
	req    *http.Request
	repos  string
}

var _ = Suite(&RepoListTests{})

func (t *RepoListTests) SetUpTest(c *C) {
	req, err := http.NewRequest("GET", "http://example.org/repositories", nil)
	c.Assert(err, IsNil)
	t.req = req
}

func (t *RepoListTests) SetUpSuite(c *C) {
	t.repos = c.MkDir()
	rm, err := repository.NewManager(t.repos)
	c.Assert(err, IsNil)
	t.server = New(adminPubkey, time.Minute, rm)
}

func (t *RepoListTests) getResponse(req *http.Request) *httptest.ResponseRecorder {
	rw := httptest.NewRecorder()
	t.server.router.ServeHTTP(rw, req)
	return rw
}

func (t *RepoListTests) TestRepoListUnauthorized(c *C) {
	resp := t.getResponse(t.req)
	c.Assert(resp.Code, Equals, 401)
	c.Assert(resp.Body.String(), Equals, "Unauthorized\n")
}

func (t *RepoListTests) TestRepoListAdmin(c *C) {
	SignAsAdmin(t.req, adminSecret)
	resp := t.getResponse(t.req)
	c.Assert(resp.Code, Equals, 200)
}

func (t *RepoListTests) TestRepoListOutput(c *C) {
	SignAsAdmin(t.req, adminSecret)
	resp := t.getResponse(t.req)
	//FIXME test repo list output
	c.Assert(resp.Code, Equals, 200)
	c.Assert(resp.Body.String(), Equals, "[]")
	c.Assert(resp.Body.Len(), Not(Equals), 0)
}

func (t *RepoListTests) TestRepoListOutputExcludeFiles(c *C) {
	f, err := os.Create(filepath.Join(t.repos, "somefile"))
	c.Assert(err, IsNil)
	f.Close()
	SignAsAdmin(t.req, adminSecret)
	resp := t.getResponse(t.req)
	c.Assert(resp.Code, Equals, 200)
	c.Assert(resp.Body.String(), Equals, "[]")
	c.Assert(resp.Body.Len(), Not(Equals), 0)
}
