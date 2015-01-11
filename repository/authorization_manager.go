package repository

import (
	"bytes"
	"encoding/hex"
	"errors"
	"io"
	"io/ioutil"

	"github.com/hoffie/larasync/helpers/crypto"
)

var (
	// ErrInvalidPublicKeySize will get thrown if a string is passed
	// which couldn't be encoded to the correct size to pass it as a
	// Public Key signature.
	ErrInvalidPublicKeySize = errors.New("Invalid public key size.")
)

// AuthorizationManager handles the Authorizations of a specific
//
type AuthorizationManager struct {
	storage ContentStorage
}

func newAuthorizationManager(storage ContentStorage) *AuthorizationManager {
	return &AuthorizationManager{
		storage: storage,
	}
}

// Set adds the authorization to the given backend and encrypts it
// first.
func (am *AuthorizationManager) Set(
	signaturePubKey [PublicKeySize]byte,
	encryptionKey [EncryptionKeySize]byte,
	authorization *Authorization,
) error {

	data := &bytes.Buffer{}
	_, err := authorization.WriteTo(data)
	if err != nil {
		return err
	}

	box := crypto.NewBox(encryptionKey)
	enc, err := box.EncryptWithRandomKey(data.Bytes())
	if err != nil {
		return err
	}

	return am.SetData(signaturePubKey, bytes.NewReader(enc))
}

// SetData adds for the already encrypted byte data and the given public key
// to the storage backend.
func (am *AuthorizationManager) SetData(
	pubKey [PublicKeySize]byte,
	reader io.Reader,
) error {
	pubKeyString := hex.EncodeToString(pubKey[:])
	return am.storage.Set(pubKeyString, reader)
}

// GetReaderString returns the reader for the given publicKey string representation.
func (am *AuthorizationManager) GetReaderString(key string) (io.ReadCloser, error) {
	byteKey, err := hex.DecodeString(key)
	if err != nil {
		return nil, err
	}

	if len(byteKey) != PublicKeySize {
		return nil, ErrInvalidPublicKeySize
	}
	inputKey := [PublicKeySize]byte{}
	copy(inputKey[:], byteKey)

	return am.GetReader(inputKey)
}

// GetReader returns a reader for the authorization stored with the passed PublicKey.
func (am *AuthorizationManager) GetReader(key [PublicKeySize]byte) (io.ReadCloser, error) {
	publicKeyString := hex.EncodeToString(key[:])
	return am.storage.Get(publicKeyString)
}

// Get returns the Authorization for the given public Signature Key and
// tries to decrypt it with the passed encryptionKey.
func (am *AuthorizationManager) Get(
	signaturePubKey [PublicKeySize]byte,
	encryptionKey [EncryptionKeySize]byte,
) (*Authorization, error) {
	reader, err := am.GetReader(signaturePubKey)
	if err != nil {
		return nil, err
	}

	enc, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	box := crypto.NewBox(encryptionKey)
	data, err := box.DecryptContent(enc)
	if err != nil {
		return nil, err
	}

	auth := &Authorization{}
	_, err = auth.ReadFrom(bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	_ = reader.Close()

	return auth, nil
}

// ExistsForString returns if there is a authorization stored
// for a given publicKey string representation.
func (am *AuthorizationManager) ExistsForString(publicKey string) bool {
	return am.storage.Exists(publicKey)
}

// Exists returns if there is a key existing for the given
// publicKey.
func (am *AuthorizationManager) Exists(key [PublicKeySize]byte) bool {
	keyString := hex.EncodeToString(key[:])
	return am.ExistsForString(keyString)
}

// DeleteForString deletes the authorization which is stored for the
// given publicKey string representation.
func (am *AuthorizationManager) DeleteForString(publicKey string) error {
	return am.storage.Delete(publicKey)
}

// Delete removes the authorization which is stored for the signature
// which has the given PublicKey.
func (am *AuthorizationManager) Delete(key [PublicKeySize]byte) error {
	keyString := hex.EncodeToString(key[:])
	return am.DeleteForString(keyString)
}
