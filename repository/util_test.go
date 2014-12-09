package repository

import (
	"os"
	"path/filepath"

	. "gopkg.in/check.v1"
)

type UtilTests struct {
	dir string
}

var _ = Suite(&UtilTests{})

func (t *UtilTests) SetUpTest(c *C) {
	t.dir = c.MkDir()
}

func (t *UtilTests) TestRepoRootInvalid(c *C) {
	r, err := GetRoot(t.dir)
	c.Assert(err, NotNil)
	c.Assert(r, Equals, "")
}

func (t *UtilTests) TestRepoRootValid(c *C) {
	path := filepath.Join(t.dir, ".lara")
	err := os.Mkdir(path, 0700)
	c.Assert(err, IsNil)
	r, err := GetRoot(t.dir)
	c.Assert(err, IsNil)
	c.Assert(r, Equals, t.dir)
}

func (t *UtilTests) TestRepoRootSubdirValid(c *C) {
	err := os.Mkdir(filepath.Join(t.dir, ".lara"), 0700)
	c.Assert(err, IsNil)
	err = os.Mkdir(filepath.Join(t.dir, "foo"), 0700)
	c.Assert(err, IsNil)
	r, err := GetRoot(filepath.Join(t.dir, "foo"))
	c.Assert(err, IsNil)
	c.Assert(r, Equals, t.dir)
}

func (t *UtilTests) TestRepoRootNoLeadingSlash(c *C) {
	_, err := GetRoot(filepath.Join("non/existing/dir"))
	c.Assert(err, NotNil)
}
