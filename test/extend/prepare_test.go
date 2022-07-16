//go:build prepare_integration_test

package extend

import (
	"testing"

	"github.com/mavolin/corgi/test/internal/compile"
)

func TestBlock(t *testing.T) {
	t.Parallel()
	compile.Compile(t, "block.corgi", compile.Options{})
}

func TestAppend(t *testing.T) {
	t.Parallel()
	compile.Compile(t, "append.corgi", compile.Options{})
}
