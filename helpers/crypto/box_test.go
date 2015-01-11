package crypto

import (
	"crypto/rand"

	. "gopkg.in/check.v1"
)

type TestBox struct {
	encKey [EncryptionKeySize]byte
}

var _ = Suite(&TestBox{})

func (t *TestBox) SetUpTest(c *C) {
	t.encKey = [EncryptionKeySize]byte{}
	_, err := rand.Read(t.encKey[:])
	c.Assert(err, IsNil)
}

func (t *TestBox) getBox() *Box {
	return NewBox(t.encKey)
}

func (t *TestBox) TestEncryptionDecryption(c *C) {
	testData := []byte("This is testdata")
	encrypted, err := t.getBox().EncryptWithRandomKey(testData)
	c.Assert(err, IsNil)

	decrypted, err := t.getBox().DecryptContent(encrypted)
	c.Assert(err, IsNil)

	c.Assert(testData, DeepEquals, decrypted)
}

func (t *TestBox) TestDecryptionUnderMinimalLength(c *C) {
	testData := []byte{}
	_, err := t.getBox().DecryptContent(testData)

	c.Assert(err, NotNil)
}

func (t *TestBox) TestDecryptionOtherKey(c *C) {
	testData := []byte("This is testdata")
	encrypted, err := t.getBox().EncryptWithRandomKey(testData)
	c.Assert(err, IsNil)

	t.encKey = [EncryptionKeySize]byte{}

	_, err = t.getBox().DecryptContent(encrypted)

	c.Assert(err, NotNil)
}
