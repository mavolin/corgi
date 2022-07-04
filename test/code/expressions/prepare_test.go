//go:build prepare_integration_test

package expressions

import (
	"testing"

	"github.com/mavolin/corgi/test/internal/compile"
)

func TestGoExpression(t *testing.T) {
	t.Parallel()

	compile.Compile(t, "go_expression.corgi", compile.Options{})
}

func TestTernaryExpression(t *testing.T) {
	t.Parallel()

	compile.Compile(t, "ternary_expression.corgi", compile.Options{})
}

func TestNilCheckExpression(t *testing.T) {
	t.Parallel()

	compile.Compile(t, "nil_check_expression.corgi", compile.Options{})
}
