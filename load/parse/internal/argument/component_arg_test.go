package argument

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/test/should"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

func TestComponentArgument(t *testing.T) {
	t.Parallel()
	parsetest.AssertAlsoFulfils(t, ComponentArgument(), testComponentArgument)
}

func testComponentArgument(t *testing.T, f parser.Func[*ast.ComponentArgument]) {
	t.Run("success", func(t *testing.T) {
		t.Parallel()

		in := "name: value"

		want := &ast.ComponentArgument{
			Name: &ast.Identifier{
				Name:     "name",
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
		p := parsetest.NewParser(t, in+", other")
		got := parsetest.AssertNoError(t, p, f)
		if should.Equal(t, got, want) {
			parsetest.AssertPosition(t, p, want.End().Line, want.End().Col, len(in))
		}
	})

	t.Run("recover", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name string
			in   string
			want *ast.ComponentArgument
		}{
			{
				name: "no colon space: no equal sign",
				in:   "name:value 2",
				want: &ast.ComponentArgument{
					Name: &ast.Identifier{
						Name:     "name",
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
				want: &ast.ComponentArgument{
					Name: &ast.Identifier{
						Name:     "name",
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

		for _, c := range tests {
			t.Run(c.name, func(t *testing.T) {
				t.Parallel()
				// we add ", other" to test that the parser stops at the comma

				p := parsetest.NewParser(t, c.in+", other")
				got := parsetest.AssertMatchesButError(t, p, f)
				if should.Equal(t, got, c.want) {
					parsetest.AssertPosition(t, p, c.want.End().Line, c.want.End().Col, len(c.in))
				}
			})
		}
	})
}
