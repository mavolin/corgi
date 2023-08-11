//go:build integration_test && !prepare_integration_test

package elements

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mavolin/corgi/test/internal/outcheck"
)

func TestEmptyElement(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "empty_element.expect")

	err := EmptyElement(w)
	require.NoError(t, err)
}

func TestVoidElements(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "void_elements.expect")

	err := VoidElements(w)
	require.NoError(t, err)
}

func TestAttributes(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "attributes.expect")

	err := Attributes(w)
	require.NoError(t, err)
}

func TestBlockExpansion(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "block_expansion.expect")

	err := BlockExpansion(w)
	require.NoError(t, err)
}

func TestSelfClosing(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "self_closing.expect")

	err := SelfClosing(w)
	require.NoError(t, err)
}

func TestBoolAttrs(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "bool_attrs.expect")

	err := BoolAttrs(w)
	require.NoError(t, err)
}
