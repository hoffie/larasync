package repository

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	authPubkeyFileName    = "auth.pub"
	encryptionKeyFileName = "encryption.key"
	managementDirName     = ".lara"
	blobDirName           = "blobs"
	defaultFilePerms      = 0600
	defaultDirPerms       = 0700
)

// Repository represents an on-disk repository and provides methods to
// access its sub-items.
type Repository struct {
	Path    string
	storage BlobStorage
}

// New returns a new repository instance with the given base path
func New(path string) *Repository {
	return &Repository{Path: path}
}

// getStorage returns the currently configured blob storage backend
// for the repository.
func (r *Repository) getStorage() (*BlobStorage, error) {
	if r.storage == nil {
		storage := FileBlobStorage{
			StoragePath: filepath.Join(
				r.GetManagementDir(),
				blobDirName)}
		err := storage.CreateDir()
		if err != nil {
			return nil, err
		}
		r.storage = storage
	}
	return &r.storage, nil
}

// CreateManagementDir ensures that this repository's management
// directory exists.
func (r *Repository) CreateManagementDir() error {
	err := os.Mkdir(r.GetManagementDir(), defaultDirPerms)
	if err != nil && err != os.ErrExist {
		return err
	}
	return nil
}

// GetManagementDir returns the path to the management directory.
func (r *Repository) GetManagementDir() string {
	return filepath.Join(r.Path, managementDirName)
}

// Create initially creates the repository directory structure.
func (r *Repository) Create() error {
	err := os.Mkdir(r.Path, defaultDirPerms)
	if err != nil {
		return err
	}
	err = r.CreateManagementDir()
	return err
}

// getAuthPubkeyPath returns the path of the repository's auth pubkey
// storage location.
func (r *Repository) getAuthPubkeyPath() string {
	return filepath.Join(r.Path, managementDirName, authPubkeyFileName)
}

// getEncryptionKeyPath returns the path of the repository's encryption key
// storage location.
func (r *Repository) getEncryptionKeyPath() string {
	return filepath.Join(r.Path, managementDirName, encryptionKeyFileName)
}

// GetAuthPubkey returns the repository auth key's public key.
func (r *Repository) GetAuthPubkey() ([]byte, error) {
	pubkey, err := ioutil.ReadFile(r.getAuthPubkeyPath())
	return pubkey, err
}

// SetAuthPubkey sets the repository auth key's public key.
func (r *Repository) SetAuthPubkey(key []byte) error {
	return ioutil.WriteFile(r.getAuthPubkeyPath(), key, defaultFilePerms)
}

// SetEncryptionKey sets the repository encryption key
func (r *Repository) SetEncryptionKey(key []byte) error {
	return ioutil.WriteFile(r.getEncryptionKeyPath(), key, defaultFilePerms)
}

// GetEncryptionKey returns the repository encryption key.
func (r *Repository) GetEncryptionKey() ([]byte, error) {
	key, err := ioutil.ReadFile(r.getEncryptionKeyPath())
	return key, err
}

// AddBlob adds a blob into the storage with the given
// id and adds the data in the reader to it.
func (r *Repository) AddBlob(blobID string, data io.Reader) error {
	return r.storage.Set(blobID, data)
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
