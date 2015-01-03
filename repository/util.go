package repository

import (
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
)

// GetRoot returns the repository root of the given path.
func GetRoot(path string) (string, error) {
	prevPath := path
	for {
		if isRoot(path) {
			return path, nil
		}
		prevPath = path
		path = filepath.Dir(path)
		if path == prevPath {
			break
		}
	}
	return "", errors.New("unable to find repository root")
}

// isRoot checks whether the given path is the root of a repository.
func isRoot(path string) bool {
	mgmtPath := filepath.Join(path, managementDirName)
	s, err := os.Stat(mgmtPath)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// formatUUID converts a binary UUID to the readable string representation
func formatUUID(uuid []byte) string {
	return hex.EncodeToString(uuid)
}
