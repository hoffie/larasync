package repository

import (
	"bytes"
	"io"

	"github.com/golang/protobuf/proto"

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
	auth := &Authorization{
		SigningKey:    [PrivateKeySize]byte{},
		EncryptionKey: [EncryptionKeySize]byte{},
		HashingKey:    [HashingKeySize]byte{},
	}

	auth.setFromPb(pbAuthorization)

	return auth
}

// setFromPb is used to copy data from a protobuf Authorization to the
// this Authorization struct.
func (a *Authorization) setFromPb(pbAuthorization *odf.Authorization) {
	protoSigningKey := pbAuthorization.GetSigningKey()
	protoEncryptionKey := pbAuthorization.GetEncryptionKey()
	protoHashingKey := pbAuthorization.GetHashingKey()

	copy(a.SigningKey[:], protoSigningKey[0:PrivateKeySize])
	copy(a.EncryptionKey[:], protoEncryptionKey[0:EncryptionKeySize])
	copy(a.HashingKey[:], protoHashingKey[0:HashingKeySize])
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

// ReadFrom fills this Authorization's data with the contents supplied by
// the binary representation available through the given reader.
func (a *Authorization) ReadFrom(r io.Reader) (int64, error) {
	buf := &bytes.Buffer{}
	read, err := io.Copy(buf, r)
	if err != nil {
		return read, err
	}
	pb := &odf.Authorization{}
	err = proto.Unmarshal(buf.Bytes(), pb)
	if err != nil {
		return read, err
	}

	a.setFromPb(pb)
	return read, nil
}

// WriteTo encodes this Authorization to the supplied Writer in binary form.
// Returns the number of bytes written and an error if applicable.
func (a *Authorization) WriteTo(w io.Writer) (int64, error) {
	pb, err := a.toPb()
	if err != nil {
		return 0, err
	}

	buf, err := proto.Marshal(pb)

	if err != nil {
		return 0, err
	}
	written, err := io.Copy(w, bytes.NewBuffer(buf))
	return written, err
}
