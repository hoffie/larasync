package repository

import (
	"io"
)

// ContentStorage is the generic interface for implementations of
// Backends which store blob data.
type ContentStorage interface {
	// Get returns the file handle for the given contentID.
	// If there is no data stored for the Id it should return a
	// os.ErrNotExists error.
	Get(contentID string) (io.ReadCloser, error)
	// Set sets the data of the given contentID in the blob storage.
	Set(contentID string, reader io.Reader) error
	// Exists checks if the given entry is stored in the database.
	Exists(contentID string) bool
	// Delete removes the data with the given contentID from the store.
	Delete(contentID string) error
}
