//go:build integration_test && !prepare_integration_test

package _switch

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mavolin/corgi/test/internal/outcheck"
)

func TestSwitch(t *testing.T) {
	t.Parallel()

	t.Run("foo", func(t *testing.T) {
		t.Parallel()

		w := outcheck.New(t, "switch.foo.expect")
		err := Switch(w, "foo")
		require.NoError(t, err)
	})

	t.Run("bar", func(t *testing.T) {
		t.Parallel()

		w := outcheck.New(t, "switch.bar.expect")
		err := Switch(w, "bar")
		require.NoError(t, err)
	})

	t.Run("abc", func(t *testing.T) {
		t.Parallel()

		w := outcheck.New(t, "switch.abc.expect")
		err := Switch(w, "abc")
		require.NoError(t, err)
	})

	t.Run("foobar", func(t *testing.T) {
		t.Parallel()

		w := outcheck.New(t, "switch.foobar.expect")
		err := Switch(w, "foobar")
		require.NoError(t, err)
	})
}
