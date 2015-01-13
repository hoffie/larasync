package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"

	. "gopkg.in/check.v1"
)

type CheckoutTests struct {
	dir   string
	oldWd string
	out   *bytes.Buffer
	d     *Dispatcher
}

var _ = Suite(&CheckoutTests{})

func (t *CheckoutTests) SetUpTest(c *C) {
	t.dir = c.MkDir()
	wd, err := os.Getwd()
	c.Assert(err, IsNil)
	t.oldWd = wd
	err = os.Chdir(t.dir)
	c.Assert(err, IsNil)
	t.out = new(bytes.Buffer)
	t.d = &Dispatcher{stderr: t.out}
}

func (t *CheckoutTests) TearDownTest(c *C) {
	os.Chdir(t.oldWd)
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

func (t *CheckoutTests) TestAddAndCheckoutSubdir(c *C) {
	t.addAndCheckoutSubdir(c, "subdir", "subdir")
}

func (t *CheckoutTests) TestAddAndCheckoutSubdirNested(c *C) {
	t.addAndCheckoutSubdir(c, filepath.Join("subdir", "2", "3"), "subdir")
}

// addAndCheckoutSubdir adds a file in a subdir, removes the subdir,
// runs checkout on it and verifies that the file the subdir is created,
// and the file restored properly.
func (t *CheckoutTests) addAndCheckoutSubdir(c *C, subdir, subdirRoot string) {
	expContent := []byte("test123")
	repoDir := filepath.Join(t.dir, "repo")
	c.Assert(t.d.run([]string{"init", repoDir}), Equals, 0)

	subdir = filepath.Join(repoDir, subdir)
	subdirRoot = filepath.Join(repoDir, subdirRoot)
	err := os.MkdirAll(subdir, 0700)
	c.Assert(err, IsNil)

	path := filepath.Join(subdir, "foo")
	err = ioutil.WriteFile(path, expContent, 0600)
	c.Assert(err, IsNil)
	c.Assert(t.d.run([]string{"add", path}), Equals, 0)

	err = os.Remove(path)
	c.Assert(err, IsNil)

	err = os.RemoveAll(subdirRoot)
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

// TestCheckoutAll tests lara checkout without arguments.
func (t *CheckoutTests) TestCheckoutAll(c *C) {
	repoDir := "repo"
	c.Assert(t.d.run([]string{"init", repoDir}), Equals, 0)
	err := os.Chdir(repoDir)
	c.Assert(err, IsNil)

	expData := map[string][]byte{
		"foo1": []byte("content of file foo1"),
		"foo2": []byte("content of file foo2"),
	}
	for path, content := range expData {
		err := ioutil.WriteFile(path, content, 0600)
		c.Assert(err, IsNil)
		c.Assert(t.d.run([]string{"add", path}), Equals, 0)
	}

	c.Assert(t.d.run([]string{"checkout"}), Equals, 0)
	for path, expContent := range expData {
		content, err := ioutil.ReadFile(path)
		c.Assert(err, IsNil)
		c.Assert(content, DeepEquals, expContent)
	}
}
