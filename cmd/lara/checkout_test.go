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
