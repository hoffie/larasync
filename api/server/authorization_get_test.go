package server

import (
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	. "github.com/hoffie/larasync/api/common"
	. "gopkg.in/check.v1"
)

type AuthorizationGetTests struct {
	AuthorizationTests
}

var _ = Suite(&AuthorizationGetTests{getAuthorizationTest()})

func (t *AuthorizationGetTests) TestRepositoryNotExists(c *C) {
	resp := t.getResponse(t.req)
	c.Assert(resp.Code, Equals, http.StatusUnauthorized)
}

func (t *AuthorizationGetTests) TestNotSigned(c *C) {
	t.createRepository(c)
	resp := t.getResponse(t.req)
	c.Assert(resp.Code, Equals, http.StatusUnauthorized)
}

func (t *AuthorizationGetTests) TestSignedWithRepositoryKey(c *C) {
	t.createRepository(c)
	t.signRequest()
	resp := t.getResponse(t.req)

	c.Assert(resp.Code, Equals, http.StatusUnauthorized)
}

func (t *AuthorizationGetTests) TestNotFound(c *C) {
	t.createRepository(c)
	t.signRequestWithAuthKey()
	resp := t.getResponse(t.req)

	c.Assert(resp.Code, Equals, http.StatusUnauthorized)
}

func (t *AuthorizationGetTests) setUpWithExist(c *C) {
	t.createRepository(c)
	auth := t.testAuthorization(c)
	t.addAuthorization(c, auth)
	t.signRequestWithAuthKey()
}

func (t *AuthorizationGetTests) TestGet(c *C) {
	t.setUpWithExist(c)

	resp := t.getResponse(t.req)
	c.Assert(resp.Code, Equals, http.StatusOK)
}

func (t *AuthorizationGetTests) TestGetMimeType(c *C) {
	t.setUpWithExist(c)

	resp := t.getResponse(t.req)
	c.Assert(resp.Header().Get("Content-Type"), Equals, "application/octet-stream")
}

func (t *AuthorizationGetTests) TestGetBody(c *C) {
	t.setUpWithExist(c)

	resp := t.getResponse(t.req)
	data, err := ioutil.ReadAll(resp.Body)
	c.Assert(err, IsNil)

	c.Assert(len(data) > 0, Equals, true)
}

func (t *AuthorizationGetTests) TestRemove(c *C) {
	t.setUpWithExist(c)

	t.getResponse(t.req)

	repo := t.getRepository(c)
	_, err := repo.GetAuthorizationReader(t.authPublicKey)
	c.Assert(os.IsNotExist(err), Equals, true)
}

func (t *AuthorizationGetTests) TestPublicKeyExtractionFailure(c *C) {
	t.setUpWithExist(c)
	urlString := t.req.URL.String()
	urlString = urlString[:len(urlString)-2]
	var err error
	t.req.URL, err = url.Parse(urlString)
	c.Assert(err, IsNil)

	resp := t.getResponse(t.req)
	c.Assert(resp.Code, Equals, http.StatusUnauthorized)
}

func (t *AuthorizationGetTests) signRequestWithAuthKey() {
	SignWithKey(t.req, t.authPrivateKey)
}
