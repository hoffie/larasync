package repository

import (
	"errors"
	"io"
)

var (
	// ErrSignatureVerification gets returned if a signature of a signed NIB could
	// not be verified.
	ErrSignatureVerification = errors.New("Signature verification failed")
	// ErrUnMarshalling gets returned if a NIB could not be extracted from stored
	// bytes.
	ErrUnMarshalling = errors.New("Couldn't extract item from byte stream")
)

// NIBStore represents an interface which can be used
// to access NIB information in a repository.
type NIBStore interface {
	// Add adds the given NIB to the store.
	Add(nib *NIB) error
	// AddContent adds the given data in the reader to the
	// store. This should be only used if the NIB object
	// is not available.
	AddContent(UUID string, reader io.Reader) error
	// Get returns the NIB of the given uuid.
	Get(UUID string) (*NIB, error)
	// GetBytes returns the Byte representation of the
	// given NIB UUID.
	GetBytes(UUID string) ([]byte, error)
	// GetReader returns the Reader which stores the bytes
	// of the given NIB UUID.
	GetReader(UUID string) (io.Reader, error)
	// GetAll returns all NIBs in the given store in the order added to
	// the store.
	GetAll() (<-chan *NIB, error)
	// GetFrom returns all NIBs added after the given transaction ID.
	GetFrom(fromTransactionId int64) (<-chan *NIB, error)
	// Exists returns if there is a NIB with
	// the given UUID in the store.
	Exists(UUID string) bool
	// VerifyContent verifies the correctness of the given
	// data in the reader.
	VerifyContent(reader io.Reader) error
}
