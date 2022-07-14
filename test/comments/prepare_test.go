//go:build prepare_integration_test

package comments

import (
	"testing"

	"github.com/mavolin/corgi/test/internal/compile"
)

func TestComments(t *testing.T) {
	t.Parallel()
	compile.Compile(t, "comments.corgi", compile.Options{})
}
