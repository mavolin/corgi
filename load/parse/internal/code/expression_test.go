package code

import (
	"fmt"
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/test/should"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

func TestExpression(t *testing.T) {
	t.Parallel()

	parsetest.AlsoFulfils(t, Expression(), func(t *testing.T, f parser.Func[*ast.Expression]) {
		testZeroCoalescing(t, func(p *parser.Parser) *ast.ZeroCoalescing {
			e := f(p)
			if e == nil {
				return nil
			}

			if len(e.Nodes) != 1 {
				return nil
			}

			return e.Nodes[0].(*ast.ZeroCoalescing)
		})
	})
	parsetest.AlsoFulfils(t, Expression(), testSimpleExpression)

	t.Run("body follows", func(t *testing.T) {
		t.Parallel()

		in := "func(){ return block(foo) }()"
		want := &ast.Expression{
			Nodes: ast.Code{
				&ast.GoCode{
					Code:     "func(){ return",
					Position: &ast.Position{Line: 1, Col: 1},
				}, &ast.BlockFunction{
					Block:  &ast.Position{Line: 1, Col: 1 + len("func(){ return ")},
					LParen: &ast.Position{Line: 1, Col: 1 + len("func(){ return block")},
					BlockName: &ast.Identifier{
						Name:     "foo",
						Position: &ast.Position{Line: 1, Col: 1 + len("func(){ return block(")},
					},
					RParen: &ast.Position{Line: 1, Col: 1 + len("func(){ return block(foo")},
				}, &ast.GoCode{
					Code:     "}()",
					Position: &ast.Position{Line: 1, Col: 1 + len("func(){ return block(foo) ")},
				},
			},
		}

		got := parsetest.ParsesUntilBody(t, in, Expression())
		should.Equal(t, got, want)
	})
}

func TestSimpleExpression(t *testing.T) {
	t.Parallel()
	testSimpleExpression(t, SimpleExpression())
}

func testSimpleExpression(t *testing.T, f parser.Func[*ast.Expression]) {
	parsetest.AlsoFulfils(t, f, testEnhancedExpression())
	parsetest.AlsoFulfils(t, f, nodeAsExpression(testBlockFunction()))
	parsetest.AlsoFulfils(t, f, nodeAsExpression(testString()))
	parsetest.AlsoFulfils(t, f, nodeAsExpression(testTernary()))
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

		got := parsetest.ParsesUntilEOS(t, in, f)
		should.Equal(t, got, want)
	})
}

func TestEnhancedExpression(t *testing.T) {
	t.Parallel()
	testEnhancedExpression()(t, EnhancedExpression())
}

func testEnhancedExpression() func(t *testing.T, f parser.Func[*ast.Expression]) {
	return func(t *testing.T, f parser.Func[*ast.Expression]) {
		tests := []struct {
			name string
			code string
			want ast.Code
		}{
			{
				name: "identifier",
				code: "woof",
			}, {
				name: "comparison",
				code: "len(woof) >= len(bark)",
			}, {
				name: "comma in parentheses",
				code: "(woof, bark)",
			}, {
				name: "comma in brackets",
				code: "a[woof, bark]",
			}, {
				name: "comma in braces",
				code: "func(){woof, bark}",
			}, {
				name: "semicolon in braces",
				code: "func(){woof; bark}",
			}, {
				name: "rune literal",
				code: "';'",
			}, {
				name: "mix",
				code: "foo(bar, baz) + string(';')",
			}, {
				name: "string in parentheses",
				code: "(\"foo\")",
				want: ast.Code{
					&ast.GoCode{Code: "(", Position: &ast.Position{Line: 1, Col: 1}},
					&ast.String{
						Open:  &ast.Position{Line: 1, Col: 2},
						Quote: '"',
						Contents: []ast.StringNode{
							&ast.StringText{Text: "foo", Position: &ast.Position{Line: 1, Col: 3}},
						},
						Close: &ast.Position{Line: 1, Col: 6},
					},
					&ast.GoCode{Code: ")", Position: &ast.Position{Line: 1, Col: 7}},
				},
			},
		}

		for _, c := range tests {
			t.Run(c.name, func(t *testing.T) {
				t.Parallel()

				if c.want == nil {
					c.want = ast.Code{&ast.GoCode{Code: c.code, Position: &ast.Position{Line: 1, Col: 1}}}
				}

				got := parsetest.ParsesUntilEOS(t, c.code, f)
				should.Equal(t, got, &ast.Expression{Nodes: c.want})
			})
		}
	}
}

func nodeAsExpression[N comparableNode](subTest func(t *testing.T, f parser.Func[N])) func(*testing.T, parser.Func[*ast.Expression]) {
	var zero N
	return func(t *testing.T, f parser.Func[*ast.Expression]) {
		subTest(t, func(p *parser.Parser) N {
			e := f(p)
			if e == nil {
				return zero
			}
			if len(e.Nodes) != 1 {
				panic(fmt.Sprintf("expected exactly one node, got %d", len(e.Nodes)))
			}
			n, _ := e.Nodes[0].(N)
			if n == zero {
				panic(fmt.Sprintf("expected node of type %T, got %T", zero, e.Nodes[0]))
			}
			return n
		})
	}
}
