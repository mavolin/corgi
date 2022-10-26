//go:build integration_test && !prepare_integration_test

package mixins

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mavolin/corgi/test/internal/outcheck"
	"github.com/mavolin/corgi/test/internal/voidwriter"
)

func TestMixins(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "mixins.expect")
	err := Mixins(w)
	require.NoError(t, err)
}

func TestShadowing(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "shadowing.expect")
	err := Nesting(w)
	require.NoError(t, err)
}

func TestBlocks(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "blocks.expect")
	err := Blocks(w)
	require.NoError(t, err)
}

func TestAnd(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "and.expect")
	err := And(w)
	require.NoError(t, err)
}

func TestExternal(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "external.expect")
	err := External(w)
	require.NoError(t, err)
}

func TestExternalAlias(t *testing.T) {
	t.Parallel()

	w := outcheck.New(t, "external_alias.expect")
	err := ExternalAlias(w)
	require.NoError(t, err)
}

func TestInit(t *testing.T) {
	t.Parallel()

	var executed bool

	err := Init(voidwriter.Writer, &executed)
	require.NoError(t, err)

	assert.True(t, executed, "init mixin was not executed")
}
