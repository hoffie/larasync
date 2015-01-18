// +build windows

package atomic

import (
	"os"
)

var writerErrorHook error

// closeHook is a hook which can be implemented by platform specific code
// to inject custom code.
func (aw *Writer) closeHook() error {
	// On windows you can not move a file on an already existing one.
	// This is however expected behaviour in the application. Thus the necessity
	// to remove the item in Windows first.

	_, err = os.Stat(aw.path)
	if err == nil {
		err = os.Remove(aw.path)
		if err != nil {
			return err
		}
	}
	return writerErrorHook
}

// initFileHook gets called to do platform specific functionality for an
// initialized file.
func (aw *Writer) initFileHook(f *os.File) error {
	// Chmod not supported on windows.
	return writerErrorHook
}
