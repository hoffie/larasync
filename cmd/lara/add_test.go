package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	. "gopkg.in/check.v1"
)

type AddTests struct {
	dir string
	out *bytes.Buffer
	d   *Dispatcher
}

var _ = Suite(&AddTests{})

func (t *AddTests) SetUpTest(c *C) {
	t.dir = c.MkDir()
	t.out = new(bytes.Buffer)
	t.d = &Dispatcher{stderr: t.out}
}

func (t *AddTests) TestAddNoArgs(c *C) {
	c.Assert(t.d.run([]string{"add"}), Equals, 1)
}

func (t *AddTests) TestAddNotPartOfRepo(c *C) {
	path := filepath.Join(t.dir, "foo")
	fh, err := os.Create(path)
	c.Assert(err, IsNil)
	fh.Close()
	c.Assert(t.d.run([]string{"add", path}), Equals, 1)
}

func (t *AddTests) TestAdd(c *C) {
	repoDir := filepath.Join(t.dir, "repo")
	c.Assert(t.d.run([]string{"init", repoDir}), Equals, 0)
	file := filepath.Join(repoDir, "foo")
	fh, err := os.Create(file)
	c.Assert(err, IsNil)
	fh.Close()
	c.Assert(t.d.run([]string{"add", file}), Equals, 0)
}

func (t *AddTests) TestAddDir(c *C) {
	repoDir := filepath.Join(t.dir, "repo")
	c.Assert(t.d.run([]string{"init", repoDir}), Equals, 0)
	dir := filepath.Join(repoDir, "subdir")
	err := os.Mkdir(dir, 0700)
	file := filepath.Join(dir, "foo.txt")
	realContent := []byte("test")
	err = ioutil.WriteFile(file, realContent, 0600)
	c.Assert(err, IsNil)
	c.Assert(t.d.run([]string{"add", dir}), Equals, 0)
	err = os.Remove(file)
	c.Assert(err, IsNil)
	c.Assert(t.d.run([]string{"checkout", file}), Equals, 0)

	content, err := ioutil.ReadFile(file)
	c.Assert(err, IsNil)
	c.Assert(content, DeepEquals, realContent)
}

func (t *AddTests) TestAddMultipleTimes(c *C) {
	repoDir := filepath.Join(t.dir, "repo")
	c.Assert(t.d.run([]string{"init", repoDir}), Equals, 0)
	dir := filepath.Join(repoDir, "subdir")
	err := os.Mkdir(dir, 0700)

	for i := 0; i < 10; i++ {
		file := filepath.Join(dir, fmt.Sprintf("foo%d.txt", i))
		realContent := []byte("test")
		err = ioutil.WriteFile(file, realContent, 0600)
		c.Assert(err, IsNil)

		c.Assert(t.d.run([]string{"add", repoDir}), Equals, 0)
	}
}
