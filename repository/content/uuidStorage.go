package content

import (
	"crypto/rand"
	"crypto/sha512"
	"encoding/hex"
	"os"
)

// UUIDStorage is a base struct which implements functionality
// to find new UUIDs for the system.
type UUIDStorage struct {
	Storage
}

// NewUUIDStorage wraps a given Storage object and returns a
// UUIDStorage object.
func NewUUIDStorage(storage Storage) *UUIDStorage {
	return &UUIDStorage{
		Storage: storage,
	}
}

// FindFreeUUID generates a new UUID; it tries to avoid
// local collisions.
func (m *UUIDStorage) FindFreeUUID() ([]byte, error) {
	hostname := os.Getenv("HOSTNAME")
	rnd := make([]byte, 32)
	for {
		_, err := rand.Read(rnd)
		if err != nil {
			return nil, err
		}
		hasher := sha512.New()
		hasher.Write([]byte(hostname))
		hasher.Write(rnd)
		hash := hasher.Sum(nil)
		hasUUID := m.HasUUID(hash)
		if err != nil {
			return nil, err
		}
		if !hasUUID {
			return hash, nil
		}
	}
}

// HasUUID checks if the given UUID is already in use in this repository;
// this is a local-only check.
func (m *UUIDStorage) HasUUID(hash []byte) bool {
	UUID := hex.EncodeToString(hash)
	return m.Exists(UUID)
}
