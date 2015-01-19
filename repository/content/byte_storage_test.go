package content

import (
	. "gopkg.in/check.v1"
)

type ByteStorageTests struct {
	dir string
}

var _ = Suite(&ByteStorageTests{})

func (t *ByteStorageTests) SetUpTest(c *C) {
	t.dir = c.MkDir()
}

func (t *ByteStorageTests) Test(c *C) {
	s := NewFileStorage(t.dir)
	b := NewByteStorage(s)
	myContent := []byte("asd")
	err := b.SetBytes("foo", myContent)
	c.Assert(err, IsNil)
	gotContent, err := b.GetBytes("foo")
	c.Assert(err, IsNil)
	c.Assert(gotContent, DeepEquals, myContent)
}
