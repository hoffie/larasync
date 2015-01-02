package repository

import (
	"bytes"
	"time"

	. "gopkg.in/check.v1"
)

type NIBTests struct{}

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

func (t *NIBTests) TestRevisionEnDecode(c *C) {
	r := &Revision{MetadataID: "1234"}
	r.AddContentID("5678")
	r.UTCTimestamp = time.Now().UnixNano()
	r.DeviceID = "localhost"
	n := NIB{}
	n.AppendRevision(r)
	buf := &bytes.Buffer{}
	written, err := n.WriteTo(buf)
	c.Assert(err, IsNil)
	n2 := NIB{}
	read, err := n2.ReadFrom(buf)
	c.Assert(err, IsNil)
	c.Assert(written, Equals, read)
	r2, err := n2.LatestRevision()
	c.Assert(err, IsNil)
	c.Assert(r, DeepEquals, r2)
}

func (t *NIBTests) TestLatestRevisionFailure(c *C) {
	n := NIB{}
	r, err := n.LatestRevision()
	c.Assert(r, IsNil)
	c.Assert(err, NotNil)
}
