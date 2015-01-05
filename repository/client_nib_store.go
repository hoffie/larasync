package repository

import (
	"bytes"
)

// ClientNIBStore implements the NIBStore interface from the
// client perspective.
type ClientNIBStore struct {
	contentStorage ContentStorage
	repository     Repository
}

// Get returns the NIB of the given uuid.
func (f ClientNIBStore) Get(UUID string) (*NIB, error) {
	reader, err := f.contentStorage.Get(UUID)

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
	return f.contentStorage.Set(UUID, buf)
}

// Exists returns if there is a NIB with
// the given UUID in the store.
func (f ClientNIBStore) Exists(UUID string) bool {
	return f.contentStorage.Exists(UUID)
}
