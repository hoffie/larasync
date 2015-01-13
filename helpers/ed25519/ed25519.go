package ed25519

import (
	"bytes"
	"crypto/rand"
	"io"

	e "github.com/agl/ed25519"
)

// GetPublicKeyFromPrivate takes an Ed25519 private key and generates the public
// key compartment from it.
func GetPublicKeyFromPrivate(privateKey [e.PrivateKeySize]byte) [e.PublicKeySize]byte {
	buf := make([]byte, len(privateKey))
	copy(buf, privateKey[0:e.PrivateKeySize])
	keyProvidingReader := bytes.NewBuffer([]byte(buf))
	pub, _, _ := e.GenerateKey(keyProvidingReader)
	return *pub
}

// GenerateKey creates ed25519 keys with the standard rand.Reader.
func GenerateKey() (publicKey *[e.PublicKeySize]byte, privateKey *[e.PrivateKeySize]byte, err error) {
	return GenerateKeyFrom(rand.Reader)
}

// GenerateKeyFrom creates ed25519 keys and gets its entropy from the passed rand reader.
func GenerateKeyFrom(rand io.Reader) (publicKey *[e.PublicKeySize]byte, privateKey *[e.PrivateKeySize]byte, err error) {
	return e.GenerateKey(rand)
}
