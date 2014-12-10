package repository

import (
	"bytes"

	. "gopkg.in/check.v1"
)

type NIBTests struct {}

var _ = Suite(&NIBTests{})

func (t *NIBTests) TestUUID(c *C) {
	n := NIB{}
	n.UUID = "1234"
	buf := &bytes.Buffer{}
	written, err := n.WriteTo(buf)
	c.Assert(err, IsNil)
	n2 := NIB{}
	read, err := n2.ReadFrom(buf)
	c.Assert(err, IsNil)
	c.Assert(written, Equals, read)
	c.Assert(n2.UUID, Equals, n.UUID)
}
