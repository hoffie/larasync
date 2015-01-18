package api

import (
	. "gopkg.in/check.v1"
)

type RepositoriesClientTest struct {
	BaseClientTest
}

var _ = Suite(&RepositoriesClientTest{newBaseClientTest()})

func (t *RepositoriesClientTest) SetUpTest(c *C) {
	t.BaseClientTest.SetUpTest(c)
	t.client.SetAdminSecret(adminSecret)
}

func (t *RepositoriesClientTest) TestRegister(c *C) {
	err := t.client.Register(t.pubKey)
	c.Assert(err, IsNil)
}

func (t *RepositoriesClientTest) TestConnError(c *C) {
	t.server.Close()
	err := t.client.Register(t.pubKey)
	c.Assert(err, NotNil)
}

func (t *RepositoriesClientTest) TestAdminSecretError(c *C) {
	t.client.adminSecret = []byte{}
	err := t.client.Register(t.pubKey)
	c.Assert(err, NotNil)
}
