package repository

import (
	"bytes"
	"io/ioutil"
)

// ByteContentStorage wraps a ContentStorage and provides it with
// SetBytes and GetBytes methods
type ByteContentStorage struct {
	ContentStorage
}

// newByteContentStorage returns a new ByteContentStorage instance,
// wrapping the given ContentStorage.
func newByteContentStorage(s ContentStorage) *ByteContentStorage {
	return &ByteContentStorage{ContentStorage: s}
}

// GetBytes returns the data stored under the given id as bytes.
func (s *ByteContentStorage) GetBytes(id string) ([]byte, error) {
	r, err := s.ContentStorage.Get(id)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// SetBytes initially sets or updates the bytes for the given id.
func (s *ByteContentStorage) SetBytes(id string, data []byte) error {
	r := bytes.NewReader(data)
	return s.ContentStorage.Set(id, r)
}
