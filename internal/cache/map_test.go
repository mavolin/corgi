package cache

import (
	"testing"

	"github.com/mavolin/corgi/v2/internal/should"
)

func TestMap_Get(t *testing.T) {
	t.Parallel()

	key := "woof"
	wantV := 42

	m := NewMap[string, int]()

	var calls int
	v := m.Get(key, func() int {
		calls++
		return wantV
	})
	should.Equal(t, v, wantV)
	should.Equal(t, calls, 1)

	v = m.Get(key, func() int {
		calls++
		return wantV + 1
	})
	should.Equal(t, v, wantV)
	should.Equal(t, calls, 1)
}

func TestMap_Preload(t *testing.T) {
	t.Parallel()

	t.Run("compute in background", func(t *testing.T) {
		t.Parallel()

		key := "woof"
		wantV := 77

		m := NewMap[string, int]()

		var calls int
		m.Preload(key, func() int {
			calls++
			return wantV
		})

		v := m.Get(key, func() int {
			return wantV + 1
		})
		should.Equal(t, v, wantV)
		should.Equal(t, calls, 1)
	})

	t.Run("does not overwrite existing", func(t *testing.T) {
		t.Parallel()

		key := "woof"
		wantV := 88

		m := NewMap[string, int]()
		_ = m.Get(key, func() int { return wantV })

		var calls int
		m.Preload(key, func() int {
			calls++
			return wantV + 1
		})

		v := m.Get(key, func() int { return wantV + 2 })
		should.Equal(t, v, wantV)
		should.Equal(t, calls, 0)
	})
}
