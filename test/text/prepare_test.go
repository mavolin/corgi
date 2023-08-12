//go:build prepare_integration_test

package text

import (
	"testing"

	"github.com/mavolin/corgi/test/internal/compile"
)

func TestDotBlock(t *testing.T) {
	t.Parallel()
	compile.Compile(t, "arrow_block.corgi", compile.Options{})
}
