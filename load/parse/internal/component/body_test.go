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
	parsetest.AlsoFulfils(t, Body(), testExtend)
	t.Run("Body", func(t *testing.T) {
		t.Parallel()
		in := "{\n" +
			"\tbr\n" +
			"}"
		want := &ast.Scope{
			LBrace: &ast.Position{Line: 1, Col: 1},
			Nodes: []ast.ScopeNode{
				&ast.Element{
					Header: &ast.ElementHeader{
						Name: &ast.ElementReference{
							Name: &ast.ElementName{
								Name:     "br",
								Position: &ast.Position{Line: 2, Col: 1 + len("\t")},
							},
						},
					},
				},
			},
			RBrace: &ast.Position{Line: 3, Col: 1},
		}

		got := parsetest.ParsesExact(t, in, Body())
		should.Equal[ast.ComponentBody](t, got, want)
	})
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

	got := parsetest.ParsesExact(t, in, f)
	should.Equal(t, got, want)
}
