package repository

import (
	"io"
)

// BlobStorage is the generic interface for implementations of
// Backends which store blob data.
type BlobStorage interface {
	// Get returns the file handle for the given blobId
	Get(blobID string) (io.Reader, error)
	// Set sets the data of the given blobId in the blob storage
	Set(blobID string, reader io.Reader) error
	// Exists
	Exists(blobID string) bool
}
