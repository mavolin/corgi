package code

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	t.Parallel()
	testString(false)(t, String())
}

func testString(withEOS bool) func(t *testing.T, f parser.Func[*ast.String]) {
	return func(t *testing.T, f parser.Func[*ast.String]) {
		t.Run("success", func(t *testing.T) {
			t.Parallel()

			testCases := []struct {
				name   string
				in     string
				expect *ast.String
			}{
				{
					name: "simple interpreted string",
					in:   `"foo"`,
					expect: &ast.String{
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
					expect: &ast.String{
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
					expect: &ast.String{
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
									Code: ast.Code{
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

			for _, c := range testCases {
				t.Run(c.name, func(t *testing.T) {
					t.Parallel()

					if withEOS {
						c.in += ";"
					}
					actual := parsesCodeNodeFully(t, c.in, f)
					assert.Equal(t, c.expect, actual)
				})
			}
		})
		t.Run("failure", func(t *testing.T) {
			t.Parallel()

			testCases := []struct {
				name    string
				in      string
				expect  *ast.String
				needEOS bool
			}{
				{
					name: "missing closing quote",
					in:   `"foo`,
					expect: &ast.String{
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

			for _, c := range testCases {
				t.Run(c.name, func(t *testing.T) {
					t.Parallel()

					if c.needEOS && withEOS {
						c.in += ";"
					}
					actual := testutil.MatchesButError(t, c.in, f)
					assert.Equal(t, c.expect, actual)
				})
			}
		})
	}
}
