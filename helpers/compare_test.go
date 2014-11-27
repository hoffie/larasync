package helpers

import (
	"testing"

	. "gopkg.in/check.v1"
)

type Tests struct{}

var _ = Suite(&Tests{})

func TestCompare(t *testing.T) {
	TestingT(t)
}

func (t *Tests) TestConstantTimeBytesEqualDiff(c *C) {
	c.Assert(ConstantTimeBytesEqual([]byte("a"), []byte("b")), Equals, false)
}

func (t *Tests) TestConstantTimeBytesEqualLengthDiff(c *C) {
	c.Assert(ConstantTimeBytesEqual([]byte("a"), []byte("b")), Equals, false)
}

func (t *Tests) TestConstantTimeBytesEqualOk(c *C) {
	c.Assert(ConstantTimeBytesEqual([]byte("a"), []byte("a")), Equals, true)
}
