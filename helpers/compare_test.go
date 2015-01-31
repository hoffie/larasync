package helpers

import (
	. "gopkg.in/check.v1"
)

type CompareTests struct{}

var _ = Suite(&CompareTests{})

func (t *CompareTests) TestConstantTimeBytesEqualDiff(c *C) {
	c.Assert(ConstantTimeBytesEqual([]byte("a"), []byte("b")), Equals, false)
}

func (t *CompareTests) TestConstantTimeBytesEqualLengthDiff(c *C) {
	c.Assert(ConstantTimeBytesEqual([]byte("a"), []byte("aa")), Equals, false)
}

func (t *CompareTests) TestConstantTimeBytesEqualOk(c *C) {
	c.Assert(ConstantTimeBytesEqual([]byte("a"), []byte("a")), Equals, true)
}
