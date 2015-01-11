package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"

	. "gopkg.in/check.v1"
)

type CheckoutTests struct {
	dir string
	out *bytes.Buffer
	d   *Dispatcher
}

var _ = Suite(&CheckoutTests{})

func (t *CheckoutTests) SetUpTest(c *C) {
	t.dir = c.MkDir()
	t.out = new(bytes.Buffer)
	t.d = &Dispatcher{stderr: t.out}
}

// TestAddAndCheckout adds a file, removes the real file, runs checkout
// on it and verifies that it looks like the original file.
func (t *CheckoutTests) TestAddAndCheckout(c *C) {
	expContent := []byte("test123")
	repoDir := filepath.Join(t.dir, "repo")
	c.Assert(t.d.run([]string{"init", repoDir}), Equals, 0)

	path := filepath.Join(repoDir, "foo")
	err := ioutil.WriteFile(path, expContent, 0600)
	c.Assert(err, IsNil)
	c.Assert(t.d.run([]string{"add", path}), Equals, 0)

	err = os.Remove(path)
	c.Assert(err, IsNil)
	_, err = ioutil.ReadFile(path)
	c.Assert(err, NotNil)

	c.Assert(t.d.run([]string{"checkout", path}), Equals, 0)
	content, err := ioutil.ReadFile(path)
	c.Assert(err, IsNil)
	c.Assert(content, DeepEquals, expContent)
}

// TestAddAndCheckoutSkipRev adds two revisions of a file, changes the content
// back to the first revision and checks if checking out still works.
func (t *CheckoutTests) TestAddAndCheckoutSkipRev(c *C) {
	content1 := []byte("rev1")
	content2 := []byte("rev2")

	repoDir := filepath.Join(t.dir, "repo")
	c.Assert(t.d.run([]string{"init", repoDir}), Equals, 0)

	path := filepath.Join(repoDir, "foo")
	err := ioutil.WriteFile(path, content1, 0600)
	c.Assert(err, IsNil)
	c.Assert(t.d.run([]string{"add", path}), Equals, 0)

	err = ioutil.WriteFile(path, content2, 0600)
	c.Assert(err, IsNil)
	c.Assert(t.d.run([]string{"add", path}), Equals, 0)

	err = ioutil.WriteFile(path, content1, 0600)
	c.Assert(err, IsNil)

	c.Assert(t.d.run([]string{"checkout", path}), Equals, 0)
	content, err := ioutil.ReadFile(path)
	c.Assert(err, IsNil)
	c.Assert(content, DeepEquals, content2)
}

// TestAddAndCheckoutChangedFile adds a file, modifies the working dir file,
// attempts to checkout the file, expecting it to be prevented.
func (t *CheckoutTests) TestAddAndCheckoutChangedFile(c *C) {
	repoDir := filepath.Join(t.dir, "repo")
	c.Assert(t.d.run([]string{"init", repoDir}), Equals, 0)

	path := filepath.Join(repoDir, "foo")
	err := ioutil.WriteFile(path, []byte("repo content"), 0600)
	c.Assert(err, IsNil)
	c.Assert(t.d.run([]string{"add", path}), Equals, 0)

	expContent := []byte("workdir-only content")
	err = ioutil.WriteFile(path, expContent, 0600)
	c.Assert(err, IsNil)

	c.Assert(t.d.run([]string{"checkout", path}), Equals, 1)
	content, err := ioutil.ReadFile(path)
	c.Assert(err, IsNil)
	c.Assert(content, DeepEquals, expContent)
}

// TestAddAndCheckoutNoChange adds a file and attempts to checkout the file,
// expecting it to work as it is unchanged.
func (t *CheckoutTests) TestAddAndCheckoutNoChange(c *C) {
	repoDir := filepath.Join(t.dir, "repo")
	c.Assert(t.d.run([]string{"init", repoDir}), Equals, 0)

	expContent := []byte("repo content")
	path := filepath.Join(repoDir, "foo")
	err := ioutil.WriteFile(path, expContent, 0600)
	c.Assert(err, IsNil)
	c.Assert(t.d.run([]string{"add", path}), Equals, 0)

	err = ioutil.WriteFile(path, expContent, 0600)
	c.Assert(err, IsNil)

	c.Assert(t.d.run([]string{"checkout", path}), Equals, 0)
	content, err := ioutil.ReadFile(path)
	c.Assert(err, IsNil)
	c.Assert(content, DeepEquals, expContent)
}
