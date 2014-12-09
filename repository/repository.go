package repository

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

const (
	authkeyFilename   = "auth.pub"
	managementDirname = ".lara"
	defaultFilePerms  = 0600
	defaultDirPerms   = 0700
)

// Repository represents an on-disk repository and provides methods to
// access its sub-items.
type Repository struct {
	Path string
}

// New returns a new repository instance with the given base path
func New(path string) *Repository {
	return &Repository{Path: path}
}

// CreateManagementDir ensures that this repository's management
// directory exists.
func (r *Repository) CreateManagementDir() error {
	path := filepath.Join(r.Path, managementDirname)
	err := os.Mkdir(path, defaultDirPerms)
	if err != nil && err != os.ErrExist {
		return err
	}
	return nil
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
	return filepath.Join(r.Path, managementDirname, authkeyFilename)
}

// GetAuthPubkey returns the repository auth key's public key.
func (r *Repository) GetAuthPubkey() ([]byte, error) {
	pubkey, err := ioutil.ReadFile(r.getAuthPubkeyPath())
	return pubkey, err
}

// SetAuthPubkey sets the repository auth key's public key.
func (r *Repository) SetAuthPubkey(key []byte) error {
	return ioutil.WriteFile(r.getAuthPubkeyPath(), key,
		defaultFilePerms)
}
