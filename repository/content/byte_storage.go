package content

import (
	"bytes"
	"io/ioutil"
)

// ByteStorage wraps a Storage and provides it with
// SetBytes and GetBytes methods
type ByteStorage struct {
	Storage
}

// NewByteStorage returns a new ByteStorage instance,
// wrapping the given Storage.
func NewByteStorage(s Storage) *ByteStorage {
	return &ByteStorage{Storage: s}
}

// GetBytes returns the data stored under the given id as bytes.
func (s *ByteStorage) GetBytes(id string) ([]byte, error) {
	r, err := s.Storage.Get(id)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// SetBytes initially sets or updates the bytes for the given id.
func (s *ByteStorage) SetBytes(id string, data []byte) error {
	r := bytes.NewReader(data)
	return s.Storage.Set(id, r)
}
