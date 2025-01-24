package body

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestBody(t *testing.T) {
	t.Parallel()

	testutil.AssertAlsoFulfils(t, Body(), testBracketText)
	testutil.AssertAlsoFulfils(t, Body(), testScope)
}

func TestComponentCallBody(t *testing.T) {
	t.Parallel()

	testutil.AssertAlsoFulfils(t, ComponentCallBody(), testBracketText)
	testutil.AssertAlsoFulfils(t, ComponentCallBody(), testScope)
	testutil.AssertAlsoFulfils(t, ComponentCallBody(), testUnderscoreBlockShorthand)
}

func TestUnderscoreBlockShorthand(t *testing.T) {
	t.Parallel()
	testUnderscoreBlockShorthand(t, UnderscoreBlockShorthand())
}

func testUnderscoreBlockShorthand(t *testing.T, f parser.Func[*ast.UnderscoreBlockShorthand]) {
	testCases := []struct {
		name   string
		in     string
		expect *ast.UnderscoreBlockShorthand
	}{
		{
			name: "bracket text",
			in:   "_[ foo ]",
			expect: &ast.UnderscoreBlockShorthand{
				Body: &ast.BracketText{
					LBracket: ast.Position{Line: 1, Col: 2},
					Lines: ast.TextBlock{
						ast.TextLine{
							&ast.Text{
								Text:     "foo",
								Position: ast.Position{Line: 1, Col: 4},
							},
						},
					},
					RBracket: &ast.Position{Line: 1, Col: 8},
				},
				Position: ast.Position{Line: 1, Col: 1},
			},
		}, {
			name: "scope",
			in:   "_{}",
			expect: &ast.UnderscoreBlockShorthand{
				Body: &ast.Scope{
					LBrace: ast.Position{Line: 1, Col: 2},
					RBrace: &ast.Position{Line: 1, Col: 3},
				},
				Position: ast.Position{Line: 1, Col: 1},
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
