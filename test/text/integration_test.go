//go:build integration_test

package text

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mavolin/corgi/test/internal/outcheck"
)

func TestDotBlock(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "dot_block.expect")

	err := DotBlock(w)
	require.NoError(t, err)
}

func TestPipes(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "pipes.expect")

	err := Pipes(w)
	require.NoError(t, err)
}
