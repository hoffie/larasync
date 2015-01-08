package lock

import (
	"sync"
)

// Manager interface which can be used
// to request locks based on a role and
// a path.
type Manager interface {
	// Get returns a unique Lock for the given path and role.
	// Calling this function again with the same input parameters
	// has to return the same lock.
	Get(path string, role string) sync.Locker
}
