package repository

import (
	"io"
)

// NIBStore represents an interface which can be used
// to access NIB information in a repository.
type NIBStore interface {
	// Add adds the given NIB to the store.
	Add(nib *NIB) error
	// Get returns the NIB of the given uuid.
	Get(UUID string) (*NIB, error)
	// GetBytes returns the Byte representation of the
	// given NIB UUID.
	GetBytes(UUID string) ([]byte, error)
	// GetReader returns the Reader which stores the bytes
	// of the given NIB UUID.
	GetReader(UUID string) (io.Reader, error)
	// Exists returns if there is a NIB with
	// the given UUID in the store.
	Exists(UUID string) bool
}
