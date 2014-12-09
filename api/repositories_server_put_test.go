package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"

	. "gopkg.in/check.v1"

	"github.com/hoffie/larasync/repository"
)

type RepoListCreateTests struct {
	server         *Server
	rm             *repository.Manager
	req            *http.Request
	repos          string
	repositoryName string
}

var _ = Suite(&RepoListCreateTests{})

func (t *RepoListCreateTests) requestWithBody(c *C, body []byte) *http.Request {
	var http_body io.Reader
	if body == nil {
		http_body = nil
	} else {
		http_body = bytes.NewReader(body)
	}
	req, err := http.NewRequest(
		"PUT",
		fmt.Sprintf(
			"http://example.org/repositories/%s",
			t.repositoryName,
		),
		http_body)
	c.Assert(err, IsNil)
	if http_body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req
}

func (t *RepoListCreateTests) SetUpTest(c *C) {
	t.repos = c.MkDir()
	t.repositoryName = "test"
	rm, err := repository.NewManager(t.repos)
	c.Assert(err, IsNil)
	t.server = New(adminPubkey, time.Minute, rm)
	c.Assert(rm.Exists(t.repositoryName), Equals, false)
	t.rm = rm
	t.req = t.requestWithBody(c, nil)
}

func (t *RepoListCreateTests) TearDownTest(c *C) {
	os.RemoveAll(t.repos)
}

func (t *RepoListCreateTests) getResponse(req *http.Request) *httptest.ResponseRecorder {
	rw := httptest.NewRecorder()
	t.server.router.ServeHTTP(rw, req)
	return rw
}

func (t *RepoListCreateTests) addPubKey(c *C) {
	json_repository, err := json.Marshal(JsonRepository{
		pubKey: make([]byte, PubkeySize),
	})
	c.Assert(err, IsNil)
	t.req = t.requestWithBody(c, json_repository)
}

func (t *RepoListCreateTests) TestRepoCreateUnauthorized(c *C) {
	resp := t.getResponse(t.req)
	c.Assert(resp.Code, Equals, http.StatusUnauthorized)
}

func (t *RepoListCreateTests) TestRepoCreateAdmin(c *C) {
	t.addPubKey(c)
	SignWithPassphrase(t.req, adminSecret)
	resp := t.getResponse(t.req)
	c.Assert(resp.Code, Equals, http.StatusCreated)
}

func (t *RepoListCreateTests) TestRepoCreateContentType(c *C) {
	t.addPubKey(c)
	SignWithPassphrase(t.req, adminSecret)
	resp := t.getResponse(t.req)

	content_type := resp.Header().Get("Content-Type")
	c.Assert(
		strings.HasPrefix(
			content_type,
			"application/json"),
		Equals,
		true)
}

func (t *RepoListCreateTests) TestRepoCreateMangled(c *C) {
	SignWithPassphrase(t.req, adminSecret)
	t.req.Header.Set("Mangled", "Yes")
	resp := t.getResponse(t.req)
	c.Assert(resp.Code, Equals, http.StatusUnauthorized)
}

func (t *RepoListTests) TestRepositoryCreate(c *C) {
	SignWithPassphrase(t.req, adminSecret)
}
