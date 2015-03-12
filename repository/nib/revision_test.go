package nib

import (
	. "gopkg.in/check.v1"
)

var _ = Suite(&RevisionTests{})

type RevisionTests struct {
	dir string
}

func (t *RevisionTests) TestHasSameContentEmptyMetadata(c *C) {
	rev1 := &Revision{MetadataID: "123"}
	rev2 := &Revision{}
	c.Assert(rev1.HasSameContent(rev2), Equals, false)
}

func (t *RevisionTests) TestHasSameContent(c *C) {
	rev1 := &Revision{MetadataID: "123", ContentIDs: []string{"34", "45"}}
	rev2 := &Revision{MetadataID: "123", ContentIDs: []string{"34", "45"}}
	c.Assert(rev1.HasSameContent(rev2), Equals, true)
	c.Assert(rev2.HasSameContent(rev1), Equals, true)
}

func (t *RevisionTests) TestHasSameContentSelf(c *C) {
	rev := &Revision{MetadataID: "ab", ContentIDs: []string{"x"}}
	c.Assert(rev.HasSameContent(rev), Equals, true)
}

func (t *RevisionTests) TestIsDeletePositive(c *C) {
	rev := &Revision{MetadataID: "123", ContentIDs: []string{}}
	c.Assert(rev.IsDelete(), Equals, true)
}

func (t *RevisionTests) TestIsDeleteNegative(c *C) {
	rev := &Revision{MetadataID: "123", ContentIDs: []string{"34", "45"}}
	c.Assert(rev.IsDelete(), Equals, false)
}

func (t *RevisionTests) TestCopy(c *C) {
	rev := &Revision{
		MetadataID:   "123",
		ContentIDs:   []string{"34", "45"},
		UTCTimestamp: 100,
		DeviceID:     "asdf",
	}

	rev2 := rev.Clone()

	c.Assert(rev2.MetadataID, Equals, "123")
	c.Assert(rev2.ContentIDs, DeepEquals, []string{"34", "45"})
	c.Assert(rev2.UTCTimestamp, Equals, int64(100))
	c.Assert(rev2.DeviceID, Equals, "asdf")
}

func (t *RevisionTests) TestCopyEmptyContentIDs(c *C) {
	rev := &Revision{
		ContentIDs: []string{},
	}

	rev2 := rev.Clone()
	c.Assert(rev2.ContentIDs, DeepEquals, []string{})
}
