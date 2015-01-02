package repository

import (
	"bytes"
	"time"

	. "gopkg.in/check.v1"
)

type NIBTests struct{}

var _ = Suite(&NIBTests{})

func (t *NIBTests) TestEncode(c *C) {
	n := NIB{}
	n.UUID = "1234"
	n.HistoryOffset = 30
	buf := &bytes.Buffer{}
	written, err := n.WriteTo(buf)
	c.Assert(err, IsNil)
	n2 := NIB{}
	read, err := n2.ReadFrom(buf)
	c.Assert(err, IsNil)
	c.Assert(written, Equals, read)
	c.Assert(n, DeepEquals, n2)
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

func (t *NIBTests) TestRevisionsTotalSimple(c *C) {
	n := NIB{}
	n.AppendRevision(&Revision{})
	n.AppendRevision(&Revision{})
	c.Assert(n.RevisionsTotal(), Equals, int64(2))
}

func (t *NIBTests) TestRevisionsTotalhWithOffset(c *C) {
	n := NIB{
		HistoryOffset: 1097,
	}
	n.AppendRevision(&Revision{})
	n.AppendRevision(&Revision{})
	c.Assert(n.RevisionsTotal(), Equals, int64(1097+2))
}
