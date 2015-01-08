package lock

import (
	"sync"
)

// Manager interface which can be used
// to request locks based on a role and
// a path.
type Manager interface {
	// Get returns for the given path and the given
	// role the allocated Locker.
	Get(path string, role string) sync.Locker
}
