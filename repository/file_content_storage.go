package repository

import (
	"io"
	"io/ioutil"
	"os"
	"path"
)

// FileContentStorage is the basic implementation of the ContentStorage
// implementation which stores the data into the file system.
type FileContentStorage struct {
	StoragePath string
}

// CreateDir ensures that the file blob storage directory exists.
func (f *FileContentStorage) CreateDir() error {
	err := os.Mkdir(f.StoragePath, defaultDirPerms)

	if err != nil && !os.IsExist(err) {
		return err
	}
	return nil
}

// storagePathFor returns the storage path for the data entry.
func (f *FileContentStorage) storagePathFor(contentID string) string {
	return path.Join(f.StoragePath, contentID)
}

// Get returns the file handle for the given contentID.
// If there is no data stored for the Id it should return a
// os.ErrNotExists error.
func (f FileContentStorage) Get(contentID string) (io.Reader, error) {
	if f.Exists(contentID) {
		return os.Open(f.storagePathFor(contentID))
	}
	return nil, os.ErrNotExist
}

// Set sets the data of the given contentID in the blob storage.
func (f FileContentStorage) Set(contentID string, reader io.Reader) error {
	blobStoragePath := f.storagePathFor(contentID)
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

// Exists checks if the given entry is stored in the database.
func (f FileContentStorage) Exists(contentID string) bool {
	_, err := os.Stat(f.storagePathFor(contentID))
	if err != nil {
		return !os.IsNotExist(err)
	}
	return true
}
