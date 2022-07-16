//go:build integration_test

package filter

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mavolin/corgi/test/internal/outcheck"
)

func TestFilter(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "filter.expect")
	err := Filter(w)
	require.NoError(t, err)
}
