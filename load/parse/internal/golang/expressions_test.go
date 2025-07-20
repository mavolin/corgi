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

func testQualifiedIdent(t *testing.T, f parser.Func[*ast.QualifiedIdentifier]) {
	in := "foo.bar"
	expect := &ast.QualifiedIdentifier{
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

	actual := testutil.ParsesFully(t, in, f)
	assert.Equal(t, expect, actual)
}

func TestAddOp(t *testing.T) {
	t.Parallel()

	testCases := []string{"+", "-", "|", "^"}
	for _, c := range testCases {
		t.Run(c, func(t *testing.T) {
			t.Parallel()
			actual := testutil.ParsesFully(t, c, AddOp())
			assert.Equal(t, c, actual)
		})
	}
}

func TestMulOp(t *testing.T) {
	t.Parallel()

	testCases := []string{"*", "/", "%", "<<", ">>", "&", "&^"}
	for _, c := range testCases {
		t.Run(c, func(t *testing.T) {
			t.Parallel()
			actual := testutil.ParsesFully(t, c, MulOp())
			assert.Equal(t, c, actual)
		})
	}
}
