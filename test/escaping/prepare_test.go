//go:build prepare_integration_test

package escaping

import (
	"testing"

	"github.com/mavolin/corgi/test/internal/compile"
)

func TestHTML(t *testing.T) {
	t.Parallel()
	compile.Compile(t, "html.corgi", compile.Options{})
}

func TestScript(t *testing.T) {
	t.Parallel()
	compile.Compile(t, "script.corgi", compile.Options{})
}

func TestCSS(t *testing.T) {
	t.Parallel()
	compile.Compile(t, "css.corgi", compile.Options{})
}
