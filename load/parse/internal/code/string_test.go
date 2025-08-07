package code

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/test/should"
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
					in:   `"foo #%1.2f{bar} baz"`,
					want: &ast.String{
						Open:  &ast.Position{Line: 1, Col: 1},
						Quote: '"',
						Contents: []ast.StringNode{
							&ast.StringText{
								Text:     "foo ",
								Position: &ast.Position{Line: 1, Col: 2},
							}, &ast.ExpressionInterpolation{
								Hash:            &ast.Position{Line: 1, Col: 6},
								FormatDirective: "1.2f",
								LBrace:          &ast.Position{Line: 1, Col: 12},
								Expression: &ast.Expression{
									Nodes: ast.Code{
										&ast.GoCode{
											Code:     "bar",
											Position: &ast.Position{Line: 1, Col: 13},
										},
									},
								},
								RBrace: &ast.Position{Line: 1, Col: 16},
							}, &ast.StringText{
								Text:     " baz",
								Position: &ast.Position{Line: 1, Col: 17},
							},
						},
						Close: &ast.Position{Line: 1, Col: 21},
					},
				},
			}

			for _, c := range tests {
				t.Run(c.name, func(t *testing.T) {
					t.Parallel()

					got := parsesCodeNodeFully(t, c.in, f)
					should.Equal(t, got, c.want)
				})
			}
		})
		t.Run("failure", func(t *testing.T) {
			t.Parallel()

			tests := []struct {
				name string
				in   string
				want *ast.String
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
				},
			}

			for _, c := range tests {
				t.Run(c.name, func(t *testing.T) {
					t.Parallel()

					got := parsetest.MatchesButError(t, c.in, f)
					should.Equal(t, got, c.want)
				})
			}
		})
	}
}
