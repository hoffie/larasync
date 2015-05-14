package client

import (
	"crypto/rand"
	"os"
	"path"
	"path/filepath"

	"github.com/hoffie/larasync/api/common"
	edhelpers "github.com/hoffie/larasync/helpers/ed25519"
	"github.com/hoffie/larasync/helpers/x509"
	"github.com/hoffie/larasync/repository"

	. "gopkg.in/check.v1"
)

func newBaseTest() BaseTest {
	return BaseTest{}
}

type BaseTest struct {
	rm             *repository.Manager
	repositoryName string
	repos          string
	certFile       string
	keyFile        string
	pubKey         [PublicKeySize]byte
	privateKey     [PrivateKeySize]byte
	encryptionKey  [repository.EncryptionKeySize]byte
	hashingKey     [repository.HashingKeySize]byte
	client         *Client
	server         *TestServer
}

func (t *BaseTest) serverURL(c *C) string {
	return "https://" + path.Join(t.server.hostAndPort, "repositories", t.repositoryName)
}

func (t *BaseTest) SetUpTest(c *C) {
	t.createRepoManager(c)

	t.repositoryName = "test"
	c.Assert(t.rm.Exists(t.repositoryName), Equals, false)

	var err error
	t.server, err = NewTestServer(t.certFile, t.keyFile, t.rm)
	c.Assert(err, IsNil)

	_, err = rand.Read(t.encryptionKey[:])
	c.Assert(err, IsNil)

	_, err = rand.Read(t.hashingKey[:])
	c.Assert(err, IsNil)

	t.client = New(
		t.serverURL(c), "",
		func(string) bool { return true })
	t.client.SetSigningPrivateKey(t.privateKey)
}

func (t *BaseTest) SetUpSuite(c *C) {
	byteArray := make([]byte, PrivateKeySize)
	_, err := rand.Read(byteArray)
	c.Assert(err, IsNil)
	t.privateKey, err = common.PassphraseToKey(byteArray)
	c.Assert(err, IsNil)
	t.pubKey = edhelpers.GetPublicKeyFromPrivate(t.privateKey)
	t.createServerCert(c)
}

func (t *BaseTest) TearDownTest(c *C) {
	t.server.Close()
	os.RemoveAll(t.repos)
}

func (t *BaseTest) repositoryPath(c *C) string {
	return filepath.Join(t.repos, t.repositoryName)
}

func (t *BaseTest) getClientRepository(c *C) *repository.ClientRepository {
	repo, err := repository.NewClient(t.repositoryPath(c))
	c.Assert(err, IsNil)
	err = repo.SetKeysFromAuth(&repository.Authorization{
		SigningKey:    t.privateKey,
		EncryptionKey: t.encryptionKey,
		HashingKey:    t.hashingKey,
	})
	c.Assert(err, IsNil)
	_, err = os.Stat(filepath.Join(repo.GetManagementDir(), "nib_tracker.db"))
	if os.IsNotExist(err) {
		repo.InitializeNIBTracker()
	}
	return repo
}

func (t *BaseTest) createRepoManager(c *C) {
	t.repos = c.MkDir()
	rm, err := repository.NewManager(t.repos)
	c.Assert(err, IsNil)
	t.rm = rm
}

func (t *BaseTest) createServerCert(c *C) {
	dir := c.MkDir()
	t.certFile = filepath.Join(dir, "server.crt")
	t.keyFile = filepath.Join(dir, "server.key")
	err := x509.GenerateServerCertFiles(t.certFile, t.keyFile)
	c.Assert(err, IsNil)
}

func (t *BaseTest) createRepository(c *C) *repository.Repository {
	err := t.rm.Create(t.repositoryName, t.pubKey[:])
	if err != nil && !os.IsExist(err) {
		c.Assert(err, IsNil)
	}
	return t.getRepository(c)
}

func (t *BaseTest) getRepository(c *C) *repository.Repository {
	rep, err := t.rm.Open(t.repositoryName)
	c.Assert(err, IsNil)
	return rep
}
