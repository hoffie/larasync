package repository

import (
	"bytes"
)

// ClientNIBStore implements the NIBStore interface from the
// client perspective.
type ClientNIBStore struct {
	storage    UUIDContentStorage
	repository Repository
}

// newClientNibStore generates the clientNibStore with the given data
// and returns the new entry.
func newClientNibStore(storage ContentStorage, repository Repository) *ClientNIBStore {
	return &ClientNIBStore{
		storage:    UUIDContentStorage{storage},
		repository: repository}
}

// Get returns the NIB of the given uuid.
func (f ClientNIBStore) Get(UUID string) (*NIB, error) {
	reader, err := f.storage.Get(UUID)

	if err != nil {
		return nil, err
	}

	nib := NIB{}
	_, err = nib.ReadFrom(reader)
	if err != nil {
		return nil, err
	}

	return &nib, nil
}

// Add adds the given NIB to the store.
func (f ClientNIBStore) Add(nib *NIB) error {
	// Empty UUID. Generating new one.
	if nib.UUID == "" {
		uuid, err := f.storage.findFreeUUID()
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

	return f.writeBytes(nib.UUID, buf.Bytes())
}

func (f ClientNIBStore) writeBytes(UUID string, data []byte) error {
	key, err := f.repository.GetSigningPrivkey()

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
	return f.storage.Set(UUID, buf)
}

// Exists returns if there is a NIB with
// the given UUID in the store.
func (f ClientNIBStore) Exists(UUID string) bool {
	return f.storage.Exists(UUID)
}
