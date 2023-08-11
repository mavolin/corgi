//go:build integration_test && !prepare_integration_test

package expressions

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mavolin/corgi/test/internal/outcheck"
)

func TestGoExpression(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "go_expression.expect")
	err := GoExpression(w, "corgi")
	require.NoError(t, err)
}

func TestTernaryExpression(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "ternary_expression.expect")
	err := TernaryExpression(w)
	require.NoError(t, err)
}

func TestNilCheckExpression(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "chain_expression.expect")
	err := NilCheckExpression(w)
	require.NoError(t, err)
}
