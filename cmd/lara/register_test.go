package main

import (
	"bytes"

	. "gopkg.in/check.v1"
)

type RegisterTests struct {
	dir string
	out *bytes.Buffer
	d   *Dispatcher
}

var _ = Suite(&RegisterTests{})

func (t *RegisterTests) SetUpTest(c *C) {
	t.dir = c.MkDir()
	t.out = new(bytes.Buffer)
	t.d = &Dispatcher{stderr: t.out}
}

func (t *RegisterTests) TestRegisterNoArgs(c *C) {
	c.Assert(t.d.run([]string{"register"}), Equals, 1)
}

func (t *RegisterTests) TestRegisterOnlyURL(c *C) {
	url := "http://127.0.0.1:14124"
	c.Assert(t.d.run([]string{"register", url}), Equals, 1)
}

func (t *RegisterTests) TestRegister(c *C) {
	url := "http://127.0.0.1:14124"
	c.Assert(t.d.run([]string{"register", url, "example"}), Equals, 0)
}
