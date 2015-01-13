package main

import (
	"bytes"
	"os"

	. "gopkg.in/check.v1"
)

type BaseTests struct {
	dir   string
	oldWd string
	err   *bytes.Buffer
	out   *bytes.Buffer
	in    *bytes.Buffer
	d     *Dispatcher
	ts    *TestServer
}

func (t *BaseTests) SetUpTest(c *C) {
	t.dir = c.MkDir()
	wd, err := os.Getwd()
	c.Assert(err, IsNil)
	t.oldWd = wd

	err = os.Chdir(t.dir)
	c.Assert(err, IsNil)

	t.err = new(bytes.Buffer)
	t.out = new(bytes.Buffer)
	t.in = new(bytes.Buffer)
	t.d = &Dispatcher{stderr: t.err, stdout: t.out, stdin: t.in}

	ts, err := NewTestServer()
	c.Assert(err, IsNil)
	t.ts = ts
}

func (t *BaseTests) TearDownTest(c *C) {
	t.ts.Close()
	os.Chdir(t.oldWd)
}
