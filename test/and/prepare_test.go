//go:build prepare_integration_test

package and

import (
	"testing"

	"github.com/mavolin/corgi/test/internal/compile"
)

func TestAnd(t *testing.T) {
	t.Parallel()
	compile.Compile(t, "and.corgi", compile.Options{})
}
