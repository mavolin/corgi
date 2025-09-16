package control

import (
	"testing"

	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

func parsesCodeNodeFully[T any](t *testing.T, input string, f parser.Func[T]) T {
	t.Helper()

	p := parsetest.NewParser(t, input+"; 1other stuff")
	v := parsetest.AssertNoError(t, p, f)

	line, col, index := parsetest.CalcEnd(1, 1, 0, input)
	parsetest.AssertPosition(t, p, line, col, index)

	return v
}
