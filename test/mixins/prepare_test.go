//go:build prepare_integration_test

package mixins

import (
	"testing"

	"github.com/mavolin/corgi/test/internal/compile"
)

func TestMixins(t *testing.T) {
	t.Parallel()
	compile.Compile(t, "mixins.corgi", compile.Options{})
}

func TestBlocks(t *testing.T) {
	t.Parallel()
	compile.Compile(t, "blocks.corgi", compile.Options{})
}

func TestAnd(t *testing.T) {
	t.Parallel()
	compile.Compile(t, "and.corgi", compile.Options{})
}

func TestExternal(t *testing.T) {
	t.Parallel()
	compile.Compile(t, "external.corgi", compile.Options{})
}

func TestExternalAlias(t *testing.T) {
	t.Parallel()
	compile.Compile(t, "external_alias.corgi", compile.Options{})
}

func TestInit(t *testing.T) {
	t.Parallel()
	compile.Compile(t, "init.corgi", compile.Options{})
}
