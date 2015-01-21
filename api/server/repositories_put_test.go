package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"github.com/hoffie/larasync/api"

	. "github.com/hoffie/larasync/api/common"

	. "gopkg.in/check.v1"
)

type RepoListCreateTests struct {
	BaseTests
	req    *http.Request
	pubKey []byte
}

var _ = Suite(&RepoListCreateTests{
	BaseTests: newBaseTest(),
})

func (t *RepoListCreateTests) requestWithBytes(c *C, body []byte) *http.Request {
	var httpBody io.Reader
	if body == nil {
		httpBody = nil
	} else {
		httpBody = bytes.NewReader(body)
	}
	return t.requestWithReader(c, httpBody)
}

func (t *RepoListCreateTests) requestWithReader(c *C, httpBody io.Reader) *http.Request {
	req, err := http.NewRequest(
		"PUT",
		fmt.Sprintf(
			"http://example.org/repositories/%s",
			t.repositoryName,
		),
		httpBody)
	c.Assert(err, IsNil)
	if httpBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req
}

func (t *RepoListCreateTests) SetUpTest(c *C) {
	t.repositoryName = "test"
	t.createRepoManager(c)
	t.createServer(c)
	t.req = t.requestWithBytes(c, nil)
}

func (t *RepoListCreateTests) SetUpSuite(c *C) {
	t.pubKey = make([]byte, PublicKeySize)
	t.createServerCert(c)
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
	repository, err := json.Marshal(api.JSONRepository{
		PubKey: t.pubKey,
	})
	c.Assert(err, IsNil)
	t.req = t.requestWithBytes(c, repository)
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

	contentType := resp.Header().Get("Content-Type")
	c.Assert(
		strings.HasPrefix(
			contentType,
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
	jsonBytes := bytes.NewBufferString("{'hello':'world'}").Bytes()
	t.req = t.requestWithBytes(c, jsonBytes)
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
