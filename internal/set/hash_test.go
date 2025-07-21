package set

import "testing"

func TestHashSet(t *testing.T) {
	testSet(t,
		func() Set[string] { return NewHashSet[string](2) },
		func(s Set[string], k string) int {
			m := s.(*HashSet[string]).s
			if _, ok := m[k]; ok {
				return 1
			}
			return 0
		})
}
