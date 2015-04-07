package path

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

// CleanUpEmptyDirs is used to search all sub directories and removes
// the directories which are empty.
func CleanUpEmptyDirs(absPath string) error {
	stat, err := os.Stat(absPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	} else if os.IsNotExist(err) || !stat.IsDir() {
		return nil
	}

	fileItems, err := ioutil.ReadDir(absPath)
	if err != nil {
		return err
	}

	for _, fileItem := range fileItems {
		if fileItem.IsDir() {
			err = CleanUpEmptyDirs(filepath.Join(absPath, fileItem.Name()))
			if err != nil {
				return err
			}
		}
	}

	// Intentionally not checking for an error. It is fine if it fails
	// this just means that the directory is not empty and has files in it
	// (or has subdirectories which have files in it).
	os.Remove(absPath)

	return nil
}
