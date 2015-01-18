package api

import (
	"crypto/rand"
	"os"
	"path/filepath"

	. "gopkg.in/check.v1"

	edhelpers "github.com/hoffie/larasync/helpers/ed25519"
	"github.com/hoffie/larasync/helpers/x509"
	"github.com/hoffie/larasync/repository"
)

type BaseTests struct {
	rm             *repository.Manager
	repositoryName string
	repos          string
	certFile       string
	keyFile        string
	pubKey         [PublicKeySize]byte
	privateKey     [PrivateKeySize]byte
}

func (t *BaseTests) SetUpTest(c *C) {
	t.createRepoManager(c)

	t.repositoryName = "test"
	c.Assert(t.rm.Exists(t.repositoryName), Equals, false)
}

func (t *BaseTests) SetUpSuite(c *C) {
	byteArray := make([]byte, PrivateKeySize)
	_, err := rand.Read(byteArray)
	c.Assert(err, IsNil)
	t.privateKey, err = passphraseToKey(byteArray)
	c.Assert(err, IsNil)
	t.pubKey = edhelpers.GetPublicKeyFromPrivate(t.privateKey)
	t.createServerCert(c)
}

func (t *BaseTests) createRepoManager(c *C) {
	t.repos = c.MkDir()
	rm, err := repository.NewManager(t.repos)
	c.Assert(err, IsNil)
	t.rm = rm
}

func (t *BaseTests) createServerCert(c *C) {
	dir := c.MkDir()
	t.certFile = filepath.Join(dir, "server.crt")
	t.keyFile = filepath.Join(dir, "server.key")
	err := x509.GenerateServerCertFiles(t.certFile, t.keyFile)
	c.Assert(err, IsNil)
}

func (t *BaseTests) TearDownTest(c *C) {
	os.RemoveAll(t.repos)
}

func (t *BaseTests) createRepository(c *C) *repository.Repository {
	err := t.rm.Create(t.repositoryName, t.pubKey[:])
	if err != nil && !os.IsExist(err) {
		c.Assert(err, IsNil)
	}
	return t.getRepository(c)
}

func (t *BaseTests) getRepository(c *C) *repository.Repository {
	rep, err := t.rm.Open(t.repositoryName)
	c.Assert(err, IsNil)
	return rep
}
