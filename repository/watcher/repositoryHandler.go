package watcher

// RepositoryHandler is the interface, which must be implemented
// for objects being passed to the Watcher. This is being
// utilized to add and remove data from the repository
type RepositoryHandler interface {
	// AddItem adds the item to the internal state repository
	// which resides in the passed absolute path.
	AddItem(absPath string) error
	// DeleteItem marks the item accessible through the passed
	// absolute path as deleted in the internal repository state.
	DeleteItem(absPath string) error
}
