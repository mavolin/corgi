package must

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mavolin/corgi/v2/internal/test/should"
)

func Equal[T any](t testing.TB, want, got T, opts ...cmp.Option) bool {
	t.Helper()
	return must(t, should.Equal(t, want, got, opts...))
}

func NotEqual[T any](t testing.TB, want, got T, opts ...cmp.Option) bool {
	t.Helper()
	return must(t, should.NotEqual(t, want, got, opts...))
}

func NoError(t testing.TB, err error) bool {
	t.Helper()
	return must(t, should.NoError(t, err))
}

func Error(t testing.TB, want, got error) bool {
	t.Helper()
	return must(t, should.Error(t, want, got))
}

func True(t testing.TB, got bool) bool {
	t.Helper()
	return must(t, should.True(t, got))
}

func False(t testing.TB, got bool) bool {
	t.Helper()
	return must(t, should.False(t, got))
}

func must(t testing.TB, ok bool) bool {
	t.Helper()
	if !ok {
		t.FailNow()
	}
	return ok
}
