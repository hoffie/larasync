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
	pubKey         []byte
}

var _ = Suite(&RepoListCreateTests{})

func (t *RepoListCreateTests) requestWithBytes(c *C, body []byte) *http.Request {
	var http_body io.Reader
	if body == nil {
		http_body = nil
	} else {
		http_body = bytes.NewReader(body)
	}
	return t.requestWithReader(c, http_body)
}

func (t *RepoListCreateTests) requestWithReader(c *C, http_body io.Reader) *http.Request {
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
	t.req = t.requestWithBytes(c, nil)
}

func (t *RepoListCreateTests) SetUpSuite(c *C) {
	t.pubKey = make([]byte, PubkeySize)
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
		PubKey: t.pubKey,
	})
	c.Assert(err, IsNil)
	t.req = t.requestWithBytes(c, json_repository)
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

func (t *RepoListCreateTests) TestRepositoryCreate(c *C) {
	t.addPubKey(c)
	SignWithPassphrase(t.req, adminSecret)
	t.getResponse(t.req)
	c.Assert(
		t.rm.Exists(t.repositoryName),
		Equals,
		true,
	)
}

func (t *RepoListCreateTests) TestWrongPubKeySize(c *C) {
	t.pubKey = make([]byte, 5)
	t.addPubKey(c)
	SignWithPassphrase(t.req, adminSecret)
	c.Assert(
		t.getResponse(t.req).Code,
		Equals,
		http.StatusBadRequest,
	)
}

func (t *RepoListCreateTests) TestRepoAlreadyExists(c *C) {
	t.rm.Create(t.repositoryName, t.pubKey)
	t.addPubKey(c)
	SignWithPassphrase(t.req, adminSecret)
	c.Assert(
		t.getResponse(t.req).Code,
		Equals,
		http.StatusConflict,
	)
}

func (t *RepoListCreateTests) TestMangledJson(c *C) {
	json_bytes := bytes.NewBufferString("{'hello':'world'}").Bytes()
	t.req = t.requestWithBytes(c, json_bytes)
	SignWithPassphrase(t.req, adminSecret)
	c.Assert(
		t.getResponse(t.req).Code,
		Equals,
		http.StatusBadRequest,
	)
}

func (t *RepoListCreateTests) TestRepositoryError(c *C) {
	os.RemoveAll(t.repos)
	t.addPubKey(c)
	SignWithPassphrase(t.req, adminSecret)
	c.Assert(
		t.getResponse(t.req).Code,
		Equals,
		http.StatusInternalServerError,
	)
}
