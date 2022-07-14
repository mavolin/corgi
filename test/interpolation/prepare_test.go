//go:build prepare_integration_test

package interpolation

import (
	"testing"

	"github.com/mavolin/corgi/test/internal/compile"
)

func TestInterpolation(t *testing.T) {
	t.Parallel()
	compile.Compile(t, "interpolation.corgi", compile.Options{})
}

func TestUnescapedInterpolation(t *testing.T) {
	t.Parallel()
	compile.Compile(t, "unescaped_interpolation.corgi", compile.Options{})
}

func TestHashEscape(t *testing.T) {
	t.Parallel()
	compile.Compile(t, "hash_escape.corgi", compile.Options{})
}

func TestInlineText(t *testing.T) {
	t.Parallel()
	compile.Compile(t, "inline_text.corgi", compile.Options{})
}

func TestInlineElements(t *testing.T) {
	t.Parallel()
	compile.Compile(t, "inline_elements.corgi", compile.Options{})
}
