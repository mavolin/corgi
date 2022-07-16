//go:build prepare_integration_test

package include

import (
	"testing"

	"github.com/mavolin/corgi/test/internal/compile"
)

func TestInclude(t *testing.T) {
	t.Parallel()
	compile.Compile(t, "include.corgi", compile.Options{})
}
