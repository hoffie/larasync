package server

import (
	. "gopkg.in/check.v1"
)

type ServerTests struct {
	BaseTests
}

var _ = Suite(
	&ServerTests{
		newBaseTest(),
	},
)

func (t *ServerTests) SetUpTest(c *C) {
	t.BaseTests.SetUpTest(c)
	t.getURL = func() string {
		return "http://example.org/"
	}
	t.req = t.requestEmptyBody(c)
}

func (t *ServerTests) TestRootGet(c *C) {
	resp := t.getResponse(t.req)
	c.Assert(resp.Code, Equals, 200)
}
