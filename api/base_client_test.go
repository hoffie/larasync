package api

import (
	"crypto/rand"
	"path"
	"path/filepath"

	"github.com/hoffie/larasync/repository"

	. "gopkg.in/check.v1"
)

func newBaseClientTest() BaseClientTest {
	return BaseClientTest{
		BaseTests: BaseTests{},
	}
}

type BaseClientTest struct {
	BaseTests
	client        *Client
	server        *TestServer
	encryptionKey [repository.EncryptionKeySize]byte
	hashingKey    [repository.HashingKeySize]byte
}

func (t *BaseClientTest) serverURL(c *C) string {
	return "https://" + path.Join(t.server.hostAndPort, "repositories", t.repositoryName)
}

func (t *BaseClientTest) SetUpTest(c *C) {
	t.BaseTests.SetUpTest(c)
	var err error
	t.server, err = NewTestServer(t.certFile, t.keyFile, t.rm)
	c.Assert(err, IsNil)

	_, err = rand.Read(t.encryptionKey[:])
	c.Assert(err, IsNil)

	_, err = rand.Read(t.hashingKey[:])
	c.Assert(err, IsNil)

	t.client = NewClient(
		t.serverURL(c), "",
		func(string) bool { return true })
	t.client.SetSigningPrivateKey(t.privateKey)
}

func (t *BaseClientTest) TearDownTest(c *C) {
	t.server.Close()
}

func (t *BaseClientTest) repositoryPath(c *C) string {
	return filepath.Join(t.repos, t.repositoryName)
}

func (t *BaseClientTest) getClientRepository(c *C) *repository.ClientRepository {
	repo := repository.NewClient(t.repositoryPath(c))
	err := repo.SetKeysFromAuth(&repository.Authorization{
		SigningKey:    t.privateKey,
		EncryptionKey: t.encryptionKey,
		HashingKey:    t.hashingKey,
	})
	c.Assert(err, IsNil)
	return repo
}
