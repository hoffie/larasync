package main

import (
	"bytes"
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) {
	TestingT(t)
}

type CommonTests struct {
	out *bytes.Buffer
	d   *Dispatcher
}

var _ = Suite(&CommonTests{})

func (t *CommonTests) SetUpTest(c *C) {
	t.out = new(bytes.Buffer)
	t.d = &Dispatcher{stderr: t.out}
}

func (t *CommonTests) TestEmptyArgs(c *C) {
	c.Assert(t.d.run([]string{}), Equals, 1)
}
