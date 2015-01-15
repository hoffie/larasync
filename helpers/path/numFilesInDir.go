package path

import (
	"io/ioutil"
)

// NumFilesInDir returns the number of files in the given
// directory.
func NumFilesInDir(path string) (int, error) {
	entries, err := ioutil.ReadDir(path)
	if err != nil {
		return 0, err
	}
	return len(entries), nil
}
