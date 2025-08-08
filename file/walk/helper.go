package walk

import (
	"slices"
)

// Closest returns the closest parent of the passed type, or the zero value for
// T.
func Closest[T any](parents []*Context) T {
	i := ClosestIndex[T](parents)
	if i < 0 {
		var z T
		return z
	}
	return parents[i].Node.(T) //nolint:errcheck
}

// ClosestIndex returns the index of the closest parent of the passed type, or
// -1 if no such parent exists.
func ClosestIndex[T any](parents []*Context) int {
	for i, parent := range slices.Backward(parents) {
		if _, ok := parent.Node.(T); ok {
			return i
		}
	}

	return -1
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
