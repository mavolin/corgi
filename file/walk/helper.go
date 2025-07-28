package walk

import (
	"slices"
)

// Closest returns the closest parent of the passed type, or the zero value for
// T.
func Closest[T any](parents []*Context) T {
	for _, parent := range slices.Backward(parents) {
		if t, ok := parent.Node.(T); ok {
			return t
		}
	}

	var z T
	return z
}

// IsChildOf asserts that the visited item must be a child of T.
func IsChildOf[T any](parents []*Context) bool {
	for _, parent := range parents {
		if _, ok := parent.Node.(T); ok {
			return true
		}
	}

	return false
}
