//go:build prepare_integration_test

package filter

import (
	"testing"

	"github.com/mavolin/corgi/test/internal/compile"
)

func TestFilter(t *testing.T) {
	t.Parallel()
	compile.Compile(t, "filter.corgi", compile.Options{AllowedFilters: []string{"rev"}})
}
