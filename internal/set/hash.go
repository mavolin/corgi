package set

// HashSet is a set implementation backed by a hash map.
// It is suitable for bigger sets of elements.
type HashSet[K comparable] struct {
	s map[K]struct{}
}

var _ Set[any] = (*HashSet[any])(nil)

func NewHashSet[K comparable](capacity int) *HashSet[K] {
	return &HashSet[K]{s: make(map[K]struct{}, capacity)}
}

func (s *HashSet[K]) Add(k K) {
	s.s[k] = struct{}{}
}

func (s *HashSet[K]) Contains(k K) bool {
	_, ok := s.s[k]
	return ok
}

func (s *HashSet[K]) Clear() {
	s.s = make(map[K]struct{}, len(s.s))
}
