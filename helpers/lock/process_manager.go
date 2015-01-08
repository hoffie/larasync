package lock

import (
	"sync"
)

// ProcessManager implements the Manager interface and
// can be used to request locks on a process level.
type ProcessManager struct {
	locks map[string]map[string]sync.Locker
}

func newProcessManager() *ProcessManager {
	pm := &ProcessManager{}
	pm.reset()
	return pm
}

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

func (pm *ProcessManager) reset() {
	pm.locks = map[string]map[string]sync.Locker{}
}
