package repository

import (
	"github.com/agl/ed25519"

	. "gopkg.in/check.v1"
)

type KeyStoreTests struct {
	dir     string
	storage *FileContentStorage
	ks      *KeyStore
}

var _ = Suite(&KeyStoreTests{})

func (t *KeyStoreTests) SetUpTest(c *C) {
	t.dir = c.MkDir()
	t.storage = newFileContentStorage(t.dir)
	t.ks = NewKeyStore(t.storage)
}

func (t *KeyStoreTests) TestEncryptionKey(c *C) {
	var k [EncryptionKeySize]byte
	k[0] = 'z'
	_, err := t.ks.EncryptionKey()
	c.Assert(err, NotNil)

	err = t.ks.SetEncryptionKey(k)
	c.Assert(err, IsNil)

	k2, err := t.ks.EncryptionKey()
	c.Assert(err, IsNil)
	c.Assert(k2, DeepEquals, k)
}

func (t *KeyStoreTests) TestSigningPrivkey(c *C) {
	var k [PrivateKeySize]byte
	k[0] = 'z'
	_, err := t.ks.SigningPrivateKey()
	c.Assert(err, NotNil)

	err = t.ks.SetSigningPrivateKey(k)
	c.Assert(err, IsNil)

	k2, err := t.ks.SigningPrivateKey()
	c.Assert(err, IsNil)
	c.Assert(k2, DeepEquals, k)
}

func (t *KeyStoreTests) TestHashingKey(c *C) {
	var k [HashingKeySize]byte
	k[0] = 'z'
	_, err := t.ks.HashingKey()
	c.Assert(err, NotNil)

	err = t.ks.SetHashingKey(k)
	c.Assert(err, IsNil)

	k2, err := t.ks.HashingKey()
	c.Assert(err, IsNil)
	c.Assert(k2, DeepEquals, k)
}

func (t *KeyStoreTests) TestCreateSigningKey(c *C) {
	err := t.ks.CreateSigningKey()
	c.Assert(err, IsNil)

	key, err := t.ks.SigningPrivateKey()
	c.Assert(err, IsNil)
	c.Assert(len(key), Equals, PrivateKeySize)
}

func (t *KeyStoreTests) TestCreateSigningKeyTestEncryption(c *C) {
	err := t.ks.CreateSigningKey()
	c.Assert(err, IsNil)

	priv, err := t.ks.SigningPrivateKey()
	c.Assert(err, IsNil)

	pub, err := t.ks.SigningPublicKey()
	c.Assert(err, IsNil)

	content := []byte("test")
	sig := ed25519.Sign(&priv, content)
	res := ed25519.Verify(&pub, content, sig)
	c.Assert(res, Equals, true)
}

func (t *KeyStoreTests) TestCreateHashingKey(c *C) {
	err := t.ks.CreateHashingKey()
	c.Assert(err, IsNil)

	key, err := t.ks.HashingKey()
	c.Assert(err, IsNil)
	c.Assert(len(key), Equals, HashingKeySize)
}

func (t *KeyStoreTests) TestCreateEncryptionKey(c *C) {
	err := t.ks.CreateEncryptionKey()
	c.Assert(err, IsNil)

	key, err := t.ks.EncryptionKey()
	c.Assert(err, IsNil)
	c.Assert(len(key), Equals, EncryptionKeySize)
}
