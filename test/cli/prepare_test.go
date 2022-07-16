//go:build prepare_integration_test

package cli

import (
	"os"
	"testing"

	"github.com/mavolin/corgi/test/internal/compile"
)

func TestFileName(t *testing.T) {
	t.Parallel()

	_ = os.Remove("tacos.go")

	compile.Compile(t, "test.corgi", compile.Options{
		OutName: "tacos.go",
	})
}

func TestFormat(t *testing.T) {
	t.Parallel()
	compile.Compile(t, "format.corgi", compile.Options{Format: true})
}
