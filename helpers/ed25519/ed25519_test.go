package ed25519

import (
	"bytes"
	e "github.com/agl/ed25519"
	"testing"

	. "gopkg.in/check.v1"
)

type Tests struct{}

var _ = Suite(&Tests{})

func Test(t *testing.T) {
	TestingT(t)
}

func (t *Tests) TestEd25519GetPublicFromPrivate(c *C) {
	fakeRandReader := bytes.NewBufferString("012345678901234567890123456789012")
	pub, priv, err := e.GenerateKey(fakeRandReader)
	c.Assert(err, IsNil)
	myPub := GetPublicKeyFromPrivate(*priv)
	c.Assert(err, IsNil)
	c.Assert(*pub, DeepEquals, myPub)
}
