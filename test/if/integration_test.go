//go:build integration_test && !prepare_integration_test

package _if

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mavolin/corgi/test/internal/outcheck"
)

func TestIf(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "if.expect")
	err := If(w)
	require.NoError(t, err)
}
