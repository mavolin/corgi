//go:build integration_test && !prepare_integration_test

package escaping

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mavolin/corgi/test/internal/outcheck"
)

func TestHTML(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "html.expect")

	err := HTML(w)
	require.NoError(t, err)
}

func TestScript(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "script.expect")

	err := Script(w)
	require.NoError(t, err)
}

func TestCSS(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "css.expect")

	err := CSS(w)
	require.NoError(t, err)
}
