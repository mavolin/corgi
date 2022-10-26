//go:build integration_test && !prepare_integration_test

package comments

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mavolin/corgi/test/internal/outcheck"
)

func TestComments(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "comments.expect")

	err := Comments(w)
	require.NoError(t, err)
}
