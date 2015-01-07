package repository

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
)

// ClientNIBStore implements the NIBStore interface from the
// client perspective.
type ClientNIBStore struct {
	storage    *UUIDContentStorage
	repository *Repository
}

// newClientNIBStore generates the clientNibStore with the given data
// and returns the new entry.
func newClientNIBStore(storage *ContentStorage, repository *Repository) *ClientNIBStore {
	uuidStorage := UUIDContentStorage{*storage}
	return &ClientNIBStore{
		storage:    &uuidStorage,
		repository: repository}
}

// Get returns the NIB of the given uuid.
func (s ClientNIBStore) Get(UUID string) (*NIB, error) {
	pubKey, err := s.repository.GetSigningPubkey()
	if err != nil {
		return nil, err
	}

	data, err := s.GetBytes(UUID)
	if err != nil {
		return nil, err
	}

	buffer := bytes.NewReader(data)
	signatureReader, err := NewVerifyingReader(
		pubKey,
		buffer,
	)
	if err != nil {
		return nil, err
	}

	nib := NIB{}
	_, err = nib.ReadFrom(signatureReader)
	if err != nil {
		return nil, err
	}

	if !signatureReader.Verify() {
		return nil, fmt.Errorf("Signature of NIB %s invalid", UUID)
	}

	return &nib, nil
}

// GetBytes returns the Byte representation of the
// given NIB UUID.
func (s ClientNIBStore) GetBytes(UUID string) ([]byte, error) {
	reader, err := s.GetReader(UUID)
	if err != nil {
		return []byte{}, err
	}
	return ioutil.ReadAll(reader)
}

// GetReader returns the Reader which stores the bytes
// of the given NIB UUID.
func (s ClientNIBStore) GetReader(UUID string) (io.Reader, error) {
	return s.storage.Get(UUID)
}

// Add adds the given NIB to the store.
func (s ClientNIBStore) Add(nib *NIB) error {
	// Empty UUID. Generating new one.
	if nib.UUID == "" {
		uuid, err := s.storage.findFreeUUID()
		if err != nil {
			return err
		}
		nib.UUID = formatUUID(uuid)
	}

	buf := &bytes.Buffer{}
	_, err := nib.WriteTo(buf)
	if err != nil {
		return err
	}

	return s.writeBytes(nib.UUID, buf.Bytes())
}

// writeBytes signs and adds the bytes for the given NIB UUID.
func (s ClientNIBStore) writeBytes(UUID string, data []byte) error {
	key, err := s.repository.GetSigningPrivkey()

	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}

	sw := NewSigningWriter(key, buf)
	_, err = sw.Write(data)
	if err != nil {
		return err
	}
	err = sw.Finalize()
	if err != nil {
		return err
	}

	return s.storage.Set(UUID, buf)
}

// Exists returns if there is a NIB with
// the given UUID in the store.
func (s ClientNIBStore) Exists(UUID string) bool {
	return s.storage.Exists(UUID)
}
