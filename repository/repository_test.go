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
