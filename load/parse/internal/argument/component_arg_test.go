package argument

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestComponentArgument(t *testing.T) {
	t.Parallel()
	testutil.AssertAlsoFulfils(t, ComponentArgument(), testComponentArgument)
}

func testComponentArgument(t *testing.T, f parser.Func[*ast.ComponentArgument]) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()

		in := "name: value"

		expect := &ast.ComponentArgument{
			Name: &ast.Ident{
				Ident:    "name",
				Position: &ast.Position{Line: 1, Col: 1},
			},
			Colon: &ast.Position{Line: 1, Col: 5},
			Value: &ast.Expression{
				Nodes: ast.Code{
					&ast.GoCode{
						Code:     "value",
						Position: &ast.Position{Line: 1, Col: 7},
					},
				},
			},
		}

		// we add ", other" to test that the parser stops at the comma
		p := testutil.NewParser(t, in+", other")
		actual := testutil.AssertNoError(t, p, f)
		if assert.Equal(t, expect, actual) {
			testutil.AssertPosition(t, p, expect.End().Line, expect.End().Col, len(in))
		}
	})

	t.Run("recover", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name   string
			in     string
			expect *ast.ComponentArgument
		}{
			{
				name: "no colon space: no equal sign",
				in:   "name:value 2",
				expect: &ast.ComponentArgument{
					Name: &ast.Ident{
						Ident:    "name",
						Position: &ast.Position{Line: 1, Col: 1},
					},
					Colon: &ast.Position{Line: 1, Col: 5},
					Value: &ast.Expression{
						Nodes: ast.Code{
							&ast.GoCode{
								Code:     "value 2",
								Position: &ast.Position{Line: 1, Col: 6},
							},
						},
					},
				},
			}, {
				name: "no colon space: string",
				in:   `name:"value"`,
				expect: &ast.ComponentArgument{
					Name: &ast.Ident{
						Ident:    "name",
						Position: &ast.Position{Line: 1, Col: 1},
					},
					Colon: &ast.Position{Line: 1, Col: 5},
					Value: &ast.Expression{
						Nodes: ast.Code{
							&ast.String{
								Open:  &ast.Position{Line: 1, Col: 6},
								Quote: '"',
								Contents: []ast.StringNode{
									&ast.StringText{
										Text:     "value",
										Position: &ast.Position{Line: 1, Col: 7},
									},
								},
								Close: &ast.Position{Line: 1, Col: 12},
							},
						},
					},
				},
			},
		}

		for _, c := range testCases {
			t.Run(c.name, func(t *testing.T) {
				t.Parallel()
				// we add ", other" to test that the parser stops at the comma
				p := testutil.NewParser(t, c.in+", other")
				actual := testutil.AssertMatchesButError(t, p, f)
				if assert.Equal(t, c.expect, actual) {
					testutil.AssertPosition(t, p, c.expect.End().Line, c.expect.End().Col, len(c.in))
				}
			})
		}
	})
}
