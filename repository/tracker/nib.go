package tracker

// NIBTracker enables a client repository to have a path
// to nib lookup.
type NIBTracker interface {
	// Add registers the given nibID for the given path.
	Add(path string, nibID string) error
	// Remove removes the given path from being tracked.
	Remove(path string) error
	// Get returns the nibID for the given path.
	Get(path string) (*NIBSearchResponse, error)
	// SearchPrefix returns all nibIDs with the given path.
	// The map being returned has the paths
	SearchPrefix(prefix string) ([]*NIBSearchResponse, error)
}
