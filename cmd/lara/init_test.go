package main

import (
	"os"
	"path/filepath"
	"testing"
	"bytes"

	. "gopkg.in/check.v1"
)

func TestInit(t *testing.T) {
	TestingT(t)
}

type InitTests struct {
	dir string
	out *bytes.Buffer
}

var _ = Suite(&InitTests{})

func (t *InitTests) SetUpTest(c *C) {
	t.dir = c.MkDir()
	t.out = new(bytes.Buffer)
}

func (t *InitTests) TestCwd(c *C) {
	oldpwd, err := os.Getwd()
	c.Assert(err, IsNil)
	os.Chdir(t.dir)
	defer os.Chdir(oldpwd)
	c.Assert(t.out.String(), Equals, "")
	c.Assert(dispatch(t.out, []string{"init"}), Equals, 0)
	s, err := os.Stat(filepath.Join(t.dir, ".lara"))
	c.Assert(err, IsNil)
	c.Assert(s.IsDir(), Equals, true)
}

func (t *InitTests) TestOtherDir(c *C) {
	path := filepath.Join(t.dir, "foo")
	c.Assert(dispatch(t.out, []string{"init", path}), Equals, 0)
	s, err := os.Stat(filepath.Join(path, ".lara"))
	c.Assert(err, IsNil)
	c.Assert(s.IsDir(), Equals, true)
}

func (t *InitTests) TestOtherDirExisting(c *C) {
	path := filepath.Join(t.dir, "foo")
	err := os.Mkdir(path, 0700)
	c.Assert(err, IsNil)
	c.Assert(dispatch(t.out, []string{"init", path}), Equals, 1)
	_, err = os.Stat(filepath.Join(path, ".lara"))
	c.Assert(err, Not(IsNil))
}
