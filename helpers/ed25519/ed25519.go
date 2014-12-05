package ed25519

import (
	"bytes"

	e "github.com/agl/ed25519"
)

// GetPublicKeyFromPrivate takes an Ed25519 private key and generates the public
// key compartment from it.
func GetPublicKeyFromPrivate(privateKey [e.PrivateKeySize]byte) ([e.PublicKeySize]byte, error) {
	buf := make([]byte, len(privateKey))
	copy(buf, privateKey[0:e.PrivateKeySize])
	keyProvidingReader := bytes.NewBuffer([]byte(buf))
	pub, _, err := e.GenerateKey(keyProvidingReader)
	return *pub, err
}
