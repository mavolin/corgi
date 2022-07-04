//go:build integration_test

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

	t.Run("html", func(t *testing.T) {
		t.Parallel()

		w := outcheck.New(t, "void_elements.html.expect")

		err := VoidElementsHTML(w)
		require.NoError(t, err)
	})

	t.Run("xhtml", func(t *testing.T) {
		t.Parallel()

		w := outcheck.New(t, "void_elements.xhtml.expect")

		err := VoidElementsXHTML(w)
		require.NoError(t, err)
	})
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

func TestMirror(t *testing.T) {
	t.Parallel()

	t.Run("html", func(t *testing.T) {
		t.Parallel()

		w := outcheck.New(t, "mirror.html.expect")

		err := MirrorHTML(w)
		require.NoError(t, err)
	})

	t.Run("xhtml", func(t *testing.T) {
		t.Parallel()

		w := outcheck.New(t, "mirror.xhtml.expect")

		err := MirrorXHTML(w)
		require.NoError(t, err)
	})

	t.Run("xml", func(t *testing.T) {
		t.Parallel()

		w := outcheck.New(t, "mirror.xml.expect")

		err := MirrorXML(w)
		require.NoError(t, err)
	})
}
