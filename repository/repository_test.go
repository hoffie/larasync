package repository

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"

	. "gopkg.in/check.v1"
)

type RepositoryTests struct {
	dir string
}

var _ = Suite(&RepositoryTests{})

func (t *RepositoryTests) SetUpTest(c *C) {
	t.dir = c.MkDir()
}

func (t *RepositoryTests) TestGetEncryptionKey(c *C) {
	r := New(filepath.Join(t.dir, "foo"))
	var k [EncryptionKeySize]byte
	k[0] = 'z'
	_, err := r.GetEncryptionKey()
	c.Assert(err, NotNil)

	err = r.SetEncryptionKey(k)
	c.Assert(err, NotNil)

	err = r.Create()
	c.Assert(err, IsNil)

	err = r.SetEncryptionKey(k)
	c.Assert(err, IsNil)

	k2, err := r.GetEncryptionKey()
	c.Assert(err, IsNil)
	c.Assert(k2, DeepEquals, k)
}

func (t *RepositoryTests) TestGetSigningPrivkey(c *C) {
	r := New(filepath.Join(t.dir, "foo"))
	var k [PrivateKeySize]byte
	k[0] = 'z'
	_, err := r.GetSigningPrivkey()
	c.Assert(err, NotNil)

	err = r.SetSigningPrivkey(k)
	c.Assert(err, NotNil)

	err = r.Create()
	c.Assert(err, IsNil)

	err = r.SetSigningPrivkey(k)
	c.Assert(err, IsNil)

	k2, err := r.GetSigningPrivkey()
	c.Assert(err, IsNil)
	c.Assert(k2, DeepEquals, k)
}

func (t *RepositoryTests) TestGetRepoRelativePath(c *C) {
	r := New(filepath.Join(t.dir, "foo"))
	err := r.Create()
	c.Assert(err, IsNil)
	in := filepath.Join(t.dir, "foo", "test", "bar")
	out, err := r.getRepoRelativePath(in)
	c.Assert(err, IsNil)
	c.Assert(out, Equals, filepath.Join("test", "bar"))
}

func (t *RepositoryTests) TestGetRepoRelativePathFail(c *C) {
	r := New(filepath.Join(t.dir, "foo"))
	err := r.Create()
	c.Assert(err, IsNil)
	in := t.dir
	out, err := r.getRepoRelativePath(in)
	c.Assert(err, NotNil)
	c.Assert(out, Equals, "")
}

func (t *RepositoryTests) TestCreateManagementDir(c *C) {
	r := New(t.dir)
	err := r.CreateManagementDir()
	c.Assert(err, IsNil)

	s, err := os.Stat(filepath.Join(t.dir, ".lara"))
	c.Assert(err, IsNil)
	c.Assert(s.IsDir(), Equals, true)

	s, err = os.Stat(filepath.Join(t.dir, ".lara", "objects"))
	c.Assert(err, IsNil)
	c.Assert(s.IsDir(), Equals, true)

	s, err = os.Stat(filepath.Join(t.dir, ".lara", "nibs"))
	c.Assert(err, IsNil)
	c.Assert(s.IsDir(), Equals, true)

}

func (t *RepositoryTests) TestCreateEncryptionKey(c *C) {
	r := New(t.dir)
	err := r.CreateManagementDir()
	c.Assert(err, IsNil)

	err = r.CreateEncryptionKey()
	c.Assert(err, IsNil)

	key, err := r.GetEncryptionKey()
	c.Assert(err, IsNil)
	c.Assert(len(key), Equals, EncryptionKeySize)
}

func (t *RepositoryTests) TestCreateSigningKey(c *C) {
	r := New(t.dir)
	err := r.CreateManagementDir()
	c.Assert(err, IsNil)

	err = r.CreateSigningKey()
	c.Assert(err, IsNil)

	key, err := r.GetSigningPrivkey()
	c.Assert(err, IsNil)
	c.Assert(len(key), Equals, PrivateKeySize)
}

func (t *RepositoryTests) TestAddObject(c *C) {
	r := New(t.dir)
	err := r.CreateManagementDir()
	c.Assert(err, IsNil)
	objectID := "1234567890"
	objectReader := bytes.NewReader([]byte("Test data"))

	err = r.AddObject(objectID, objectReader)
	c.Assert(err, IsNil)
}

func (t *RepositoryTests) TestGetObject(c *C) {
	r := New(t.dir)
	err := r.CreateManagementDir()
	c.Assert(err, IsNil)
	objectID := "1234567890"
	objectData := []byte("Test data")
	objectReader := bytes.NewReader(objectData)

	r.AddObject(objectID, objectReader)

	reader, err := r.GetObjectData(objectId)
	c.Assert(err, IsNil)

	data, err := ioutil.ReadAll(reader)

	c.Assert(objectData, DeepEquals, data)
}
