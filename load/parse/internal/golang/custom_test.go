package golang

import (
	"testing"

	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

func TestFullIdent(t *testing.T) {
	t.Parallel()

	parsetest.AlsoFulfils(t, FullIdent(), testIdentifier)
	parsetest.AlsoFulfils(t, FullIdent(), testQualifiedIdent)
}
