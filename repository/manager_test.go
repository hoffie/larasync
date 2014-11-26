package repository

import (
	"os"
	"path/filepath"
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) {
	TestingT(t)
}

type CreationTests struct {
	dir string
}

type Tests struct {
	dir string
	m   *Manager
}

var _ = Suite(&CreationTests{})
var _ = Suite(&Tests{})

func (t *CreationTests) SetUpTest(c *C) {
	t.dir = c.MkDir()
}

func (t *CreationTests) TestNew(c *C) {
	m, err := NewManager(t.dir)
	c.Assert(m, NotNil)
	c.Assert(err, IsNil)
}

func (t *CreationTests) TestNewBadTarget(c *C) {
	m, err := NewManager(filepath.Join(t.dir, "foo/bar"))
	c.Assert(m, IsNil)
	c.Assert(err, NotNil)
}

func (t *Tests) SetUpTest(c *C) {
	t.dir = c.MkDir()
	m, err := NewManager(t.dir)
	c.Assert(err, IsNil)
	t.m = m
}

func (t *Tests) TestList(c *C) {
	e, err := t.m.ListNames()
	c.Assert(err, IsNil)
	c.Assert(e, DeepEquals, []string(nil))
}

func (t *Tests) TestCreate(c *C) {
	err := t.m.Create("test", "pubkey")
	c.Assert(err, IsNil)
	e, err := t.m.ListNames()
	c.Assert(err, IsNil)
	c.Assert(e, DeepEquals, []string{"test"})
}

func (t *Tests) TestOpen(c *C) {
	t.m.Create("test", "pubkey")
	r, err := t.m.Open("test")
	c.Assert(err, IsNil)
	c.Assert(r, FitsTypeOf, &Repository{})
	c.Assert(r.Name, Equals, "test")
}

func (t *Tests) TestOpenNonExisting(c *C) {
	r, err := t.m.Open("test")
	c.Assert(err, NotNil)
	c.Assert(r, IsNil)
}

func (t *Tests) TestOpenNonDir(c *C) {
	const name = "test"
	f, err := os.Create(filepath.Join(t.dir, name))
	c.Assert(err, IsNil)
	f.Close()
	r, err := t.m.Open(name)
	c.Assert(err, NotNil)
	c.Assert(r, IsNil)
}
