//go:build integration_test

package doctype

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/mavolin/corgi/test/internal/outcheck"
)

func TestOneDotOne(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "1.1.expect")

	err := OneDotOne(w)
	require.NoError(t, err)
}

func TestBasic(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "basic.expect")

	err := Basic(w)
	require.NoError(t, err)
}

func TestCustomDoctype(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "custom_doctype.expect")

	err := CustomDoctype(w)
	require.NoError(t, err)
}

func TestCustomProlog(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "custom_prolog.expect")

	err := CustomProlog(w)
	require.NoError(t, err)
}

func TestFrameset(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "frameset.expect")

	err := Frameset(w)
	require.NoError(t, err)
}

func TestHTML(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "html.expect")

	err := HTML(w)
	require.NoError(t, err)
}

func TestMobile(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "mobile.expect")

	err := Mobile(w)
	require.NoError(t, err)
}

func TestPList(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "plist.expect")

	err := PList(w)
	require.NoError(t, err)
}

func TestStrict(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "strict.expect")

	err := Strict(w)
	require.NoError(t, err)
}

func TestTransitional(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "transitional.expect")

	err := Transitional(w)
	require.NoError(t, err)
}

func TestXML(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "xml.expect")

	err := XML(w)
	require.NoError(t, err)
}
