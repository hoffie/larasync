package main

import (
	"bytes"
	"strings"

	. "gopkg.in/check.v1"
)

type AdminSecretTests struct {
	err *bytes.Buffer
	out *bytes.Buffer
	in  *bytes.Buffer
	d   *Dispatcher
	ts  *TestServer
}

var _ = Suite(&AdminSecretTests{})

func (t *AdminSecretTests) SetUpTest(c *C) {
	t.err = new(bytes.Buffer)
	t.out = new(bytes.Buffer)
	t.in = new(bytes.Buffer)
	t.d = &Dispatcher{stderr: t.err, stdout: t.out, stdin: t.in}
}

func (t *AdminSecretTests) TestShow(c *C) {
	_, err := t.in.Write([]byte("test\n"))
	c.Assert(err, IsNil)
	c.Assert(t.d.run([]string{"admin-secret"}), Equals, 0)
	res := t.out.String()
	pubkey := "52c6d86830ceecff385f843022af1a8a88f6de5be0d3e58d99d9e4377feb8c03"
	c.Assert(strings.HasSuffix(res, pubkey+"\n"), Equals, true)
}
