package set

import "testing"

func TestSliceSet(t *testing.T) {
	testSet(t,
		func() Set[string] { return NewSliceSet[string](2) },
		func(s Set[string], k string) int {
			m := s.(*SliceSet[string]).s
			count := 0
			for _, v := range m {
				if v == k {
					count++
				}
			}
			return count
		})
}
