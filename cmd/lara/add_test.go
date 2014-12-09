package main

import (
	"bytes"
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
