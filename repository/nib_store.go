package repository

// NIBStore represents an interface which can be used
// to access NIB information in a repository.
type NIBStore interface {
	// Add adds the given NIB to the store.
	Add(nib *NIB) error
	// Get returns the NIB of the given uuid.
	Get(UUID string) (*NIB, error)
	// Exists returns if there is a NIB with
	// the given UUID in the store.
	Exists(UUID string) bool
}
