package repository

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path"
)

// FileBlobStorage is the basic implementation of the BlobStorage
// implementation which stores the data into the file system.
type FileBlobStorage struct {
	StoragePath string
}

// CreateDir ensures that the file blob storage directory exists.
func (f *FileBlobStorage) CreateDir() error {
	err := os.Mkdir(f.StoragePath, defaultDirPerms)

	if err != nil && !os.IsExist(err) {
		return err
	}
	return nil
}

// storagePathFor returns the storage path for the data entry.
func (f *FileBlobStorage) storagePathFor(blobID string) string {
	return path.Join(f.StoragePath, blobID)
}

// Get returns the file handle for the given blobId
func (f FileBlobStorage) Get(blobID string) (io.Reader, error) {
	if f.Exists(blobID) {
		return os.Open(f.storagePathFor(blobID))
	}
	return nil, errors.New("File does not exist.")
}

// Set adds data from a reader and assigns it to the passed blobID
func (f FileBlobStorage) Set(blobID string, reader io.Reader) error {
	blobStoragePath := f.storagePathFor(blobID)
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(blobStoragePath, data, defaultFilePerms)
	if err != nil {
		return err
	}

	return nil
}

// Exists checks if a blob is stored for the given blobID.
func (f FileBlobStorage) Exists(blobID string) bool {
	_, err := os.Stat(f.storagePathFor(blobID))
	if err != nil {
		return !os.IsNotExist(err)
	}
	return true
}
