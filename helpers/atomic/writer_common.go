// +build !windows

package atomic

import (
	"os"
)

var writerErrorHook error

// closeHook is a hook which can be implemented by platform specific code
// to inject custom code.
func (aw *Writer) closeHook() error {
	return writerErrorHook
}

// initFileHook gets called to do platform specific functionality for an
// initialized file.
func (aw *Writer) initFileHook(f *os.File) error {
	err := f.Chmod(aw.filePerms)
	if err != nil {
		f.Close()
		return err
	}
	return writerErrorHook
}
