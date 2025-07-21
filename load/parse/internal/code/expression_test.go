package code

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/internal/test/should"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

func TestExpression(t *testing.T) {
	t.Parallel()

	parsetest.AssertAlsoFulfils(t, Expression(Regular), func(t *testing.T, f parser.Func[*ast.Expression]) {
		testZeroCoalescing(t, func(p *parser.Parser) (*ast.ZeroCoalescing, *diagnostic.Diagnostic) {
			e, err := f(p)
			if err != nil {
				return nil, err
			}

			if len(e.Nodes) != 1 {
				return nil, &diagnostic.Diagnostic{
					Message: "expected exactly one expression node",
					Primary: quickanno.Expected(p, p.Pos(), "an expression node"),
				}
			}

			return e.Nodes[0].(*ast.ZeroCoalescing), nil
		})
	})
	parsetest.AssertAlsoFulfils(t, Expression(Regular), testNonZCExpression)

	t.Run("body follows", func(t *testing.T) {
		t.Parallel()

		in := "func() { return block(foo) }()"
		want := &ast.Expression{
			Nodes: ast.Code{
				&ast.GoCode{
					Code:     "func",
					Position: &ast.Position{Line: 1, Col: 1},
				}, &ast.GoCode{
					Code:     "()",
					Position: &ast.Position{Line: 1, Col: 5},
				}, &ast.GoCode{
					Code:     "{ return",
					Position: &ast.Position{Line: 1, Col: 8},
				}, &ast.BlockFunction{
					Block:  &ast.Position{Line: 1, Col: 17},
					LParen: &ast.Position{Line: 1, Col: 22},
					BlockName: &ast.Identifier{
						Name:     "foo",
						Position: &ast.Position{Line: 1, Col: 23},
					},
					RParen: &ast.Position{Line: 1, Col: 26},
				}, &ast.GoCode{
					Code:     "}",
					Position: &ast.Position{Line: 1, Col: 28},
				}, &ast.GoCode{
					Code:     "()",
					Position: &ast.Position{Line: 1, Col: 29},
				},
			},
		}

		tests := []struct {
			name string
			body string
			want *ast.Expression
		}{
			{
				name: "scope",
				body: "{\n\tfoo\n}",
			}, {
				name: "bracket text",
				body: "[\n\tfoo\n]",
			},
		}

		for _, c := range tests {
			t.Run(c.name, func(t *testing.T) {
				t.Parallel()
				t.Run("inline", func(t *testing.T) {
					t.Parallel()

					p := parsetest.NewParser(t, in+" "+c.body+" 1other stuff")
					var got *ast.Expression
					p.DoInline(func() {
						got = parsetest.AssertNoError(t, p, Expression(BodyFollows))
					})

					line, col, index := parsetest.CalcEnd(1, 1, 0, in)
					parsetest.AssertPosition(t, p, line, col, index)
					should.Equal(t, want, got)
				})
				t.Run("not inline", func(t *testing.T) {
					t.Parallel()

					p := parsetest.NewParser(t, in+" "+c.body+" 1other stuff")
					got := parsetest.AssertNoError(t, p, Expression(BodyFollows))

					line, col, index := parsetest.CalcEnd(1, 1, 0, in)
					parsetest.AssertPosition(t, p, line, col, index)
					should.Equal(t, want, got)
				})
			})
		}
	})
}

func TestNonZCExpression(t *testing.T) {
	t.Parallel()
	testNonZCExpression(t, NonZCExpression(Regular))
}

func testNonZCExpression(t *testing.T, f parser.Func[*ast.Expression]) {
	parsetest.AssertAlsoFulfils(t, f, nodesAsExpression(testGoCode()))
	parsetest.AssertAlsoFulfils(t, f, nodesAsExpression(nodeAsNodes(testBlockFunction())))
	parsetest.AssertAlsoFulfils(t, f, nodesAsExpression(nodeAsNodes(testString())))
	parsetest.AssertAlsoFulfils(t, f, nodesAsExpression(nodeAsNodes(testTernary())))
	t.Run("mix", func(t *testing.T) {
		t.Parallel()

		in := `foo(bar, "baz #{woof}") || block(myBlock) || ?(cond, ifT, ifF)`
		want := &ast.Expression{
			Nodes: ast.Code{
				&ast.GoCode{
					Code:     "foo(bar,",
					Position: &ast.Position{Line: 1, Col: 1},
				}, &ast.String{
					Open:  &ast.Position{Line: 1, Col: 10},
					Quote: '"',
					Contents: []ast.StringNode{
						&ast.StringText{
							Text:     "baz ",
							Position: &ast.Position{Line: 1, Col: 11},
						}, &ast.ExpressionInterpolation{
							Hash:   &ast.Position{Line: 1, Col: 15},
							LBrace: &ast.Position{Line: 1, Col: 16},
							Expression: &ast.Expression{
								Nodes: ast.Code{
									&ast.GoCode{
										Code:     "woof",
										Position: &ast.Position{Line: 1, Col: 17},
									},
								},
							},
							RBrace: &ast.Position{Line: 1, Col: 21},
						},
					},
					Close: &ast.Position{Line: 1, Col: 22},
				}, &ast.GoCode{
					Code:     ") ||",
					Position: &ast.Position{Line: 1, Col: 23},
				}, &ast.BlockFunction{
					LParen:    &ast.Position{Line: 1, Col: 33},
					BlockName: &ast.Identifier{Name: "myBlock", Position: &ast.Position{Line: 1, Col: 34}},
					RParen:    &ast.Position{Line: 1, Col: 41},
					Block:     &ast.Position{Line: 1, Col: 28},
				}, &ast.GoCode{
					Code:     "||",
					Position: &ast.Position{Line: 1, Col: 43},
				}, &ast.Ternary{
					QuestionMark: &ast.Position{Line: 1, Col: 46},
					LParen:       &ast.Position{Line: 1, Col: 47},
					Condition: &ast.Expression{
						Nodes: ast.Code{
							&ast.GoCode{
								Code:     "cond",
								Position: &ast.Position{Line: 1, Col: 48},
							},
						},
					},
					TrueVal: &ast.Expression{
						Nodes: ast.Code{
							&ast.GoCode{
								Code:     "ifT",
								Position: &ast.Position{Line: 1, Col: 54},
							},
						},
					},
					FalseVal: &ast.Expression{
						Nodes: ast.Code{
							&ast.GoCode{
								Code:     "ifF",
								Position: &ast.Position{Line: 1, Col: 59},
							},
						},
					},
					RParen: &ast.Position{Line: 1, Col: 62},
				},
			},
		}

		got := parsesCodeNodeFully(t, in, f)
		should.Equal(t, want, got)
	})
}

func nodesAsExpression(subTest func(t *testing.T, f parser.Func[[]ast.CodeNode])) func(*testing.T, parser.Func[*ast.Expression]) {
	return func(t *testing.T, f parser.Func[*ast.Expression]) {
		subTest(t, func(p *parser.Parser) ([]ast.CodeNode, *diagnostic.Diagnostic) {
			e, err := f(p)
			if err != nil {
				return nil, err
			}
			ns := make([]ast.CodeNode, len(e.Nodes))
			copy(ns, e.Nodes)
			return ns, err
		})
	}
}

func nodeAsNodes[N ast.CodeNode](subTest func(t *testing.T, f parser.Func[N])) func(*testing.T, parser.Func[[]ast.CodeNode]) {
	return func(t *testing.T, f parser.Func[[]ast.CodeNode]) {
		subTest(t, func(p *parser.Parser) (N, *diagnostic.Diagnostic) {
			var zero N
			ns, err := f(p)
			if err != nil {
				return zero, err
			}

			if len(ns) != 1 {
				return zero, &diagnostic.Diagnostic{
					Message: "expected exactly one expression node",
					Primary: quickanno.Expected(p, p.Pos(), "an expression node"),
				}
			}

			return ns[0].(N), nil
		})
	}
}
