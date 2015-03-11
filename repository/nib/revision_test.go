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
