package repository

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Manager keeps track of indivudal repositories.
type Manager struct {
	basePath string
}

// NewManager returns a new manager instance.
func NewManager(basePath string) (*Manager, error) {
	stat, err := os.Stat(basePath)
	if err != nil {
		return nil, err
	}
	if !stat.IsDir() {
		return nil, errors.New("not a directory")
	}
	return &Manager{basePath: basePath}, nil
}

// ListNames returns the names of all registered repositories.
func (m *Manager) ListNames() ([]string, error) {
	entries, err := ioutil.ReadDir(m.basePath)
	if err != nil {
		return nil, err
	}
	res := []string{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		res = append(res, e.Name())
	}
	return res, nil
}

// Create registers a new repository.
func (m *Manager) Create(name string, pubKey []byte) error {
	r := &Repository{Path: filepath.Join(m.basePath, name)}
	err := r.Create()
	if err != nil {
		return err
	}
	return r.SetAuthPubkey(pubKey)
}

// Open returns a handle for the given existing repository.
func (m *Manager) Open(name string) (*Repository, error) {
	absPath := filepath.Join(m.basePath, name)
	r := &Repository{Path: absPath}
	s, err := os.Stat(absPath)
	if err != nil {
		return nil, err
	}
	if !s.IsDir() {
		return nil, errors.New("not a directory")
	}
	return r, nil
}

// Exists returns true if the repository with the given name exists.
func (m *Manager) Exists(name string) bool {
	r, _ := m.Open(name)
	return r != nil
}
