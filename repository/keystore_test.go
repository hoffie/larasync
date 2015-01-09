package repository

import (
	"github.com/agl/ed25519"

	. "gopkg.in/check.v1"
)

type KeyStoreTests struct {
	dir string
}

var _ = Suite(&KeyStoreTests{})

func (t *KeyStoreTests) SetUpTest(c *C) {
	t.dir = c.MkDir()
}

func (t *KeyStoreTests) TestEncryptionKey(c *C) {
	ks := NewKeyStore(t.dir)
	var k [EncryptionKeySize]byte
	k[0] = 'z'
	_, err := ks.EncryptionKey()
	c.Assert(err, NotNil)

	err = ks.SetEncryptionKey(k)
	c.Assert(err, IsNil)

	k2, err := ks.EncryptionKey()
	c.Assert(err, IsNil)
	c.Assert(k2, DeepEquals, k)
}

func (t *KeyStoreTests) TestSigningPrivkey(c *C) {
	ks := NewKeyStore(t.dir)
	var k [PrivateKeySize]byte
	k[0] = 'z'
	_, err := ks.SigningPrivateKey()
	c.Assert(err, NotNil)

	err = ks.SetSigningPrivateKey(k)
	c.Assert(err, IsNil)

	k2, err := ks.SigningPrivateKey()
	c.Assert(err, IsNil)
	c.Assert(k2, DeepEquals, k)
}

func (t *KeyStoreTests) TestHashingKey(c *C) {
	ks := NewKeyStore(t.dir)
	var k [HashingKeySize]byte
	k[0] = 'z'
	_, err := ks.HashingKey()
	c.Assert(err, NotNil)

	err = ks.SetHashingKey(k)
	c.Assert(err, IsNil)

	k2, err := ks.HashingKey()
	c.Assert(err, IsNil)
	c.Assert(k2, DeepEquals, k)
}

func (t *KeyStoreTests) TestCreateSigningKey(c *C) {
	ks := NewKeyStore(t.dir)
	err := ks.CreateSigningKey()
	c.Assert(err, IsNil)

	key, err := ks.SigningPrivateKey()
	c.Assert(err, IsNil)
	c.Assert(len(key), Equals, PrivateKeySize)
}

func (t *KeyStoreTests) TestCreateSigningKeyTestEncryption(c *C) {
	ks := NewKeyStore(t.dir)
	err := ks.CreateSigningKey()
	c.Assert(err, IsNil)

	priv, err := ks.SigningPrivateKey()
	c.Assert(err, IsNil)

	pub, err := ks.SigningPublicKey()
	c.Assert(err, IsNil)

	content := []byte("test")
	sig := ed25519.Sign(&priv, content)
	res := ed25519.Verify(&pub, content, sig)
	c.Assert(res, Equals, true)
}

func (t *KeyStoreTests) TestCreateHashingKey(c *C) {
	ks := NewKeyStore(t.dir)

	err := ks.CreateHashingKey()
	c.Assert(err, IsNil)

	key, err := ks.HashingKey()
	c.Assert(err, IsNil)
	c.Assert(len(key), Equals, HashingKeySize)
}

func (t *KeyStoreTests) TestCreateEncryptionKey(c *C) {
	ks := NewKeyStore(t.dir)
	err := ks.CreateEncryptionKey()
	c.Assert(err, IsNil)

	key, err := ks.EncryptionKey()
	c.Assert(err, IsNil)
	c.Assert(len(key), Equals, EncryptionKeySize)
}
