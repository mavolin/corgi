//go:build integration_test && !prepare_integration_test

package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mavolin/corgi/test/internal/outcheck"
)

func TestFileName(t *testing.T) {
	t.Parallel()

	assert.FileExists(t, "tacos.go")
}

func TestFormat(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "format.expect")
	err := Test(w)
	require.NoError(t, err)
}
