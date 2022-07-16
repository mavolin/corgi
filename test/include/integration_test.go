//go:build integration_test

package include

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mavolin/corgi/test/internal/outcheck"
)

func TestInclude(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "include.expect")
	err := Include(w)
	require.NoError(t, err)
}
