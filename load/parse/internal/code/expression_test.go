package code

import (
	"testing"

	"github.com/mavolin/corgi/v2/fancyerr"
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestExpression(t *testing.T) {
	t.Parallel()

	testutil.AssertAlsoFulfils(t, Expression(Regular), func(t *testing.T, f parser.Func[*ast.Expression]) {
		testZeroCoalescing(t, func(p *parser.Parser) (*ast.ZeroCoalescing, *fancyerr.Error) {
			e, err := f(p)
			if err != nil {
				return nil, err
			}

			if len(e.Code) != 1 {
				return nil, &fancyerr.Error{
					Message: "expected exactly one expression node",
					Primary: quickanno.Expected(p, p.Pos(), "an expression node"),
				}
			}

			return e.Code[0].(*ast.ZeroCoalescing), nil
		})
	})
	testutil.AssertAlsoFulfils(t, Expression(Regular), testNonZCExpression)

	t.Run("body follows", func(t *testing.T) {
		t.Parallel()

		in := "func() { return block(foo) }()"
		expect := &ast.Expression{
			Code: ast.Code{
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
					BlockName: &ast.Ident{
						Ident:    "foo",
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

		testCases := []struct {
			name   string
			body   string
			expect *ast.Expression
		}{
			{
				name: "scope",
				body: "{\n\tfoo\n}",
			}, {
				name: "bracket text",
				body: "[\n\tfoo\n]",
			},
		}

		for _, c := range testCases {
			t.Run(c.name, func(t *testing.T) {
				t.Parallel()
				t.Run("inline", func(t *testing.T) {
					t.Parallel()

					p := testutil.NewParser(t, in+" "+c.body+" 1other stuff")
					var actual *ast.Expression
					p.DoInline(func() {
						actual = testutil.AssertNoError(t, p, Expression(BodyFollows))
					})

					line, col, index := testutil.CalcEnd(1, 1, 0, in)
					testutil.AssertPosition(t, p, line, col, index)
					assert.Equal(t, expect, actual)
				})
				t.Run("not inline", func(t *testing.T) {
					t.Parallel()

					p := testutil.NewParser(t, in+" "+c.body+" 1other stuff")
					actual := testutil.AssertNoError(t, p, Expression(BodyFollows))

					line, col, index := testutil.CalcEnd(1, 1, 0, in)
					testutil.AssertPosition(t, p, line, col, index)
					assert.Equal(t, expect, actual)
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
	testutil.AssertAlsoFulfils(t, f, nodesAsExpression(testGoCode()))
	testutil.AssertAlsoFulfils(t, f, nodesAsExpression(nodeAsNodes(testBlockFunction())))
	testutil.AssertAlsoFulfils(t, f, nodesAsExpression(nodeAsNodes(testString())))
	testutil.AssertAlsoFulfils(t, f, nodesAsExpression(nodeAsNodes(testTernary())))
	t.Run("mix", func(t *testing.T) {
		t.Parallel()

		in := `foo(bar, "baz #{woof}") || block(myBlock) || ?(cond, ifT, ifF)`
		expect := &ast.Expression{
			Code: ast.Code{
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
								Code: ast.Code{
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
					BlockName: &ast.Ident{Ident: "myBlock", Position: &ast.Position{Line: 1, Col: 34}},
					RParen:    &ast.Position{Line: 1, Col: 41},
					Block:     &ast.Position{Line: 1, Col: 28},
				}, &ast.GoCode{
					Code:     "||",
					Position: &ast.Position{Line: 1, Col: 43},
				}, &ast.Ternary{
					QuestionMark: &ast.Position{Line: 1, Col: 46},
					LParen:       &ast.Position{Line: 1, Col: 47},
					Condition: &ast.Expression{
						Code: ast.Code{
							&ast.GoCode{
								Code:     "cond",
								Position: &ast.Position{Line: 1, Col: 48},
							},
						},
					},
					TrueVal: &ast.Expression{
						Code: ast.Code{
							&ast.GoCode{
								Code:     "ifT",
								Position: &ast.Position{Line: 1, Col: 54},
							},
						},
					},
					FalseVal: &ast.Expression{
						Code: ast.Code{
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

		actual := parsesCodeNodeFully(t, in, f)
		assert.Equal(t, expect, actual)
	})
}

func nodesAsExpression(subTest func(t *testing.T, f parser.Func[[]ast.CodeNode])) func(*testing.T, parser.Func[*ast.Expression]) {
	return func(t *testing.T, f parser.Func[*ast.Expression]) {
		subTest(t, func(p *parser.Parser) ([]ast.CodeNode, *fancyerr.Error) {
			e, err := f(p)
			if err != nil {
				return nil, err
			}
			ns := make([]ast.CodeNode, len(e.Code))
			for i, n := range e.Code {
				ns[i] = n.(ast.CodeNode)
			}
			return ns, err
		})
	}
}

func nodeAsNodes[N ast.CodeNode](subTest func(t *testing.T, f parser.Func[N])) func(*testing.T, parser.Func[[]ast.CodeNode]) {
	return func(t *testing.T, f parser.Func[[]ast.CodeNode]) {
		subTest(t, func(p *parser.Parser) (N, *fancyerr.Error) {
			var zero N
			ns, err := f(p)
			if err != nil {
				return zero, err
			}

			if len(ns) != 1 {
				return zero, &fancyerr.Error{
					Message: "expected exactly one expression node",
					Primary: quickanno.Expected(p, p.Pos(), "an expression node"),
				}
			}

			return ns[0].(N), nil
		})
	}
}
