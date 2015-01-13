package bincontainer

import (
	"bytes"
	"io"
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) {
	TestingT(t)
}

type Tests struct{}

var _ = Suite(&Tests{})

func (t *Tests) TestEncodeDecode(c *C) {
	buf := &bytes.Buffer{}
	testChunk1 := []byte("foo")
	testChunk2 := []byte("bar")

	e := NewEncoder(buf)
	err := e.WriteChunk(testChunk1)
	c.Assert(err, IsNil)

	err = e.WriteChunk(testChunk2)
	c.Assert(err, IsNil)

	r := NewDecoder(buf)
	chunk, err := r.ReadChunk()
	c.Assert(err, IsNil)
	c.Assert(chunk, DeepEquals, testChunk1)

	chunk, err = r.ReadChunk()
	c.Assert(err, IsNil)
	c.Assert(chunk, DeepEquals, testChunk2)

	chunk, err = r.ReadChunk()
	c.Assert(err, Equals, io.EOF)
}

func (t *Tests) TestDecodeEmpty(c *C) {
	d := NewDecoder(&bytes.Buffer{})
	_, err := d.ReadChunk()
	c.Assert(err, Equals, io.EOF)
}

func (t *Tests) TestDecodeEmptyChunk(c *C) {
	d := NewDecoder(bytes.NewBufferString("\x00\x00\x00\x00"))
	chunk, err := d.ReadChunk()
	c.Assert(err, IsNil)
	c.Assert(chunk, DeepEquals, []byte{})
}

func (t *Tests) TestDecodeTruncatedLength(c *C) {
	d := NewDecoder(bytes.NewBufferString("\x01"))
	_, err := d.ReadChunk()
	c.Assert(err, Equals, ErrIncomplete)
}

func (t *Tests) TestDecodeTruncatedContent(c *C) {
	d := NewDecoder(bytes.NewBufferString("\x01\x01\x01\x01"))
	_, err := d.ReadChunk()
	c.Assert(err, Equals, ErrIncomplete)
}

type eofReader4ZeroBytes struct {
	state int
}

func (r *eofReader4ZeroBytes) Read(buf []byte) (int, error) {
	r.state++
	switch r.state {
	case 1:
		buf[0] = 0
		buf[1] = 0
		buf[2] = 0
		buf[3] = 0
		return 4, io.EOF
	}
	return 0, io.EOF
}

type eofReaderLength1Data1 struct {
	state int
}

func (r *eofReaderLength1Data1) Read(buf []byte) (int, error) {
	r.state++
	switch r.state {
	case 1:
		buf[0] = 1
		buf[1] = 0
		buf[2] = 0
		buf[3] = 0
		return 4, nil
	}
	buf[0] = 97
	return 1, io.EOF
}

func (t *Tests) TestDecodeLengthWithEOF(c *C) {
	r := &eofReader4ZeroBytes{}
	d := NewDecoder(r)
	chunk, err := d.ReadChunk()
	c.Assert(err, IsNil)
	c.Assert(chunk, DeepEquals, []byte{})
}

func (t *Tests) TestDecodeDataWithEOF(c *C) {
	r := &eofReaderLength1Data1{}
	d := NewDecoder(r)
	chunk, err := d.ReadChunk()
	c.Assert(err, IsNil)
	c.Assert(chunk, DeepEquals, []byte{'a'})
}
