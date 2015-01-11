package crypto

import (
	"crypto/rand"
	"errors"

	"code.google.com/p/go.crypto/nacl/secretbox"
)

const (
	// EncryptionKeySize is the keySize which is used to encrypt data.
	EncryptionKeySize = 32

	// secretbox nonceSize
	nonceSize = 24

	// pre-computed minimal length of ciphertext; anything less cannot be valid
	// and will be rejected before attempting any other operations.
	encryptedContentMinSize = 2*(nonceSize+secretbox.Overhead) + EncryptionKeySize
)

// Box encapsulates the encryption and decryption of files with a given
// Encryption key.
type Box struct {
	privateKey [EncryptionKeySize]byte
}

// NewBox initializes a Box with the passed encryption key, which can be used
// to encrypt and decrypt data.
func NewBox(privateKey [EncryptionKeySize]byte) *Box {
	return &Box{
		privateKey: privateKey,
	}
}

// EncryptWithRandomKey takes a piece of data, encrypts it with a random
// key and returns the result, prefixed by the random key encrypted by
// the repository encryption key.
func (b *Box) EncryptWithRandomKey(data []byte) ([]byte, error) {
	// first generate and encrypt the per-file key and append it to
	// the result buffer:
	var nonce1 [nonceSize]byte
	_, err := rand.Read(nonce1[:])
	if err != nil {
		return nil, err
	}

	var fileKey [32]byte
	_, err = rand.Read(fileKey[:])
	if err != nil {
		return nil, err
	}

	encryptionKey := b.privateKey
	out := nonce1[:]
	out = secretbox.Seal(out, fileKey[:], &nonce1, &encryptionKey)

	// then append the actual encrypted contents
	var nonce2 [nonceSize]byte
	_, err = rand.Read(nonce2[:])
	if err != nil {
		return nil, err
	}
	out = append(out, nonce2[:]...)
	out = secretbox.Seal(out, data, &nonce2, &fileKey)
	return out, nil
}

// DecryptContent is the counter-part of EncryptWithRandomKey, i.e.
// it returns the plain text again.
func (b *Box) DecryptContent(enc []byte) ([]byte, error) {
	if len(enc) < encryptedContentMinSize {
		return nil, errors.New("truncated ciphertext")
	}

	// first decrypt the file-specific key using the master key
	var nonce [nonceSize]byte
	readNonce := func() {
		copy(nonce[:], enc[:nonceSize])
		enc = enc[nonceSize:]
	}
	readNonce()
	encryptionKey := b.privateKey

	l := EncryptionKeySize + secretbox.Overhead
	encryptedFileKey := enc[:l]
	enc = enc[l:]
	var fileKey []byte
	fileKey, success := secretbox.Open(fileKey, encryptedFileKey, &nonce, &encryptionKey)
	if !success {
		return nil, errors.New("file key decryption failed")
	}

	var arrFileKey [EncryptionKeySize]byte
	copy(arrFileKey[:], fileKey)

	readNonce()
	var content []byte
	content, success = secretbox.Open(content, enc, &nonce, &arrFileKey)
	if !success {
		return nil, errors.New("content decryption failed")
	}

	return content, nil
}
