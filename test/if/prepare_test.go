//go:build prepare_integration_test

package _if //nolint:revive
import (
	"testing"

	"github.com/mavolin/corgi/test/internal/compile"
)

func TestIf(t *testing.T) {
	t.Parallel()
	compile.Compile(t, "if.corgi", compile.Options{Package: "_if"})
}
