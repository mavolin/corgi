//go:build integration_test

package extend

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mavolin/corgi/test/internal/outcheck"
)

func TestBlock(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "block.expect")
	err := Block(w)
	require.NoError(t, err)
}

func TestAppend(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "append.expect")
	err := Append(w)
	require.NoError(t, err)
}
