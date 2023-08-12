//go:build integration_test && !prepare_integration_test

package and

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mavolin/corgi/test/internal/outcheck"
)

func TestAnd(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "and.expect")

	err := And(w)
	require.NoError(t, err)
}
