package repository

import (
	. "gopkg.in/check.v1"
)

type ByteContentStorageTests struct {
	dir string
}

var _ = Suite(&ByteContentStorageTests{})

func (t *ByteContentStorageTests) SetUpTest(c *C) {
	t.dir = c.MkDir()
}

func (t *ByteContentStorageTests) Test(c *C) {
	s := newFileContentStorage(t.dir)
	b := newByteContentStorage(s)
	myContent := []byte("asd")
	err := b.SetBytes("foo", myContent)
	c.Assert(err, IsNil)
	gotContent, err := b.GetBytes("foo")
	c.Assert(err, IsNil)
	c.Assert(gotContent, DeepEquals, myContent)
}
