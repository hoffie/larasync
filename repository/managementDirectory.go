package repository

import (
	"os"
	"path/filepath"

	"github.com/hoffie/larasync/repository/content"
)

// newManagementDirectory returns the struct which can be used
// to interact with generic management directory functionality.
func newManagementDirectory(r *Repository) *managementDirectory {
	return &managementDirectory{
		r: r,
	}
}

type managementDirectory struct {
	r *Repository
}

// subPathFor returns the full path for the given subdirectory.
func (md *managementDirectory) subPathFor(name string) string {
	return filepath.Join(md.getDir(), name)
}

// getDir returns the directory of this management directory.
func (md *managementDirectory) getDir() string {
	return filepath.Join(md.r.Path, managementDirName)
}

// create ensures that this repository's management
// directory exists.
func (md *managementDirectory) create() error {
	err := os.Mkdir(md.getDir(), defaultDirPerms)
	if err != nil && !os.IsExist(err) {
		return err
	}
	err = md.afterRootInitialization()

	storages := []*content.FileStorage{
		content.NewFileStorage(md.subPathFor(authorizationsDirName)),
		content.NewFileStorage(md.subPathFor(nibsDirName)),
		content.NewFileStorage(md.subPathFor(transactionsDirName)),
		content.NewFileStorage(md.subPathFor(objectsDirName)),
		content.NewFileStorage(md.subPathFor(keysDirName)),
	}

	for _, fileStorage := range storages {
		err = fileStorage.CreateDir()
		if err != nil {
			return err
		}
	}

	return nil
}
