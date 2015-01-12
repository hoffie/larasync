package main

import (
	"bytes"

	. "gopkg.in/check.v1"
)

type MainTests struct {
	dir string
	err *bytes.Buffer
	out *bytes.Buffer
	in  *bytes.Buffer
	d   *Dispatcher
}

var _ = Suite(&MainTests{})

func (t *MainTests) SetUpTest(c *C) {
	t.dir = c.MkDir()
	t.err = new(bytes.Buffer)
	t.out = new(bytes.Buffer)
	t.in = new(bytes.Buffer)
	t.d = &Dispatcher{stderr: t.err, stdout: t.out, stdin: t.in}
}

func (t *MainTests) TestPrompt(c *C) {
	t.in.WriteString("test\n")
	res, err := t.d.prompt("Foo: ")
	c.Assert(t.out.Bytes(), DeepEquals, []byte("Foo: "))
	c.Assert(err, IsNil)
	c.Assert(res, DeepEquals, []byte("test"))
}
