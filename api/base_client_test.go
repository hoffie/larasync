package api

import (
	"path"

	. "gopkg.in/check.v1"
)

func newBaseClientTest() BaseClientTest {
	return BaseClientTest{
		BaseTests: BaseTests{},
	}
}

type BaseClientTest struct {
	BaseTests
	client *Client
	server *TestServer
}

func (t *BaseClientTest) serverURL(c *C) string {
	return "https://" + path.Join(t.server.hostAndPort, "repositories", t.repositoryName)
}

func (t *BaseClientTest) SetUpTest(c *C) {
	t.BaseTests.SetUpTest(c)
	var err error
	t.server, err = NewTestServer(t.certFile, t.keyFile)

	c.Assert(err, IsNil)
	t.client = NewClient(
		t.serverURL(c), "",
		func(string) bool { return true })
	t.client.SetSigningPrivateKey(t.privateKey)
}

func (t *BaseClientTest) TearDownTest(c *C) {
	t.server.Close()
}
