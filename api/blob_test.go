package api

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"time"

	edhelpers "github.com/hoffie/larasync/helpers/ed25519"
	. "gopkg.in/check.v1"

	"github.com/hoffie/larasync/repository"
)

type BlobTests struct {
	server         *Server
	rm             *repository.Manager
	req            *http.Request
	repos          string
	repositoryName string
	blobID         string
	blobData       []byte
	pubKey         [PubkeySize]byte
	privateKey     [PrivateKeySize]byte
	httpMethod     string
}

func (t *BlobTests) SetUpTest(c *C) {
	t.repos = c.MkDir()
	t.httpMethod = "GET"
	t.repositoryName = "test"
	t.blobID = "1234567890"
	t.blobData = []byte("This is testdata")
	rm, err := repository.NewManager(t.repos)
	c.Assert(err, IsNil)
	t.server = New(adminPubkey, time.Minute, rm)
	c.Assert(rm.Exists(t.repositoryName), Equals, false)
	t.rm = rm
	t.req = t.requestWithBytes(c, nil)
}

func (t *BlobTests) SetUpSuite(c *C) {
	byteArray := make([]byte, PrivateKeySize)
	_, err := rand.Read(byteArray)
	c.Assert(err, IsNil)
	t.privateKey, err = passphraseToKey(byteArray)
	c.Assert(err, IsNil)
	t.pubKey = edhelpers.GetPublicKeyFromPrivate(t.privateKey)
}

func (t *BlobGetTests) TearDownTest(c *C) {
	os.RemoveAll(t.repos)
}

func (t *BlobTests) getServer() *Server {
	return t.server
}

func (t *BlobTests) setServer(server *Server) {
	t.server = server
}

func (t *BlobTests) getResponse(req *http.Request) *httptest.ResponseRecorder {
	rw := httptest.NewRecorder()
	t.server.router.ServeHTTP(rw, req)
	return rw
}

func (t *BlobTests) requestWithBytes(c *C, body []byte) *http.Request {
	var httpBody io.Reader
	if body == nil {
		httpBody = nil
	} else {
		httpBody = bytes.NewReader(body)
	}
	return t.requestWithReader(c, httpBody)
}

func (t *BlobTests) requestWithReader(c *C, httpBody io.Reader) *http.Request {
	req, err := http.NewRequest(
		t.httpMethod,
		fmt.Sprintf(
			"http://example.org/repositories/%s/blobs/%s",
			t.repositoryName,
			t.blobID,
		),
		httpBody)
	c.Assert(err, IsNil)
	if httpBody != nil {
		req.Header.Set("Content-Type", "application/octet-type")
	}
	return req
}

func (t *BlobTests) createRepository(c *C) {
	err := t.rm.Create(t.repositoryName, t.pubKey[:])
	c.Assert(err, IsNil)
}

func (t *BlobTests) createBlob(c *C) {
	t.createBlobWithData(c, t.blobData)
}

func (t *BlobTests) createBlobWithData(c *C, data []byte) {
	if !t.rm.Exists(t.repositoryName) {
		t.createRepository(c)
	}

	repository, err := t.rm.Open(t.repositoryName)
	c.Assert(err, IsNil)
	err = repository.AddObject(t.blobID, bytes.NewReader(data))
	c.Assert(err, IsNil)
}

func (t *BlobTests) expectStoredBlobData(c *C) {
	t.expectStoredData(c, t.blobData)
}

func (t *BlobTests) expectStoredData(c *C, expectedData []byte) {
	repository, err := t.rm.Open(t.repositoryName)
	c.Assert(err, IsNil)
	reader, err := repository.GetObjectData(t.blobID)
	c.Assert(err, IsNil)
	retrievedData, err := ioutil.ReadAll(reader)
	c.Assert(err, IsNil)
	c.Assert(retrievedData, DeepEquals, expectedData)
}

func (t *BlobTests) signRequest() {
	SignWithKey(t.req, t.privateKey)
}
