package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	. "gopkg.in/check.v1"

	"github.com/hoffie/larasync/helpers/test"
)

type AuthorizeNewClientTest struct {
	BaseTests
}

var _ = Suite(&AuthorizeNewClientTest{BaseTests{}})

func (t *AuthorizeNewClientTest) doAuthorization(c *C) {
	res := t.d.run([]string{"authorize-new-client"})
	if res != 0 {
		c.Error(string(t.err.Bytes()))
	}
}

func (t *AuthorizeNewClientTest) TestAuthorization(c *C) {
	t.initRepo(c)
	t.registerServerInRepo(c)
	t.doAuthorization(c)

	num, err := test.NumFilesInDir(filepath.Join(t.serverRepoPath(), "authorizations"))
	c.Assert(err, IsNil)
	c.Assert(num, Equals, 1)
}

func (t *AuthorizeNewClientTest) TestAuthorizationNotInRepo(c *C) {
	c.Assert(t.d.run([]string{"authorize-new-client"}), Equals, 1)
}

func (t *AuthorizeNewClientTest) TestKeyMissing(c *C) {
	t.initRepo(c)
	t.registerServerInRepo(c)

	err := os.Remove(filepath.Join(".lara", "keys", "signing.priv"))
	c.Assert(err, IsNil)
	c.Assert(t.d.run([]string{"authorize-new-client"}), Equals, 1)
}

func (t *AuthorizeNewClientTest) TestOtherPrivKey(c *C) {
	t.initRepo(c)
	t.registerServerInRepo(c)

	data := [PrivateKeySize]byte{}
	err := ioutil.WriteFile(filepath.Join(".lara", "keys", "signing.priv"), data[:], 0600)
	c.Assert(err, IsNil)
	c.Assert(t.d.run([]string{"authorize-new-client"}), Equals, 1)
}
