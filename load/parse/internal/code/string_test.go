package code

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/should"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

func TestString(t *testing.T) {
	t.Parallel()
	testString()(t, String())
}

func testString() func(t *testing.T, f parser.Func[*ast.String]) {
	return func(t *testing.T, f parser.Func[*ast.String]) {
		t.Run("success", func(t *testing.T) {
			t.Parallel()

			tests := []struct {
				name string
				in   string
				want *ast.String
			}{
				{
					name: "simple interpreted string",
					in:   `"foo"`,
					want: &ast.String{
						Open:  &ast.Position{Line: 1, Col: 1},
						Quote: '"',
						Contents: []ast.StringNode{
							&ast.StringText{
								Text:     "foo",
								Position: &ast.Position{Line: 1, Col: 2},
							},
						},
						Close: &ast.Position{Line: 1, Col: 5},
					},
				}, {
					name: "simple raw string",
					in:   "`bar`",
					want: &ast.String{
						Open:  &ast.Position{Line: 1, Col: 1},
						Quote: '`',
						Contents: []ast.StringNode{
							&ast.StringText{
								Text:     "bar",
								Position: &ast.Position{Line: 1, Col: 2},
							},
						},
						Close: &ast.Position{Line: 1, Col: 5},
					},
				}, {
					name: "with interpolation",
					in:   `"foo #{bar} baz"`,
					want: &ast.String{
						Open:  &ast.Position{Line: 1, Col: 1},
						Quote: '"',
						Contents: []ast.StringNode{
							&ast.StringText{
								Text:     "foo ",
								Position: &ast.Position{Line: 1, Col: ast.Col(1 + len(`"`))},
							}, &ast.ExpressionInterpolation{
								Hash:   &ast.Position{Line: 1, Col: ast.Col(1 + len(`"foo `))},
								LBrace: &ast.Position{Line: 1, Col: ast.Col(1 + len(`"foo #`))},
								Expression: &ast.Expression{
									Nodes: ast.Code{
										&ast.GoCode{
											Code:     "bar",
											Position: &ast.Position{Line: 1, Col: ast.Col(1 + len(`"foo #{`))},
										},
									},
								},
								RBrace: &ast.Position{Line: 1, Col: ast.Col(1 + len(`"foo #{bar`))},
							}, &ast.StringText{
								Text:     " baz",
								Position: &ast.Position{Line: 1, Col: ast.Col(1 + len(`"foo #{bar}`))},
							},
						},
						Close: &ast.Position{Line: 1, Col: ast.Col(1 + len(`"foo #{bar} baz`))},
					},
				},
			}

			for _, c := range tests {
				t.Run(c.name, func(t *testing.T) {
					t.Parallel()

					got := parsetest.ParsesUntilEOS(t, c.in, f)
					should.Equal(t, got, c.want)
				})
			}
		})
		t.Run("failure", func(t *testing.T) {
			t.Parallel()

			tests := []struct {
				name      string
				in        string
				want      *ast.String
				wantError string
			}{
				{
					name: "missing closing quote",
					in:   `"foo`,
					want: &ast.String{
						Open:  &ast.Position{Line: 1, Col: 1},
						Quote: '"',
						Contents: []ast.StringNode{
							&ast.StringText{
								Text:     "foo",
								Position: &ast.Position{Line: 1, Col: 2},
							},
						},
					},
					wantError: "string: missing closing quote",
				},
			}

			for _, c := range tests {
				t.Run(c.name, func(t *testing.T) {
					t.Parallel()

					got := parsetest.ParsesUntilExtra(t, c.in, "", f, parsetest.WantErrors(c.wantError))
					should.Equal(t, got, c.want)
				})
			}
		})
	}
}
