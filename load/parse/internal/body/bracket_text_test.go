package body

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestBracketText(t *testing.T) {
	t.Parallel()
	testBracketText(t, BracketText())
}

func testBracketText(t *testing.T, f parser.Func[*ast.BracketText]) {
	testCases := []struct {
		name   string
		in     string
		expect *ast.BracketText
	}{
		{
			name: "empty",
			in:   "[]",
			expect: &ast.BracketText{
				LBracket: &ast.Position{Line: 1, Col: 1},
				RBracket: &ast.Position{Line: 1, Col: 2},
			},
		}, {
			name: "single line",
			in:   "[ foo ]",
			expect: &ast.BracketText{
				LBracket: &ast.Position{Line: 1, Col: 1},
				Lines: ast.TextBlock{
					ast.TextLine{
						&ast.Text{
							Text:     "foo",
							Position: &ast.Position{Line: 1, Col: 3},
						},
					},
				},
				RBracket: &ast.Position{Line: 1, Col: 7},
			},
		}, {
			name: "multi line",
			in: "[\n" +
				"\tfoo\n" +
				"\tbar\n" +
				"]",
			expect: &ast.BracketText{
				LBracket: &ast.Position{Line: 1, Col: 1},
				Lines: ast.TextBlock{
					ast.TextLine{
						&ast.Text{
							Text:     "foo",
							Position: &ast.Position{Line: 2, Col: 2},
						},
					}, ast.TextLine{
						&ast.Text{
							Text:     "bar",
							Position: &ast.Position{Line: 3, Col: 2},
						},
					},
				},
				RBracket: &ast.Position{Line: 4, Col: 1},
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			actual := testutil.ParsesFully(t, c.in, f)
			assert.Equal(t, c.expect, actual)
		})
	}
}
