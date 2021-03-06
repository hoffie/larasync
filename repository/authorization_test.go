package repository

import (
	"bytes"
	"crypto/rand"

	. "gopkg.in/check.v1"

	"github.com/hoffie/larasync/repository/odf"
)

type AuthorizationTest struct {
	SigningKey    [PrivateKeySize]byte
	EncryptionKey [EncryptionKeySize]byte
	HashingKey    [HashingKeySize]byte
}

var _ = Suite(&AuthorizationTest{})

func (t *AuthorizationTest) SetUpTest(c *C) {
	t.SigningKey = [PrivateKeySize]byte{}
	t.EncryptionKey = [EncryptionKeySize]byte{}

	privateKeyBytes := make([]byte, PrivateKeySize)
	_, err := rand.Read(privateKeyBytes)
	c.Assert(err, IsNil)

	encryptionKeyBytes := make([]byte, EncryptionKeySize)
	_, err = rand.Read(encryptionKeyBytes)
	c.Assert(err, IsNil)

	hashingKeyBytes := make([]byte, HashingKeySize)
	_, err = rand.Read(hashingKeyBytes)
	c.Assert(err, IsNil)

	copy(t.SigningKey[:], privateKeyBytes[0:PrivateKeySize])
	copy(t.EncryptionKey[:], encryptionKeyBytes[0:EncryptionKeySize])
	copy(t.HashingKey[:], hashingKeyBytes[0:HashingKeySize])
}

func (t *AuthorizationTest) getAuthorization() *Authorization {
	return &Authorization{
		SigningKey:    t.SigningKey,
		EncryptionKey: t.EncryptionKey,
		HashingKey:    t.HashingKey,
	}
}

func (t *AuthorizationTest) getPbAuthorization() *odf.Authorization {
	return &odf.Authorization{
		SigningKey:    t.SigningKey[:],
		EncryptionKey: t.EncryptionKey[:],
		HashingKey:    t.HashingKey[:],
	}
}

func (t *AuthorizationTest) TestConversionToPb(c *C) {
	authorization := t.getAuthorization()
	pbAuthorization, err := authorization.toPb()
	c.Assert(err, IsNil)

	assertEqualAuthorizationRepresentation(c, authorization, pbAuthorization)
}

func (t *AuthorizationTest) TestConversionFromPb(c *C) {
	pbAuthorization := t.getPbAuthorization()
	authorization := newAuthorizationFromPb(pbAuthorization)
	assertEqualAuthorizationRepresentation(c, authorization, pbAuthorization)
}

func assertEqualAuthorizationRepresentation(
	c *C,
	authorization *Authorization,
	pbAuthorization *odf.Authorization,
) {
	c.Assert(
		authorization.SigningKey[:],
		DeepEquals,
		pbAuthorization.GetSigningKey(),
	)
	c.Assert(
		authorization.EncryptionKey[:],
		DeepEquals,
		pbAuthorization.GetEncryptionKey(),
	)
	c.Assert(
		authorization.HashingKey[:],
		DeepEquals,
		pbAuthorization.GetHashingKey(),
	)
}

func assertEqualAuthorizations(
	c *C,
	authorization *Authorization,
	otherAuthorization *Authorization,
) {
	c.Assert(
		authorization.SigningKey,
		DeepEquals,
		otherAuthorization.SigningKey,
	)

	c.Assert(
		authorization.EncryptionKey,
		DeepEquals,
		otherAuthorization.EncryptionKey,
	)

	c.Assert(
		authorization.HashingKey,
		DeepEquals,
		otherAuthorization.HashingKey,
	)
}

func (t *AuthorizationTest) TestReadFrom(c *C) {
	authorization := t.getAuthorization()
	buffer := &bytes.Buffer{}
	_, err := authorization.WriteTo(buffer)
	c.Assert(err, IsNil)

	otherAuth := &Authorization{}
	_, err = otherAuth.ReadFrom(buffer)
	c.Assert(err, IsNil)

	assertEqualAuthorizations(c, authorization, otherAuth)
}

func (t *AuthorizationTest) TestReadFromError(c *C) {
	authorization := t.getAuthorization()
	buffer := &bytes.Buffer{}
	_, err := authorization.WriteTo(buffer)
	c.Assert(err, IsNil)

	data := buffer.Bytes()
	copy(data, make([]byte, len(data)))

	otherAuth := &Authorization{}
	_, err = otherAuth.ReadFrom(buffer)
	c.Assert(err, NotNil)
}
