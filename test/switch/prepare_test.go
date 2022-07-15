//go:build prepare_integration_test

package _switch

import (
	"testing"

	"github.com/mavolin/corgi/test/internal/compile"
)

func TestSwitch(t *testing.T) {
	t.Parallel()
	compile.Compile(t, "switch.corgi", compile.Options{Package: "_switch"})
}
