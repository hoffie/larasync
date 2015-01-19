package repository

import (
	"crypto/rand"
	"errors"
	"fmt"

	"github.com/hoffie/larasync/helpers/crypto"
	edhelpers "github.com/hoffie/larasync/helpers/ed25519"
	"github.com/hoffie/larasync/repository/content"
)

const (
	// Key Sizes

	// PrivateKeySize is the keySize which is used for signatures in the
	// system.
	PrivateKeySize = crypto.PrivateKeySize
	// PublicKeySize is the keySize for public signature keys.
	PublicKeySize = crypto.PublicKeySize
	// EncryptionKeySize represents the size of the key used for
	// encrypting.
	EncryptionKeySize = crypto.EncryptionKeySize
	// HashingKeySize represents the size of the key used for
	// generating content hashes (HMAC).
	HashingKeySize = crypto.HashingKeySize

	// ids for our keys in the storage
	encryptionKeyName     = "encryption.key"
	hashingKeyName        = "hashing.key"
	signingPrivateKeyName = "signing.priv"
	signingPublicKeyName  = "signing.pub"
)

// KeyStore is responsible for loading keys from the storage backend.
type KeyStore struct {
	base    string
	storage *content.ByteStorage
}

// NewKeyStore returns a new KeyStore instance.
func NewKeyStore(storage content.Storage) *KeyStore {
	ks := &KeyStore{storage: content.NewByteStorage(storage)}
	return ks
}

// SetEncryptionKey sets the encryption key
func (ks *KeyStore) SetEncryptionKey(key [EncryptionKeySize]byte) error {
	return ks.storage.SetBytes(encryptionKeyName, key[:])
}

// EncryptionKey returns the encryption key.
func (ks *KeyStore) EncryptionKey() ([EncryptionKeySize]byte, error) {
	key, err := ks.storage.GetBytes(encryptionKeyName)
	if len(key) != EncryptionKeySize {
		return [EncryptionKeySize]byte{}, fmt.Errorf(
			"invalid key length (%d)", len(key))
	}
	var arrKey [EncryptionKeySize]byte
	copy(arrKey[:], key)
	return arrKey, err
}

// SetSigningPrivateKey sets the signing private key
func (ks *KeyStore) SetSigningPrivateKey(key [PrivateKeySize]byte) error {
	return ks.storage.SetBytes(signingPrivateKeyName, key[:])
}

// SigningPrivateKey returns the signing private key.
func (ks *KeyStore) SigningPrivateKey() ([PrivateKeySize]byte, error) {
	key, err := ks.storage.GetBytes(signingPrivateKeyName)
	if len(key) != PrivateKeySize {
		return [PrivateKeySize]byte{}, fmt.Errorf(
			"invalid key length (%d)", len(key))
	}
	var arrKey [PrivateKeySize]byte
	copy(arrKey[:], key)
	return arrKey, err
}

// SetSigningPublicKey sets the signing key's public key.
func (ks *KeyStore) SetSigningPublicKey(key []byte) error {
	return ks.storage.SetBytes(signingPublicKeyName, key)
}

// SigningPublicKey returns the signing public key.
func (ks *KeyStore) SigningPublicKey() ([PublicKeySize]byte, error) {
	privKey, err := ks.SigningPrivateKey()
	if err != nil {
		return ks.signingPublicKeyFromStorage()
	}
	return edhelpers.GetPublicKeyFromPrivate(privKey), nil
}

// signingPubkeyFromStorage returns the repository signing public key.
//
// It tries to retrieve the stored copy and is only called if the public key
// cannot be derived from the private key (i.e. if the private key is not
// available in this repository).
func (ks *KeyStore) signingPublicKeyFromStorage() ([PublicKeySize]byte, error) {
	key, err := ks.storage.GetBytes(signingPublicKeyName)
	if len(key) != PublicKeySize {
		return [PublicKeySize]byte{}, fmt.Errorf(
			"invalid key length (%d)", len(key))
	}
	var arrKey [PublicKeySize]byte
	copy(arrKey[:], key)
	return arrKey, err
}

// SetHashingKey sets the repository hashing key (content addressing)
func (ks *KeyStore) SetHashingKey(key [HashingKeySize]byte) error {
	return ks.storage.SetBytes(hashingKeyName, key[:])
}

// HashingKey returns the repository signing private key.
func (ks *KeyStore) HashingKey() ([HashingKeySize]byte, error) {
	key, err := ks.storage.GetBytes(hashingKeyName)
	if len(key) != HashingKeySize {
		return [HashingKeySize]byte{}, fmt.Errorf(
			"invalid key length (%d)", len(key))
	}
	var arrKey [HashingKeySize]byte
	copy(arrKey[:], key)
	return arrKey, err
}

// CreateEncryptionKey generates a random encryption key.
func (ks *KeyStore) CreateEncryptionKey() error {
	key := make([]byte, EncryptionKeySize)
	_, err := rand.Read(key)
	if err != nil {
		return err
	}
	err = ks.storage.SetBytes(encryptionKeyName, key)
	return err
}

// CreateSigningKey generates a random signing key.
func (ks *KeyStore) CreateSigningKey() error {
	_, privKey, err := edhelpers.GenerateKey()
	if err != nil {
		return err
	}
	if privKey == nil {
		return errors.New("no private key generated")
	}
	err = ks.SetSigningPrivateKey(*privKey)
	return err
}

// CreateHashingKey generates a random hashing key.
func (ks *KeyStore) CreateHashingKey() error {
	key := make([]byte, HashingKeySize)
	var arrKey [HashingKeySize]byte
	_, err := rand.Read(key)
	if err != nil {
		return err
	}
	copy(arrKey[:], key)
	err = ks.SetHashingKey(arrKey)
	return err
}
