package set

import (
	"testing"

	"github.com/mavolin/corgi/v2/internal/test/should"
)

func testSet(t *testing.T, create func() Set[string], count func(s Set[string], k string) int) {
	t.Parallel()

	key1 := "foo"
	key2 := key1 + "bar"

	t.Run("Add", func(t *testing.T) {
		t.Parallel()

		s := create()

		should.Equal(t, count(s, key1), 0)
		s.Add(key1)
		should.Equal(t, count(s, key1), 1)
		s.Add(key1)
		should.Equal(t, count(s, key1), 1)

		should.Equal(t, count(s, key2), 0)
		s.Add(key2)
		should.Equal(t, count(s, key2), 1)
		should.Equal(t, count(s, key1), 1)
	})

	t.Run("Contains", func(t *testing.T) {
		t.Parallel()

		s := create()

		should.False(t, s.Contains(""))
		should.False(t, s.Contains(key1))
		should.False(t, s.Contains(key2))

		s.Add(key1)
		should.True(t, s.Contains(key1))
		should.False(t, s.Contains(key2))

		s.Add(key2)
		should.True(t, s.Contains(key1))
		should.True(t, s.Contains(key2))
	})

	t.Run("Clear", func(t *testing.T) {
		t.Parallel()

		s := create()

		s.Add(key1)
		s.Add(key2)
		should.True(t, s.Contains(key1))
		should.True(t, s.Contains(key2))

		s.Clear()
		should.False(t, s.Contains(key1))
		should.False(t, s.Contains(key2))
	})
}
