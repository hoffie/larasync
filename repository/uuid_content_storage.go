package repository

import (
	"crypto/rand"
	"crypto/sha512"
	"os"
)

// UUIDContentStorage is a base struct which implements functionality
// to find new UUIDs for the system.
type UUIDContentStorage struct {
	ContentStorage
}

// findFreeUUID generates a new UUID; it tries to avoid
// local collisions.
func (m *UUIDContentStorage) findFreeUUID() ([]byte, error) {
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
		hasUUID, err := m.hasUUID(hash)
		if err != nil {
			return nil, err
		}
		if !hasUUID {
			return hash, nil
		}
	}
}

// hasUUID checks if the given UUID is already in use in this repository;
// this is a local-only check.
func (m *UUIDContentStorage) hasUUID(hash []byte) (bool, error) {
	UUID := formatUUID(hash)
	return m.Exists(UUID), nil
}
