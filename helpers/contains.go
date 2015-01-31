package helpers

// SliceContainsString returns if the given slice contains
// the passed string.
func SliceContainsString(slice []string, item string) bool {
	for _, str := range slice {
		if item == str {
			return true
		}
	}
	return false
}
