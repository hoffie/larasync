// +build windows

package repository

import (
	"fmt"

	"github.com/hoffie/larasync/helpers/path"
)

// hideManagementDirectory tries to mark the directory as hidden.
// If this fails returns an error.
func (md *managementDirectory) hideManagementDirectory() error {
	return path.Hide(md.getDir())
}

// afterRootInitialization can be used as a hook for system
// specific functionality which is needed after the Root
// management directory has been initialized.
// On windows this tries to mark the management directory
// folder as hidden. This will however fail silently if it
// does not work. It is not necessary for the overall system
// functionality for this to work.
func (md *managementDirectory) afterRootInitialization() error {
	err := md.hideManagementDirectory()
	// Explicitly ignoring the error code here. It is ok if it
	// fails. Just log a warning.
	if err != nil {
		Log.Warn(
			fmt.Sprintf(
				"Error while trying to hide management directory. %s",
				err.Error(),
			),
		)
	}
	return nil
}
