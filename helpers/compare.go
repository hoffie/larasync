package helpers

import (
	"reflect"

	"crypto/subtle"
)

// ConstantTimeBytesEqual returns true if both byte slices are identical.
// The comparison is attempted in constant time.
// Note: the length will leak due to timing side channels.
func ConstantTimeBytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	return subtle.ConstantTimeCompare(a, b) == 1
}

// StringsEqual returns if the slices of strings are now equal.
func StringsEqual(a, b []string) bool {
	return reflect.DeepEqual(a, b)
}
