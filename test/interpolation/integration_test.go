//go:build integration_test && !prepare_integration_test

package interpolation

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mavolin/corgi/test/internal/outcheck"
)

func TestInterpolation(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "interpolation.expect")
	err := Interpolation(w)
	require.NoError(t, err)
}

func TestHashEscape(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "hash_escape.expect")
	err := HashEscape(w)
	require.NoError(t, err)
}

func TestInlineText(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "inline_text.expect")
	err := InlineText(w)
	require.NoError(t, err)
}

func TestInlineElements(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "inline_elements.expect")
	err := InlineElement(w)
	require.NoError(t, err)
}
