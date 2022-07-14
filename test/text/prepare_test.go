//go:build prepare_integration_test

package text

import (
	"testing"

	"github.com/mavolin/corgi/test/internal/compile"
)

func TestDotBlock(t *testing.T) {
	t.Parallel()
	compile.Compile(t, "dot_block.corgi", compile.Options{})
}

func TestPipes(t *testing.T) {
	t.Parallel()
	compile.Compile(t, "pipes.corgi", compile.Options{})
}
