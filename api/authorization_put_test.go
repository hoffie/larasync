package api

import (
	"net/http"
	"io"
	"bytes"

	. "gopkg.in/check.v1"
)

type AuthorizationPutTests struct {
	AuthorizationTests
}

var _ = Suite(&AuthorizationPutTests{getAuthorizationTest()})

func (t *AuthorizationPutTests) SetUpTest(c *C) {
	t.AuthorizationTests.SetUpTest(c)
	t.httpMethod = "PUT"
	t.req = t.requestEmptyBody(c)
}

func (t *AuthorizationPutTests) TestRepositoryNotExists(c *C) {
	resp := t.getResponse(t.req)
	c.Assert(resp.Code, Equals, http.StatusUnauthorized)
}

func (t *AuthorizationPutTests) TestNotSigned(c *C) {
	t.createRepository(c)
	resp := t.getResponse(t.req)
	c.Assert(resp.Code, Equals, http.StatusUnauthorized)
}

func (t *AuthorizationPutTests) setUpWithExist(c *C) {
	t.createRepository(c)
	auth := t.testAuthorization(c)
	t.addAuthorization(c, auth)
}

func (t *AuthorizationPutTests) TestPut(c *C) {
	t.setUpWithExist(c)
	repo := t.getRepository(c)
	reader, err := repo.GetAuthorizationReader(t.authPublicKey)
	c.Assert(err, IsNil)
	buff := &bytes.Buffer{}
	_, err = io.Copy(buff, reader)
	c.Assert(err, IsNil)
	reader.Close()
	
	t.req = t.requestWithReader(c, buff)
	t.signRequest()

	resp := t.getResponse(t.req)
	c.Assert(resp.Code, Equals, http.StatusCreated)
}
