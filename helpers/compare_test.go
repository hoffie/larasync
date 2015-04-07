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

func (t *CompareTests) TestStringsEqualDiff(c *C) {
	c.Assert(StringsEqual([]string{"123", "456"}, []string{"123", "567"}), Equals, false)
}

func (t *CompareTests) TestStringsEqualLengthDiff(c *C) {
	c.Assert(StringsEqual([]string{"123", "456"}, []string{"123"}), Equals, false)
}

func (t *CompareTests) TestStringsEqual(c *C) {
	c.Assert(StringsEqual([]string{"123", "456"}, []string{"123", "456"}), Equals, true)
}
