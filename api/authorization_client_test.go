package api

import (
	"bytes"
	"encoding/hex"
	"io"
	"io/ioutil"
	"path"

	"github.com/hoffie/larasync/repository"

	. "gopkg.in/check.v1"
)

type AuthorizationClientTest struct {
	BaseClientTest
	data          []byte
	authorization *repository.Authorization
}

var _ = Suite(&AuthorizationClientTest{
	BaseClientTest: newBaseClientTest(),
})

func (t *AuthorizationClientTest) getAuthorizationURL(c *C) string {
	return t.serverURL(c) + "/" + path.Join("authorizations", t.pubKeyToString())
}

func (t *AuthorizationClientTest) SetUpTest(c *C) {
	t.BaseClientTest.SetUpTest(c)
	t.data = []byte("This is test authorization data.")
	t.authorization = &repository.Authorization{
		SigningKey:    t.privateKey,
		EncryptionKey: t.encryptionKey,
		HashingKey:    t.hashingKey,
	}

	t.createRepository(c)
}

func (t *AuthorizationClientTest) pubKeyToString() string {
	return hex.EncodeToString(t.pubKey[:])
}

func (t *AuthorizationClientTest) doAuthorization(c *C) (io.Reader, error) {
	return t.client.GetAuthorization(t.getAuthorizationURL(c), t.privateKey)
}

func (t *AuthorizationClientTest) TestGet(c *C) {
	repository := t.getClientRepository(c)
	repository.SetAuthorization(
		t.pubKey,
		t.encryptionKey,
		t.authorization,
	)

	reader, err := t.doAuthorization(c)
	c.Assert(err, IsNil)
	d, err := ioutil.ReadAll(reader)
	c.Assert(err, IsNil)
	c.Assert(len(d) > 0, Equals, true)
}

func (t *AuthorizationClientTest) TestConnError(c *C) {
	t.server.Close()
	_, err := t.doAuthorization(c)
	c.Assert(err, NotNil)
}

func (t *AuthorizationClientTest) putAuthorization(c *C) error {
	return t.client.PutAuthorization(&t.pubKey, bytes.NewReader(t.data))
}

func (t *AuthorizationClientTest) TestAdd(c *C) {
	err := t.putAuthorization(c)
	c.Assert(err, IsNil)
}

func (t *AuthorizationClientTest) TestAddConnError(c *C) {
	t.server.Close()
	err := t.putAuthorization(c)
	c.Assert(err, NotNil)
}
