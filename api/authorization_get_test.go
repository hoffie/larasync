package api

import (
	"io/ioutil"
	"net/http"

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

func (t *AuthorizationGetTests) signRequestWithAuthKey() {
	SignWithKey(t.req, t.authPrivateKey)
}
