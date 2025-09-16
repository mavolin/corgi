package golang

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/test/should"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

func TestIdentifier(t *testing.T) {
	t.Parallel()
	testIdentifier(t, Identifier())
}

func testIdentifier(t *testing.T, f parser.Func[*ast.Identifier]) {
	in := "foo"
	want := &ast.Identifier{
		Name:     "foo",
		Position: &ast.Position{Line: 1, Col: 1},
	}

	got := parsetest.ParsesExact(t, in, f)
	should.Equal(t, got, want)
}
