package path

import (
	"os"
	"path/filepath"
	"strings"
)

// Normalize resolves symlinks, makes the path absolute and removes unnecessary
// chars from it, making it as unique as possible
func Normalize(p string) (string, error) {
	p, err := filepath.EvalSymlinks(p)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	p, err = filepath.Abs(p)
	if err != nil {
		return "", err
	}
	p = filepath.Clean(p)
	return p, nil
}

// IsBelow tries hard to decide whether the given path is rooted at the given
// second argument or below.
func IsBelow(p, below string) (bool, error) {
	p, err := Normalize(p)
	if err != nil {
		return false, err
	}
	below, err = Normalize(below)
	if err != nil {
		return false, err
	}
	if strings.HasPrefix(p, below) {
		return true, nil
	}
	return false, nil
}
