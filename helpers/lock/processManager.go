package lock

import (
	"sync"
)

// ProcessManager implements the Manager interface and
// can be used to request locks on a process level.
type ProcessManager struct {
	locks map[string]map[string]sync.Locker
}

// newProcessManager is a helper function which initializes
// a process manager and returns it.
func newProcessManager() *ProcessManager {
	pm := &ProcessManager{}
	pm.reset()
	return pm
}

// Get returns a unique Lock for the given path and role.
// Calling this function again with the same input parameters
// will return the same lock.
func (pm *ProcessManager) Get(path string, role string) sync.Locker {
	roleMap, ok := pm.locks[path]
	if !ok {
		roleMap = map[string]sync.Locker{}
		pm.locks[path] = roleMap
	}

	locker, ok := roleMap[role]
	if !ok {
		locker = &sync.Mutex{}
		roleMap[role] = locker
	}
	return locker
}

// reset is an internal helper function which is used for testing purposes
// and on object initialisation. It purges the lock cache.
func (pm *ProcessManager) reset() {
	pm.locks = map[string]map[string]sync.Locker{}
}
