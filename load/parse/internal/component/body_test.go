package component

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/test/should"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

func TestBody(t *testing.T) {
	t.Parallel()
	parsetest.AssertAlsoFulfils(t, Body(), testExtend)
}

func TestExtend(t *testing.T) {
	t.Parallel()
	testExtend(t, Extend())
}

func testExtend(t *testing.T, f parser.Func[*ast.Extend]) {
	in := ":bar(baz: s)"
	want := &ast.Extend{
		ComponentCall: &ast.ComponentCall{
			Colon: &ast.Position{Line: 1, Col: 1},
			Header: &ast.ComponentCallHeader{
				Name: &ast.Identifier{
					Name:     "bar",
					Position: &ast.Position{Line: 1, Col: 1 + len(":")},
				},
				Arguments: &ast.Arguments{
					LParen: &ast.Position{Line: 1, Col: 1 + len(":bar")},
					List: []ast.Argument{
						&ast.ComponentArgument{
							Name: &ast.Identifier{
								Name:     "baz",
								Position: &ast.Position{Line: 1, Col: 1 + len(":bar(")},
							},
							Colon: &ast.Position{Line: 1, Col: 1 + len(":bar(baz")},
							Value: &ast.Expression{
								Nodes: ast.Code{
									&ast.GoCode{
										Code:     "s",
										Position: &ast.Position{Line: 1, Col: 1 + len(":bar(baz: ")},
									},
								},
							},
						},
					},
					RParen: &ast.Position{Line: 1, Col: 1 + len(":bar(baz: s")},
				},
			},
		},
	}

	got := parsetest.ParsesFully(t, in, f)
	should.Equal(t, want, got)
}
