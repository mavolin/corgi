//go:build integration_test

package _for

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mavolin/corgi/test/internal/outcheck"
)

func TestFor(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "for.expect")
	err := For(w)
	require.NoError(t, err)
}

func TestWhile(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "while.expect")
	err := While(w)
	require.NoError(t, err)
}
