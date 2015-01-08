package lock

var currentManager Manager

// CurrentManager returns the currently active Lock Manager for
// this system which can be used to query locks for specific
// paths and roles.
func CurrentManager() Manager {
	if currentManager == nil {
		currentManager = newProcessManager()
	}
	return currentManager
}
