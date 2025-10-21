package must

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/mavolin/corgi/v2/internal/test/should"
)

func Equal[T any](t testing.TB, got, want T, opts ...cmp.Option) bool {
	t.Helper()
	return must(t, should.Equal(t, got, want, opts...))
}

func NotEqual[T any](t testing.TB, got, want T, opts ...cmp.Option) bool {
	t.Helper()
	return must(t, should.NotEqual(t, got, want, opts...))
}

func Panic(t testing.TB, f func(), wantMessage string) bool {
	t.Helper()
	return must(t, should.Panic(t, f, wantMessage))
}

func NotPanic(t testing.TB, f func()) (didPanic bool) {
	t.Helper()
	return must(t, should.NotPanic(t, f))
}

func Error(t testing.TB, err error, wantMessage string) bool {
	t.Helper()
	return must(t, should.Error(t, err, wantMessage))
}

func NoError(t testing.TB, err error) bool {
	t.Helper()
	return must(t, should.NoError(t, err))
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
