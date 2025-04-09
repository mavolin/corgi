package set

import "slices"

// SliceSet is a set implementation backed by a slice.
// It is suitable for smalls sets of elements.
type SliceSet[K comparable] struct {
	s []K
}

func NewSliceSet[K comparable](cap int) *SliceSet[K] {
	return &SliceSet[K]{s: make([]K, 0, cap)}
}

func (s *SliceSet[K]) Add(k K) bool {
	if !slices.Contains(s.s, k) {
		s.s = append(s.s, k)
		return true
	}
	return false
}

func (s *SliceSet[K]) Contains(k K) bool {
	return slices.Contains(s.s, k)
}

func (s *SliceSet[K]) Clear() {
	s.s = s.s[:0]
}
