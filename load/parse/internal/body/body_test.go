package body

import (
	"testing"

	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

func TestBody(t *testing.T) {
	t.Parallel()

	parsetest.AssertAlsoFulfils(t, Body(), testBracketText)
	parsetest.AssertAlsoFulfils(t, Body(), testScope)
}
