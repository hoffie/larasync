package repository

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"os"

	"github.com/hoffie/larasync/repository/content"
	"github.com/hoffie/larasync/repository/nib"
)

const (
	// internal directory names
	managementDirName     = ".lara"
	objectsDirName        = "objects"
	nibsDirName           = "nibs"
	transactionsDirName   = "transactions"
	authorizationsDirName = "authorizations"
	keysDirName           = "keys"
	stateConfigFileName   = "state.json"

	// default permissions
	defaultFilePerms = 0600
	defaultDirPerms  = 0700

	// chunk splitting size
	chunkSize = 1 * 1024 * 1024
)

// Repository represents an on-disk repository and provides methods to
// access its sub-items.
type Repository struct {
	Path                 string
	keys                 *KeyStore
	objectStorage        content.Storage
	nibStore             *NIBStore
	transactionManager   *TransactionManager
	authorizationManager *AuthorizationManager
	managementDir        *managementDirectory
}

// New returns a new repository instance with the given base path
func New(path string) *Repository {
	r := &Repository{Path: path}

    r.managementDir = newManagementDirectory(r)

	r.objectStorage = content.NewFileStorage(r.subPathFor(objectsDirName))

	r.transactionManager = newTransactionManager(
		content.NewFileStorage(r.subPathFor(transactionsDirName)),
		r.managementDir.getDir(),
	)
	r.authorizationManager = newAuthorizationManager(
		content.NewFileStorage(r.subPathFor(authorizationsDirName)),
	)

	r.keys = NewKeyStore(content.NewFileStorage(r.subPathFor(keysDirName)))
	r.nibStore = newNIBStore(
		content.NewFileStorage(r.subPathFor(nibsDirName)),
		r.keys,
		r.transactionManager,
	)

	return r
}

// subPathFor returns the full path for the given entry.
func (r *Repository) subPathFor(name string) string {
	return r.managementDir.subPathFor(name)
}

// CreateManagementDir ensures that this repository's management
// directory exists.
func (r *Repository) CreateManagementDir() error {
    return r.managementDir.create()
}

// GetManagementDir returns the path to the management directory.
func (r *Repository) GetManagementDir() string {
	return r.managementDir.getDir()
}

// Create initially creates the repository directory structure.
func (r *Repository) Create() error {
	err := os.Mkdir(r.Path, defaultDirPerms)
	if err != nil {
		return err
	}
	err = r.CreateManagementDir()
	if err != nil {
		return err
	}

	return err
}

// AddObject adds an object into the storage with the given
// id and adds the data in the reader to it.
func (r *Repository) AddObject(objectID string, data io.Reader) error {
	return r.objectStorage.Set(objectID, data)
}

// HasObject returns whether the given object id exists in the object
// store.
func (r *Repository) HasObject(objectID string) bool {
	return r.objectStorage.Exists(objectID)
}

// VerifyAndParseNIBBytes checks the signature of the given NIB and
// deserializes it if the signature could be validated.
func (r *Repository) VerifyAndParseNIBBytes(data []byte) (*nib.NIB, error) {
	return r.nibStore.VerifyAndParseBytes(data)
}

// AddNIBContent adds NIBData to the repository after verifying it.
func (r *Repository) AddNIBContent(nibReader io.Reader) error {
	nibStore := r.nibStore

	data, err := ioutil.ReadAll(nibReader)
	if err != nil {
		return err
	}

	nib, err := r.VerifyAndParseNIBBytes(data)
	if err != nil {
		return err
	}

	missingObjectIDs := []string{}
	for _, objectID := range nib.AllObjectIDs() {
		if !r.HasObject(objectID) {
			missingObjectIDs = append(missingObjectIDs, objectID)
		}
	}

	if len(missingObjectIDs) > 0 {
		return NewErrNIBContentMissing(missingObjectIDs)
	}

	err = r.ensureConflictFreeNIBImport(nib)
	if err != nil {
		return err
	}
	return nibStore.AddContent(nib.ID, bytes.NewReader(data))
}

// ensureConflictFreeNIBImport returns an error if we cannot import
// the given NIB without conflicts or nil if everything is good.
func (r *Repository) ensureConflictFreeNIBImport(otherNIB *nib.NIB) error {
	if !r.HasNIB(otherNIB.ID) {
		return nil
	}
	myNIB, err := r.GetNIB(otherNIB.ID)
	if err != nil {
		return err
	}
	if myNIB.IsParentOf(otherNIB) {
		return nil
	}
	return ErrNIBConflict
}

// GetNIB returns a NIB for the given ID in this repository.
func (r *Repository) GetNIB(id string) (*nib.NIB, error) {
	return r.nibStore.Get(id)
}

// GetNIBReader returns the NIB with the given id in this repository.
func (r *Repository) GetNIBReader(id string) (io.ReadCloser, error) {
	return r.nibStore.getReader(id)
}

// GetNIBBytesFrom returns the signed byte structure for NIBs from the given
// transaction id
func (r *Repository) GetNIBBytesFrom(fromTransactionID int64) (<-chan []byte, error) {
	return r.nibStore.GetBytesFrom(fromTransactionID)
}

// GetNIBsFrom returns nibs added since the passed transaction ID.
func (r *Repository) GetNIBsFrom(fromTransactionID int64) (<-chan *nib.NIB, error) {
	return r.nibStore.GetFrom(fromTransactionID)
}

// GetAllNIBBytes returns all NIBs signed byte representations in this repository.
func (r *Repository) GetAllNIBBytes() (<-chan []byte, error) {
	return r.nibStore.GetAllBytes()
}

// GetAllNibs returns all the nibs which are stored in this repository.
// Those will be returned with the oldest one first and the newest added
// last.
func (r *Repository) GetAllNibs() (<-chan *nib.NIB, error) {
	return r.nibStore.GetAll()
}

// HasNIB checks if a NIB with the given ID exists in the repository.
func (r *Repository) HasNIB(id string) bool {
	return r.nibStore.Exists(id)
}

// CurrentTransaction returns the currently newest Transaction for this
// repository.
func (r *Repository) CurrentTransaction() (*Transaction, error) {
	return r.transactionManager.CurrentTransaction()
}

// GetAuthorizationReader returns the authorization configuration for the
// passed PublicKey.
func (r *Repository) GetAuthorizationReader(publicKey [PublicKeySize]byte) (io.ReadCloser, error) {
	return r.authorizationManager.GetReader(publicKey)
}

// SetAuthorizationData adds for the given publicKey the authorization structure
func (r *Repository) SetAuthorizationData(publicKey [PublicKeySize]byte, authData io.Reader) error {
	return r.authorizationManager.SetData(publicKey, authData)
}

// DeleteAuthorization removes the authorization with the given publicKey.
func (r *Repository) DeleteAuthorization(publicKey [PublicKeySize]byte) error {
	return r.authorizationManager.Delete(publicKey)
}

// SerializeAuthorization returns the encrypted and authorization which can be passed
// safely to the server.
func (r *Repository) SerializeAuthorization(encryptionKey [EncryptionKeySize]byte,
	authorization *Authorization) ([]byte, error) {
	return r.authorizationManager.Serialize(encryptionKey, authorization)
}

// GetObjectData returns the data stored for the given objectID in this
// repository.
func (r *Repository) GetObjectData(objectID string) (io.ReadCloser, error) {
	return r.objectStorage.Get(objectID)
}

// getRepoRelativePath turns the given path into a path relative to the
// repository root and returns it.
func (r *Repository) getRepoRelativePath(absPath string) (string, error) {
	if len(absPath) < len(r.Path)+1 {
		return "", errors.New("unable to resolve path: path too short")
	}
	rel := absPath[len(r.Path)+1:]
	return rel, nil
}

// GetSigningPublicKey exposes the signing public key as it is required
// in foreign packages such as api.
func (r *Repository) GetSigningPublicKey() ([PublicKeySize]byte, error) {
	return r.keys.SigningPublicKey()
}

// SetKeysFromAuth takes the keys passed through the authorization and puts
// them into the keystore.
func (r *Repository) SetKeysFromAuth(auth *Authorization) error {
	keys := r.keys
	err := keys.SetEncryptionKey(auth.EncryptionKey)
	if err != nil {
		return err
	}
	err = keys.SetHashingKey(auth.HashingKey)
	if err != nil {
		return err
	}
	err = keys.SetSigningPrivateKey(auth.SigningKey)
	if err != nil {
		return err
	}
	return nil
}
