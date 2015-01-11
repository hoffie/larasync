package repository

import (
	"github.com/hoffie/larasync/repository/odf"
)

// Authorization is being used to pass the required data
// to authorize a new client to the server system.
type Authorization struct {
	SigningKey    [PrivateKeySize]byte
	EncryptionKey [EncryptionKeySize]byte
	HashingKey    [HashingKeySize]byte
}

// newAuthorizationFromPb returns a new Authorization object
// from the protobuf definition.
func newAuthorizationFromPb(pbAuthorization *odf.Authorization) *Authorization {
	protoSigningKey := pbAuthorization.GetSigningKey()
	protoEncryptionKey := pbAuthorization.GetEncryptionKey()
	protoHashingKey := pbAuthorization.GetHashingKey()

	auth := &Authorization{
		SigningKey:    [PrivateKeySize]byte{},
		EncryptionKey: [EncryptionKeySize]byte{},
		HashingKey:    [HashingKeySize]byte{},
	}

	copy(auth.SigningKey[:], protoSigningKey[0:PrivateKeySize])
	copy(auth.EncryptionKey[:], protoEncryptionKey[0:EncryptionKeySize])
	copy(auth.HashingKey[:], protoHashingKey[0:HashingKeySize])

	return auth
}

// toPb converts this Authorization to a protobuf Authorization.
// This is used by the encoder.
func (a *Authorization) toPb() (*odf.Authorization, error) {
	signingKey := make([]byte, PrivateKeySize)
	encryptionKey := make([]byte, EncryptionKeySize)
	hashingKey := make([]byte, HashingKeySize)

	copy(signingKey[:], a.SigningKey[0:PrivateKeySize])
	copy(encryptionKey[:], a.EncryptionKey[0:EncryptionKeySize])
	copy(hashingKey[:], a.HashingKey[0:HashingKeySize])

	protoAuthorization := &odf.Authorization{
		SigningKey:    signingKey,
		EncryptionKey: encryptionKey,
		HashingKey:    hashingKey,
	}

	return protoAuthorization, nil
}
