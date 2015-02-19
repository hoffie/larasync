package tracker

import (
	"os"
	"path/filepath"
)

// NewNIBSearchResponse returns a initialized NIBSearchResponse struct with the
// given parameters.
func NewNIBSearchResponse(NIBID string, path string, repositoryPath string) *NIBSearchResponse {
	return &NIBSearchResponse{
		NIBID:          NIBID,
		Path:           path,
		repositoryPath: repositoryPath,
	}
}

// NIBSearchResponse is being returned by NIBTracker implementations
// to indicate that a NIB for the path exists.
type NIBSearchResponse struct {
	NIBID          string
	Path           string
	repositoryPath string
}

// AbsPath returns the absolute path the stored NIBID is being
// associated to.
func (r *NIBSearchResponse) AbsPath() string {
	path := filepath.Join(r.repositoryPath, r.Path)
	absPath, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	symlinkResolve, err := filepath.EvalSymlinks(absPath)
	if err != nil {
		return absPath
	}
	return symlinkResolve
}

// FileExists returns if the path still has a file stored.
func (r *NIBSearchResponse) FileExists() bool {
	stat, err := os.Stat(r.AbsPath())
	if err != nil || stat.IsDir() {
		return false
	}
	return true
}
