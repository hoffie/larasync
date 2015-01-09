package repository

import (
	"os"
	"path/filepath"

	. "gopkg.in/check.v1"
)

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

func (t *CreationTests) TestNewOnFile(c *C) {
	const name = "test"
	f, err := os.Create(filepath.Join(t.dir, name))
	c.Assert(err, IsNil)
	f.Close()
	m, err := NewManager(filepath.Join(t.dir, name))
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
	c.Assert(e, DeepEquals, []string{})
}

func (t *Tests) TestListExcludeFile(c *C) {
	const name = "test"
	f, err := os.Create(filepath.Join(t.dir, name))
	c.Assert(err, IsNil)
	f.Close()
	e, err := t.m.ListNames()
	c.Assert(err, IsNil)
	c.Assert(e, DeepEquals, []string{})
}

func (t *Tests) TestListBadBasePath(c *C) {
	// fake error condition for testing
	t.m.basePath = "/dev/null/asdf"
	e, err := t.m.ListNames()
	c.Assert(err, NotNil)
	c.Assert(e, IsNil)
}

func (t *Tests) TestCreate(c *C) {
	err := t.m.Create("test", []byte("pubkey"))
	c.Assert(err, IsNil)
	e, err := t.m.ListNames()
	c.Assert(err, IsNil)
	c.Assert(e, DeepEquals, []string{"test"})
}

func (t *Tests) TestOpen(c *C) {
	t.m.Create("test", []byte("pubkey"))
	r, err := t.m.Open("test")
	c.Assert(err, IsNil)
	c.Assert(r, FitsTypeOf, &Repository{})
	c.Assert(r.Path, Equals, filepath.Join(t.dir, "test"))
}

func (t *Tests) TestPubkey(c *C) {
	expKey := []byte("01234567890123456789012345678901")
	var arrExpKey [PublicKeySize]byte
	copy(arrExpKey[:], expKey[:PublicKeySize])

	t.m.Create("test", expKey)
	r, err := t.m.Open("test")
	c.Assert(err, IsNil)
	c.Assert(r, FitsTypeOf, &Repository{})
	key, err := r.keys.SigningPublicKey()
	c.Assert(err, IsNil)
	c.Assert(key, DeepEquals, arrExpKey)
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

func (t *Tests) TestExists(c *C) {
	const name = "test"
	err := t.m.Create("test", []byte("pubkey"))
	c.Assert(err, IsNil)
	c.Assert(t.m.Exists(name), Equals, true)
}

func (t *Tests) TestDoesNotExist(c *C) {
	const name = "test"
	c.Assert(t.m.Exists(name), Equals, false)
}
