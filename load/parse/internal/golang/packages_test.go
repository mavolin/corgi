package golang

import (
	"testing"

	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

func TestPackageName(t *testing.T) {
	t.Parallel()

	parsetest.AssertAlsoFulfils(t, PackageName(), testIdentifier)
}
