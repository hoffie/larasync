// +build !windows

package repository

var rootInitializationError error

// afterRootInitialization can be used as a hook for system
// specific functionality which is needed after the Root
// management directory has been initialized.
func (md *managementDirectory) afterRootInitialization() error {
	return rootInitializationError
}
