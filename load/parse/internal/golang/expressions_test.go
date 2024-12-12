package golang

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestQualifiedIdent(t *testing.T) {
	t.Parallel()
	testQualifiedIdent(t, QualifiedIdent())
}

func testQualifiedIdent(t *testing.T, f parser.Func[*ast.QualifiedIdent]) {
	in := "foo.bar"
	expect := &ast.QualifiedIdent{
		Package: ast.Ident{
			Ident:    "foo",
			Position: ast.Position{Line: 1, Col: 1},
		},
		Dot: &ast.Position{Line: 1, Col: 4},
		Name: &ast.Ident{
			Ident:    "bar",
			Position: ast.Position{Line: 1, Col: 5},
		},
	}

	actual := testutil.ParsesFully(t, in, f)
	assert.Equal(t, expect, actual)
}
