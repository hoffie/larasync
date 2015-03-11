package path

import (
	"syscall"
)

// getFileAttributes returns the file Attributes for the
// given path.
func getFileAttributes(path string) (uint32, error) {
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return 0, err
	}
	return syscall.GetFileAttributes(p)
}

// IsHidden returns if the given path is marked as hidden in windows.
func IsHidden(path string) (bool, error) {
	attrs, err := getFileAttributes(path)
	if err != nil {
		return false, err
	}
	return attrs&syscall.FILE_ATTRIBUTE_HIDDEN != 0, nil
}

// Hide tries to hide the given directory and returns an error
// if failed.
func Hide(path string) error {
	attrs, err := getFileAttributes(path)
	if err != nil {
		return err
	}

	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return err
	}

	attrs = attrs | syscall.FILE_ATTRIBUTE_HIDDEN
	return syscall.SetFileAttributes(p, attrs)
}
