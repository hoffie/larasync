package api

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"time"

	. "gopkg.in/check.v1"

	edhelpers "github.com/hoffie/larasync/helpers/ed25519"
	"github.com/hoffie/larasync/repository"
)

type BaseTests struct {
	server         *Server
	rm             *repository.Manager
	req            *http.Request
	repos          string
	repositoryName string
	pubKey         [PublicKeySize]byte
	privateKey     [PrivateKeySize]byte
	httpMethod     string
	getURL         func() string
	urlParams      url.Values
}

func (t *BaseTests) SetUpTest(c *C) {
	t.repos = c.MkDir()
	t.httpMethod = "GET"
	t.repositoryName = "test"

	rm, err := repository.NewManager(t.repos)
	c.Assert(err, IsNil)
	t.server = New(adminPubkey, time.Minute, rm)
	c.Assert(rm.Exists(t.repositoryName), Equals, false)
	t.rm = rm
	t.getURL = func() string {
		return fmt.Sprintf(
			"http://example.org/repositories/%s",
			t.repositoryName,
		)
	}
	t.req = t.requestWithBytes(c, nil)
	t.urlParams = url.Values{}
}

func (t *BaseTests) SetUpSuite(c *C) {
	byteArray := make([]byte, PrivateKeySize)
	_, err := rand.Read(byteArray)
	c.Assert(err, IsNil)
	t.privateKey, err = passphraseToKey(byteArray)
	c.Assert(err, IsNil)
	t.pubKey = edhelpers.GetPublicKeyFromPrivate(t.privateKey)
}

func (t *BaseTests) TearDownTest(c *C) {
	os.RemoveAll(t.repos)
}

func (t *BaseTests) getServer() *Server {
	return t.server
}

func (t *BaseTests) setServer(server *Server) {
	t.server = server
}

func (t *BaseTests) getResponse(req *http.Request) *httptest.ResponseRecorder {
	rw := httptest.NewRecorder()
	t.server.router.ServeHTTP(rw, req)
	return rw
}

func (t *BaseTests) requestEmptyBody(c *C) *http.Request {
	return t.requestWithBytes(c, nil)
}

func (t *BaseTests) requestWithBytes(c *C, body []byte) *http.Request {
	var httpBody io.Reader
	if body == nil {
		httpBody = nil
	} else {
		httpBody = bytes.NewReader(body)
	}
	return t.requestWithReader(c, httpBody)
}

func (t *BaseTests) requestWithReader(c *C, httpBody io.Reader) *http.Request {
	requestURL, err := url.Parse(t.getURL())
	c.Assert(err, IsNil)
	requestURL.RawQuery = t.urlParams.Encode()
	req, err := http.NewRequest(
		t.httpMethod,
		requestURL.String(),
		httpBody)
	c.Assert(err, IsNil)
	if httpBody != nil {
		req.Header.Set("Content-Type", "application/octet-type")
	}
	return req
}

func (t *BaseTests) createRepository(c *C) *repository.Repository {
	err := t.rm.Create(t.repositoryName, t.pubKey[:])
	if err != nil && !os.IsExist(err) {
		c.Assert(err, IsNil)
	}
	return t.getRepository(c)
}

func (t *BaseTests) getRepository(c *C) *repository.Repository {
	rep, err := t.rm.Open(t.repositoryName)
	c.Assert(err, IsNil)
	return rep
}

func (t *BaseTests) signRequest() {
	SignWithKey(t.req, t.privateKey)
}
