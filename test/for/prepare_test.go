//go:build prepare_integration_test

package _for

import (
	"testing"

	"github.com/mavolin/corgi/test/internal/compile"
)

func TestFor(t *testing.T) {
	t.Parallel()
	compile.Compile(t, "for.corgi", compile.Options{Package: "_for"})
}

func TestWhile(t *testing.T) {
	t.Parallel()
	compile.Compile(t, "while.corgi", compile.Options{Package: "_for"})
}
