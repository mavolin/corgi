//go:build integration_test && !prepare_integration_test

package text

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mavolin/corgi/test/internal/outcheck"
)

func TestArrowBlock(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "arrow_block.expect")

	err := ArrowBlock(w)
	require.NoError(t, err)
}
