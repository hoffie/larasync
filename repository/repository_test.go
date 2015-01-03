package repository

import (
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
	k := make([]byte, 32)
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
