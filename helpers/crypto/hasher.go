package crypto

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
)

const (
	// HashingKeySize is the amount of bytes used to initialize
	// the Hasher for Hashing purposes.
	HashingKeySize = 32
)

// Hasher is a helper which can be used to Hash bytes of data
// with a given key as initialization vector.
type Hasher struct {
	hashingKey [HashingKeySize]byte
}

// NewHasher is used to generate a new Hasher with the passed
// key as initialization vector.
func NewHasher(hashingKey [HashingKeySize]byte) *Hasher {
	return &Hasher{
		hashingKey: hashingKey,
	}
}

// Hash hashes the given data and returns the result as bytes.
func (h *Hasher) Hash(chunk []byte) []byte {
	key := h.hashingKey
	hasher := hmac.New(sha512.New, key[:])
	hasher.Write(chunk)
	hash := hasher.Sum(nil)
	return hash
}

// StringHash hashes the given data and returns the result as a
// hex encoded byte string.
func (h *Hasher) StringHash(chunk []byte) string {
	return hex.EncodeToString(h.Hash(chunk))
}
