package repository

import (
	"io/ioutil"
	"path/filepath"

	. "gopkg.in/check.v1"
)

type ClientRepositoryTests struct {
	RepositoryTests
}

var _ = Suite(&ClientRepositoryTests{})

func (t *ClientRepositoryTests) TestStateConfig(c *C) {
	exp := "example.org:14124"

	r := NewClient(t.dir)
	err := r.CreateManagementDir()
	c.Assert(err, IsNil)

	sc, err := r.StateConfig()
	c.Assert(err, IsNil)
	sc.DefaultServer.URL = exp
	sc.Save()

	r2 := NewClient(t.dir)
	sc2, err := r2.StateConfig()
	c.Assert(err, IsNil)
	c.Assert(sc2.DefaultServer.URL, Equals, exp)
}

func (t *RepositoryTests) TestPathToNIBID(c *C) {
	r := NewClient(t.dir)
	err := r.CreateManagementDir()
	c.Assert(err, IsNil)

	err = r.keys.CreateHashingKey()
	c.Assert(err, IsNil)

	path := "foo/bar.txt"
	id, err := r.pathToNIBID(path)
	c.Assert(err, IsNil)
	c.Assert(id, Not(Equals), "")

	id2, err := r.pathToNIBID(path)
	c.Assert(err, IsNil)
	c.Assert(id2, Equals, id)
}

func (t *RepositoryTests) TestGetFileChunkIDs(c *C) {
	r := NewClient(t.dir)
	err := r.CreateManagementDir()
	c.Assert(err, IsNil)

	err = r.keys.CreateHashingKey()
	c.Assert(err, IsNil)

	path := filepath.Join(t.dir, "foo.txt")
	err = ioutil.WriteFile(path, []byte("test"), 0600)
	c.Assert(err, IsNil)

	ids, err := r.getFileChunkIDs(path)
	c.Assert(err, IsNil)
	c.Assert(len(ids), Equals, 1)
	c.Assert(len(ids[0]), Not(Equals), 0)

	ids2, err := r.getFileChunkIDs(path)
	c.Assert(err, IsNil)
	c.Assert(ids2, DeepEquals, ids)
}

func (t *RepositoryTests) TestCurrentAuthorization(c *C) {
	r := NewClient(t.dir)
	err := r.CreateManagementDir()
	c.Assert(err, IsNil)

	err = r.CreateKeys()
	c.Assert(err, IsNil)

	keyStore := r.keys
	auth, err := r.CurrentAuthorization()
	c.Assert(err, IsNil)

	encrpytionKey, err := keyStore.EncryptionKey()
	c.Assert(err, IsNil)

	hashingKey, err := keyStore.HashingKey()
	c.Assert(err, IsNil)

	signingKey, err := keyStore.SigningPrivateKey()
	c.Assert(err, IsNil)

	c.Assert(auth.EncryptionKey, DeepEquals, encrpytionKey)
	c.Assert(auth.HashingKey, DeepEquals, hashingKey)
	c.Assert(auth.SigningKey, DeepEquals, signingKey)

}

func (t *RepositoryTests) TestGetSigningKey(c *C) {
	r := NewClient(t.dir)
	err := r.CreateManagementDir()
	c.Assert(err, IsNil)

	err = r.keys.CreateSigningKey()
	c.Assert(err, IsNil)

	data, err := r.GetSigningPrivateKey()
	c.Assert(err, IsNil)

	keyData, err := r.keys.SigningPrivateKey()
	c.Assert(err, IsNil)

	c.Assert(data, DeepEquals, keyData)
}
