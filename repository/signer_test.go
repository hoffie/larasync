package repository

import (
	"bytes"
	"io"

	"github.com/agl/ed25519"
	. "gopkg.in/check.v1"
)

type SignerTests struct{}

var _ = Suite(&SignerTests{})

func (t *SignerTests) TestReadAndWrite(c *C) {
	testBytes := []byte("Test")
	pubKey, privKey, err := ed25519.GenerateKey(
		bytes.NewBufferString("just some deterministic 'random' bytes"))
	c.Assert(err, IsNil)

	data := &bytes.Buffer{}
	s := NewSigningWriter(*privKey, data)
	written, err := s.Write(testBytes)
	c.Assert(err, IsNil)
	c.Assert(written, Equals, len(testBytes))
	err = s.Finalize()
	c.Assert(err, IsNil)

	dataReader := bytes.NewReader(data.Bytes())
	v, err := NewVerifyingReader(*pubKey, dataReader)
	c.Assert(err, IsNil)
	gotten := &bytes.Buffer{}
	_, err = io.Copy(gotten, v)

	c.Assert(gotten.Bytes(), DeepEquals, testBytes)
	c.Assert(v.VerifyAfterRead(), Equals, true)
}

func (t *SignerTests) TestReadAndWriteTamper(c *C) {
	testBytes := []byte("Test")
	pubKey, privKey, err := ed25519.GenerateKey(
		bytes.NewBufferString("just some deterministic 'random' bytes"))
	c.Assert(err, IsNil)

	data := &bytes.Buffer{}
	s := NewSigningWriter(*privKey, data)
	written, err := s.Write(testBytes)
	c.Assert(err, IsNil)
	c.Assert(written, Equals, len(testBytes))
	err = s.Finalize()
	c.Assert(err, IsNil)

	byteData := data.Bytes()
	byteData[0] = 'F'
	dataReader := bytes.NewReader(byteData)
	v, err := NewVerifyingReader(*pubKey, dataReader)
	c.Assert(err, IsNil)
	gotten := &bytes.Buffer{}
	_, err = io.Copy(gotten, v)
	c.Assert(err, IsNil)
	c.Assert(v.VerifyAfterRead(), Equals, false)
}

func (t *SignerTests) TestReadAndWriteWrongPubkey(c *C) {
	testBytes := []byte("Foo")
	_, privKey, err := ed25519.GenerateKey(
		bytes.NewBufferString("just some deterministic 'random' bytes"))
	c.Assert(err, IsNil)
	pubKey2, _, err := ed25519.GenerateKey(
		bytes.NewBufferString("and some more 'random' data......"))
	c.Assert(err, IsNil)

	data := &bytes.Buffer{}
	s := NewSigningWriter(*privKey, data)
	written, err := s.Write(testBytes)
	c.Assert(err, IsNil)
	c.Assert(written, Equals, len(testBytes))
	err = s.Finalize()
	c.Assert(err, IsNil)

	byteData := data.Bytes()
	dataReader := bytes.NewReader(byteData)
	v, err := NewVerifyingReader(*pubKey2, dataReader)
	c.Assert(err, IsNil)
	gotten := &bytes.Buffer{}
	_, err = io.Copy(gotten, v)
	c.Assert(err, IsNil)
	c.Assert(v.VerifyAfterRead(), Equals, false)
}
