package repository

import (
	"bytes"

	. "gopkg.in/check.v1"
)

type MetadataTests struct{}

var _ = Suite(&MetadataTests{})

func (t *MetadataTests) TestSerialize(c *C) {
	m1 := Metadata{
		Type:             MetadataTypeFile,
		RepoRelativePath: "foo.txt",
	}
	buf := &bytes.Buffer{}
	written, err := m1.WriteTo(buf)
	c.Assert(err, IsNil)

	m2 := Metadata{}
	read, err := m2.ReadFrom(buf)
	c.Assert(err, IsNil)

	c.Assert(read, Equals, written)

	c.Assert(m1, DeepEquals, m2)
}

func (t *MetadataTests) TestSerializeDir(c *C) {
	m1 := Metadata{
		Type:             MetadataTypeDir,
		RepoRelativePath: "foo.txt",
	}
	buf := &bytes.Buffer{}
	written, err := m1.WriteTo(buf)
	c.Assert(err, IsNil)

	m2 := Metadata{}
	read, err := m2.ReadFrom(buf)
	c.Assert(err, IsNil)

	c.Assert(read, Equals, written)

	c.Assert(m1, DeepEquals, m2)
}
