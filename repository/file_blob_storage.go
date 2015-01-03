package repository

import (
	"errors"
	"io"
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
	if err != nil && err != os.ErrExist {
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

func (f FileBlobStorage) Set(blobID string, reader io.Reader) error {
	blobStoragePath := f.storagePathFor(blobID)

	file, err := os.Create(blobStoragePath)

	if err != nil {
		return err
	}

	cleanUp := func() {
		file.Close()
		_, err := os.Stat(blobStoragePath)
		if err != nil {
			return
		}
		os.Remove(blobStoragePath)
	}
	_, err = io.Copy(file, reader)
	if err != nil {
		cleanUp()
		return err
	}

	return nil
}

func (f FileBlobStorage) Exists(blobID string) bool {
	_, err := os.Stat(f.storagePathFor(blobID))
	if err != nil {
		return !os.IsNotExist(err)
	}
	return true
}
