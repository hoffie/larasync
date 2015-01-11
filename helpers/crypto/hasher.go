package crypto

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
)

const (
	HashingKeySize = 32
)

type Hasher struct {
	hashingKey [HashingKeySize]byte
}

func NewHasher(hashingKey [HashingKeySize]byte) *Hasher {
	return &Hasher{
		hashingKey: hashingKey,
	}
}

func (h *Hasher) Hash(chunk []byte) []byte {
	key := h.hashingKey
	hasher := hmac.New(sha512.New, key[:])
	hasher.Write(chunk)
	hash := hasher.Sum(nil)
	return hash
}

func (h *Hasher) StringHash(chunk []byte) string {
	return hex.EncodeToString(h.Hash(chunk))
}
