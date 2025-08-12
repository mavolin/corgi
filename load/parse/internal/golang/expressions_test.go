package golang

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/test/should"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

func TestQualifiedIdent(t *testing.T) {
	t.Parallel()
	testQualifiedIdent(t, QualifiedIdent())
}

func testQualifiedIdent(t *testing.T, f parser.Func[*ast.QualifiedIdentifier]) {
	in := "foo.bar"
	want := &ast.QualifiedIdentifier{
		Package: &ast.Identifier{
			Name:     "foo",
			Position: &ast.Position{Line: 1, Col: 1},
		},
		Dot: &ast.Position{Line: 1, Col: 4},
		Name: &ast.Identifier{
			Name:     "bar",
			Position: &ast.Position{Line: 1, Col: 5},
		},
	}

	got := parsetest.ParsesFully(t, in, f)
	should.Equal(t, got, want)
}
