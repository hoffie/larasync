package server

import (
	"net/http"

	. "gopkg.in/check.v1"
)

type NIBGetTest struct {
	NIBItemTest
}

var _ = Suite(&NIBGetTest{getNIBItemTest()})

func (t *NIBGetTest) SetUpTest(c *C) {
	t.NIBItemTest.SetUpTest(c)
	t.addTestNIB(c)
}

func (t *NIBGetTest) TestNotFound(c *C) {
	t.nibID = "does-not-exist"
	req := t.requestEmptyBody(c)
	t.req = req
	t.signRequest()
	resp := t.getResponse(req)
	c.Assert(resp.Code, Equals, http.StatusNotFound)
}

func (t *NIBGetTest) TestUnauthorized(c *C) {
	resp := t.getResponse(t.req)
	c.Assert(resp.Code, Equals, http.StatusUnauthorized)
}

func (t *NIBGetTest) TestRepositoryNotExisting(c *C) {
	t.repositoryName = "does-not-exist"
	t.req = t.requestEmptyBody(c)
	t.signRequest()
	resp := t.getResponse(t.req)
	c.Assert(resp.Code, Equals, http.StatusUnauthorized)
}

func (t *NIBGetTest) TestGet(c *C) {
	t.signRequest()
	resp := t.getResponse(t.req)
	c.Assert(resp.Code, Equals, http.StatusOK)
}

func (t *NIBGetTest) TestSignatureResponse(c *C) {
	t.signRequest()
	resp := t.getResponse(t.req)

	c.Assert(
		t.verifyNIBSignature(c, resp),
		Equals,
		true,
	)
}

func (t *NIBGetTest) TestNibExtraction(c *C) {
	t.signRequest()
	resp := t.getResponse(t.req)

	c.Assert(
		t.extractNIB(c, resp).ID,
		Equals,
		t.nibID,
	)
}
