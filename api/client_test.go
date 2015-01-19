package api

import (
	. "gopkg.in/check.v1"
)

type ClientTests struct {
}

var _ = Suite(&ClientTests{})

func (t *ClientTests) TestClientCreation(c *C) {
	client := NewClient("http://localhost:6543/", "", func(_ string) bool {
		return true
	})
	c.Assert(client, NotNil)
}

func (t *ClientTests) TestNetlocToUrl(c *C) {
	testRepository := "https://example.org:80/repositories/testRepository"
	c.Assert(NetlocToURL("example.org:80", "testRepository"), Equals, testRepository)
}
