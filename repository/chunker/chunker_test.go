package chunker

import (
	"io/ioutil"
	"path/filepath"

	. "gopkg.in/check.v1"
)

type ChunkerTests struct {
	dir string
}

var _ = Suite(&ChunkerTests{})

func (t *ChunkerTests) SetUpTest(c *C) {
	t.dir = c.MkDir()
}

func (t *ChunkerTests) TestHandleError(c *C) {
	_, err := New(filepath.Join(t.dir, "non-existing"), 16)
	c.Assert(err, NotNil)
}

func (t *ChunkerTests) TestHandleBadChunkSiez(c *C) {
	_, err := New("test", 15)
	c.Assert(err, NotNil)
}

func (t *ChunkerTests) TestNormal(c *C) {
	path := filepath.Join(t.dir, "test")
	err := ioutil.WriteFile(path, []byte("12345678901234567890"),
		0600)
	c.Assert(err, IsNil)

	ch, err := New(path, 16)
	c.Assert(err, IsNil)
	defer ch.Close()
	c.Assert(ch.HasNext(), Equals, true)
	chunk, err := ch.Next()
	c.Assert(err, IsNil)
	c.Assert(chunk, DeepEquals, []byte("1234567890123456"))
	c.Assert(ch.HasNext(), Equals, true)
	chunk, err = ch.Next()
	c.Assert(err, IsNil)
	c.Assert(chunk, DeepEquals, []byte("7890"))
	c.Assert(ch.HasNext(), Equals, false)
}
