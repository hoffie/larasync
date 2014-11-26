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
	res := make([]string, 0)
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		res = append(res, e.Name())
	}
	return res, nil
}

// Create registers a new repository.
func (m *Manager) Create(name, pubKey string) error {
	err := os.Mkdir(filepath.Join(m.basePath, name), 0700)
	return err
}

// Open returns a handle for the given existing repository.
func (m *Manager) Open(name string) (*Repository, error) {
	r := &Repository{Name: name}
	s, err := os.Stat(filepath.Join(m.basePath, name))
	if err != nil {
		return nil, err
	}
	if !s.IsDir() {
		return nil, errors.New("not a directory")
	}
	return r, nil
}
