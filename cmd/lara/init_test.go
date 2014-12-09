package main

import (
	"testing"

	. "gopkg.in/check.v1"
)

func TestInit(t *testing.T) {
	TestingT(t)
}

type InitTests struct {
	dir string
}

var _ = Suite(&InitTests{})

func (t *InitTests) SetUpTest(c *C) {
	t.dir = c.MkDir()
}

func (t *InitTests) TestCall(c *C) {
	c.Assert(dispatch([]string{"init"}), Equals, 0)
}
