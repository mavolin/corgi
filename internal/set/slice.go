package set

import "slices"

// SliceSet is a set implementation backed by a slice.
// It is suitable for smalls sets of elements.
type SliceSet[K comparable] struct {
	s []K
}

var _ Set[any] = (*SliceSet[any])(nil)

func NewSliceSet[K comparable](capacity int) *SliceSet[K] {
	return &SliceSet[K]{s: make([]K, 0, capacity)}
}

func (s *SliceSet[K]) Add(k K) {
	if !slices.Contains(s.s, k) {
		s.s = append(s.s, k)
	}
}

func (s *SliceSet[K]) Contains(k K) bool {
	return slices.Contains(s.s, k)
}

func (s *SliceSet[K]) Clear() {
	s.s = s.s[:0]
}
