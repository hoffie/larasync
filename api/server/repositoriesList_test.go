package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

	"github.com/hoffie/larasync/api/common"

	. "gopkg.in/check.v1"
)

type RepoListTests struct {
	BaseTests
	req *http.Request
}

var _ = Suite(&RepoListTests{
	BaseTests: newBaseTest(),
})

func (t *RepoListTests) SetUpTest(c *C) {
	req, err := http.NewRequest("GET", "http://example.org/repositories", nil)
	c.Assert(err, IsNil)
	t.req = req
	t.createRepoManager(c)
	t.createServer(c)
	t.createRepository(c)
}

func (t *RepoListTests) getResponse(req *http.Request) *httptest.ResponseRecorder {
	rw := httptest.NewRecorder()
	t.server.router.ServeHTTP(rw, req)
	return rw
}

func (t *RepoListTests) TestRepoListUnauthorized(c *C) {
	resp := t.getResponse(t.req)
	c.Assert(resp.Code, Equals, http.StatusUnauthorized)
}

func (t *RepoListTests) TestRepoListAdmin(c *C) {
	common.SignWithPassphrase(t.req, adminSecret)
	resp := t.getResponse(t.req)
	c.Assert(resp.Code, Equals, http.StatusOK)
}

func (t *RepoListTests) TestRepoListContentType(c *C) {
	common.SignWithPassphrase(t.req, adminSecret)
	resp := t.getResponse(t.req)

	contentType := resp.Header().Get("Content-Type")
	c.Assert(
		strings.HasPrefix(
			contentType,
			"application/json"),
		Equals,
		true)
}

func (t *RepoListTests) TestRepoListOutput(c *C) {
	common.SignWithPassphrase(t.req, adminSecret)
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
	common.SignWithPassphrase(t.req, adminSecret)
	resp := t.getResponse(t.req)
	c.Assert(resp.Code, Equals, 200)
	c.Assert(resp.Body.String(), Equals, "[]")
	c.Assert(resp.Body.Len(), Not(Equals), 0)
}

func (t *RepoListTests) TestRepoListMangled(c *C) {
	common.SignWithPassphrase(t.req, adminSecret)
	t.req.Header.Set("Mangled", "Yes")
	resp := t.getResponse(t.req)
	c.Assert(resp.Code, Equals, http.StatusUnauthorized)
}
