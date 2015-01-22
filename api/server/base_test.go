package server

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/hoffie/larasync/api/common"
	edhelpers "github.com/hoffie/larasync/helpers/ed25519"
	"github.com/hoffie/larasync/helpers/x509"
	"github.com/hoffie/larasync/repository"

	. "gopkg.in/check.v1"
)

func newBaseTest() BaseTests {
	return BaseTests{}
}

type BaseTests struct {
	rm             *repository.Manager
	repositoryName string
	repos          string
	certFile       string
	keyFile        string
	pubKey         [PublicKeySize]byte
	privateKey     [PrivateKeySize]byte
	server         *Server
	req            *http.Request
	httpMethod     string
	getURL         func() string
	urlParams      url.Values
}

func (t *BaseTests) SetUpTest(c *C) {
	t.createRepoManager(c)

	t.repositoryName = "test"
	c.Assert(t.rm.Exists(t.repositoryName), Equals, false)
	t.createServer(c)
	t.httpMethod = "GET"
	t.getURL = func() string {
		return fmt.Sprintf(
			"http://example.org/repositories/%s",
			t.repositoryName,
		)
	}
	t.req = t.requestEmptyBody(c)
	t.urlParams = url.Values{}
}

func (t *BaseTests) SetUpSuite(c *C) {
	byteArray := make([]byte, PrivateKeySize)
	_, err := rand.Read(byteArray)
	c.Assert(err, IsNil)
	t.privateKey, err = common.PassphraseToKey(byteArray)
	c.Assert(err, IsNil)
	t.pubKey = edhelpers.GetPublicKeyFromPrivate(t.privateKey)
	t.createServerCert(c)
}

func (t *BaseTests) createServer(c *C) {
	var err error
	t.server, err = New(adminPubkey, time.Minute, t.rm, t.certFile, t.keyFile)
	c.Assert(err, IsNil)
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

func (t *BaseTests) signRequest() {
	common.SignWithKey(t.req, t.privateKey)
}

func (t *BaseTests) createRepoManager(c *C) {
	t.repos = c.MkDir()
	rm, err := repository.NewManager(t.repos)
	c.Assert(err, IsNil)
	t.rm = rm
}

func (t *BaseTests) createServerCert(c *C) {
	dir := c.MkDir()
	t.certFile = filepath.Join(dir, "server.crt")
	t.keyFile = filepath.Join(dir, "server.key")
	err := x509.GenerateServerCertFiles(t.certFile, t.keyFile)
	c.Assert(err, IsNil)
}

func (t *BaseTests) TearDownTest(c *C) {
	os.RemoveAll(t.repos)
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
