package repository

import (
	"crypto/rand"
	"encoding/hex"
	"os"

	. "gopkg.in/check.v1"

	edhelpers "github.com/hoffie/larasync/helpers/ed25519"
)

type AuthorizationManagerTests struct {
	dir                 string
	am                  *AuthorizationManager
	signaturePrivateKey [PrivateKeySize]byte
	encryptionKey       [EncryptionKeySize]byte
}

var _ = Suite(&AuthorizationManagerTests{})

func (t *AuthorizationManagerTests) SetUpTest(c *C) {
	t.dir = c.MkDir()
	t.am = newAuthorizationManager(&FileContentStorage{
		StoragePath: t.dir,
	})
	rand.Read(t.encryptionKey[:])
	rand.Read(t.signaturePrivateKey[:])
}

func (t *AuthorizationManagerTests) signaturePublicKey() [PublicKeySize]byte {
	return edhelpers.GetPublicKeyFromPrivate(t.signaturePrivateKey)
}

func (t *AuthorizationManagerTests) signaturePublicKeyString() string {
	pubKey := t.signaturePublicKey()
	return hex.EncodeToString(pubKey[:])
}

func (t *AuthorizationManagerTests) testAuthorization() *Authorization {
	auth := &Authorization{
		SigningKey:    [PrivateKeySize]byte{},
		EncryptionKey: [EncryptionKeySize]byte{},
		HashingKey:    [HashingKeySize]byte{},
	}

	rand.Read(auth.EncryptionKey[:])
	rand.Read(auth.SigningKey[:])
	rand.Read(auth.HashingKey[:])

	return auth
}

func (t *AuthorizationManagerTests) addAuthorization(c *C, auth *Authorization) {
	err := t.am.Set(
		t.signaturePublicKey(),
		t.encryptionKey,
		auth,
	)
	c.Assert(err, IsNil)
}

func (t *AuthorizationManagerTests) TestGet(c *C) {
	auth := t.testAuthorization()
	t.addAuthorization(c, auth)

	retrievedAuth, err := t.am.Get(t.signaturePublicKey(), t.encryptionKey)
	c.Assert(err, IsNil)

	c.Assert(auth.EncryptionKey, DeepEquals, retrievedAuth.EncryptionKey)
	c.Assert(auth.HashingKey, DeepEquals, retrievedAuth.HashingKey)
	c.Assert(auth.SigningKey, DeepEquals, retrievedAuth.SigningKey)
}

func (t *AuthorizationManagerTests) TestGetDecryptionError(c *C) {
	auth := t.testAuthorization()
	t.addAuthorization(c, auth)

	_, err := t.am.Get(t.signaturePublicKey(), [EncryptionKeySize]byte{})
	c.Assert(err, NotNil)
}

func (t *AuthorizationManagerTests) TestGetNotFound(c *C) {
	_, err := t.am.Get(t.signaturePublicKey(), t.encryptionKey)
	c.Assert(os.IsNotExist(err), Equals, true)
}

func (t *AuthorizationManagerTests) TestExistsNegative(c *C) {
	c.Assert(t.am.Exists(t.signaturePublicKey()), Equals, false)
}

func (t *AuthorizationManagerTests) TestExistsPositive(c *C) {
	auth := t.testAuthorization()
	t.addAuthorization(c, auth)

	c.Assert(t.am.Exists(t.signaturePublicKey()), Equals, true)
}

func (t *AuthorizationManagerTests) TestExistsForStringNegative(c *C) {
	c.Assert(t.am.ExistsForString(t.signaturePublicKeyString()), Equals, false)
}

func (t *AuthorizationManagerTests) TestExistsForStringPositive(c *C) {
	auth := t.testAuthorization()
	t.addAuthorization(c, auth)

	c.Assert(t.am.ExistsForString(t.signaturePublicKeyString()), Equals, true)
}

func (t *AuthorizationManagerTests) TestDelete(c *C) {
	auth := t.testAuthorization()
	t.addAuthorization(c, auth)
	err := t.am.Delete(t.signaturePublicKey())

	c.Assert(err, IsNil)
	c.Assert(t.am.Exists(t.signaturePublicKey()), Equals, false)
}

func (t *AuthorizationManagerTests) TestDeleteError(c *C) {
	err := t.am.Delete(t.signaturePublicKey())
	c.Assert(os.IsNotExist(err), Equals, true)
}

func (t *AuthorizationManagerTests) TestDeleteString(c *C) {
	auth := t.testAuthorization()
	t.addAuthorization(c, auth)
	err := t.am.DeleteForString(t.signaturePublicKeyString())

	c.Assert(err, IsNil)
	c.Assert(t.am.Exists(t.signaturePublicKey()), Equals, false)
}

func (t *AuthorizationManagerTests) TestDeleteForStringError(c *C) {
	err := t.am.DeleteForString(t.signaturePublicKeyString())
	c.Assert(os.IsNotExist(err), Equals, true)
}

func (t *AuthorizationManagerTests) TestSet(c *C) {
	auth := t.testAuthorization()
	err := t.am.Set(
		t.signaturePublicKey(),
		t.encryptionKey,
		auth,
	)
	c.Assert(err, IsNil)
}
