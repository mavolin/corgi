package code

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/test/should"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

func TestGoCode(t *testing.T) {
	t.Parallel()
	testGoCode()(t, GoCode(Regular))
}

func testGoCode() func(t *testing.T, f parser.Func[[]ast.CodeNode]) {
	return func(t *testing.T, f parser.Func[[]ast.CodeNode]) {
		tests := []struct {
			name string
			code string
			want []ast.CodeNode
		}{
			{
				name: "identifier",
				code: "woof",
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
				want: []ast.CodeNode{
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

				want := c.want
				if want == nil {
					want = []ast.CodeNode{&ast.GoCode{Code: c.code, Position: &ast.Position{Line: 1, Col: 1}}}
				}

				got := parsesCodeNodeFully(t, c.code, f)
				should.Equal(t, got, want)
			})
		}
	}
}

func TestBlockFunction(t *testing.T) {
	t.Parallel()
	testBlockFunction()(t, BlockFunction())
}

func testBlockFunction() func(t *testing.T, f parser.Func[*ast.BlockFunction]) {
	return func(t *testing.T, f parser.Func[*ast.BlockFunction]) {
		t.Run("success", func(t *testing.T) {
			t.Parallel()

			tests := []struct {
				name string
				in   string
				want *ast.BlockFunction
			}{
				{
					name: "default block",
					in:   "block()",
					want: &ast.BlockFunction{
						Block:  &ast.Position{Line: 1, Col: 1},
						LParen: &ast.Position{Line: 1, Col: 1 + len("block")},
						RParen: &ast.Position{Line: 1, Col: 1 + len("block(")},
					},
				}, {
					name: "named block",
					in:   "block(foo)",
					want: &ast.BlockFunction{
						Block:  &ast.Position{Line: 1, Col: 1},
						LParen: &ast.Position{Line: 1, Col: 1 + len("block")},
						BlockName: &ast.Identifier{
							Name: "foo", Position: &ast.Position{Line: 1, Col: 1 + len("block(")},
						},
						RParen: &ast.Position{Line: 1, Col: 1 + len("block(foo")},
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
				want *ast.BlockFunction
			}{
				{
					name: "missing closing parenthesis",
					in:   "block(",
					want: &ast.BlockFunction{
						LParen: &ast.Position{Line: 1, Col: 6},
						Block:  &ast.Position{Line: 1, Col: 1},
					},
				}, {
					name: "too many arguments",
					in:   "block(foo, bar)",
					want: &ast.BlockFunction{
						LParen:    &ast.Position{Line: 1, Col: 6},
						BlockName: &ast.Identifier{Name: "foo", Position: &ast.Position{Line: 1, Col: 7}},
						RParen:    &ast.Position{Line: 1, Col: 15},
						Block:     &ast.Position{Line: 1, Col: 1},
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

func TestTernary(t *testing.T) {
	t.Parallel()
	testTernary()(t, Ternary())
}

func testTernary() func(t *testing.T, f parser.Func[*ast.Ternary]) {
	return func(t *testing.T, f parser.Func[*ast.Ternary]) {
		t.Run("success", func(t *testing.T) {
			t.Parallel()

			in := "?(condition, ifTrue, ifFalse)"
			want := &ast.Ternary{
				QuestionMark: &ast.Position{Line: 1, Col: 1},
				LParen:       &ast.Position{Line: 1, Col: 2},
				Condition: &ast.Expression{
					Nodes: ast.Code{
						&ast.GoCode{Code: "condition", Position: &ast.Position{Line: 1, Col: 3}},
					},
				},
				TrueVal: &ast.Expression{
					Nodes: ast.Code{
						&ast.GoCode{Code: "ifTrue", Position: &ast.Position{Line: 1, Col: 14}},
					},
				},
				FalseVal: &ast.Expression{
					Nodes: ast.Code{
						&ast.GoCode{Code: "ifFalse", Position: &ast.Position{Line: 1, Col: 22}},
					},
				},
				RParen: &ast.Position{Line: 1, Col: 29},
			}

			got := parsesCodeNodeFully(t, in, f)
			should.Equal(t, got, want)
		})

		t.Run("failure", func(t *testing.T) {
			t.Parallel()

			tests := []struct {
				name string
				in   string
				want *ast.Ternary
			}{
				{
					name: "no args",
					in:   "?()",
					want: &ast.Ternary{
						QuestionMark: &ast.Position{Line: 1, Col: 1},
						LParen:       &ast.Position{Line: 1, Col: 2},
						RParen:       &ast.Position{Line: 1, Col: 3},
					},
				}, {
					name: "only condition",
					in:   "?(condition)",
					want: &ast.Ternary{
						QuestionMark: &ast.Position{Line: 1, Col: 1},
						LParen:       &ast.Position{Line: 1, Col: 2},
						Condition: &ast.Expression{
							Nodes: ast.Code{
								&ast.GoCode{Code: "condition", Position: &ast.Position{Line: 1, Col: 3}},
							},
						},
						RParen: &ast.Position{Line: 1, Col: 12},
					},
				}, {
					name: "missing ifFalse",
					in:   "?(condition, ifTrue)",
					want: &ast.Ternary{
						QuestionMark: &ast.Position{Line: 1, Col: 1},
						LParen:       &ast.Position{Line: 1, Col: 2},
						Condition: &ast.Expression{
							Nodes: ast.Code{
								&ast.GoCode{Code: "condition", Position: &ast.Position{Line: 1, Col: 3}},
							},
						},
						TrueVal: &ast.Expression{
							Nodes: ast.Code{
								&ast.GoCode{Code: "ifTrue", Position: &ast.Position{Line: 1, Col: 14}},
							},
						},
						RParen: &ast.Position{Line: 1, Col: 20},
					},
				}, {
					name: "too many args",
					in:   "?(condition, ifTrue, ifFalse, foo)",
					want: &ast.Ternary{
						QuestionMark: &ast.Position{Line: 1, Col: 1},
						LParen:       &ast.Position{Line: 1, Col: 2},
						Condition: &ast.Expression{
							Nodes: ast.Code{
								&ast.GoCode{Code: "condition", Position: &ast.Position{Line: 1, Col: 3}},
							},
						},
						TrueVal: &ast.Expression{
							Nodes: ast.Code{
								&ast.GoCode{Code: "ifTrue", Position: &ast.Position{Line: 1, Col: 14}},
							},
						},
						FalseVal: &ast.Expression{
							Nodes: ast.Code{
								&ast.GoCode{Code: "ifFalse", Position: &ast.Position{Line: 1, Col: 22}},
							},
						},
						RParen: &ast.Position{Line: 1, Col: 34},
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

func parsesCodeNodeFully[T any](t *testing.T, input string, f parser.Func[T]) T {
	t.Helper()

	p := parsetest.NewParser(t, input+"; 1other stuff")
	v := parsetest.AssertNoError(t, p, f)

	line, col, index := parsetest.CalcEnd(1, 1, 0, input)
	parsetest.AssertPosition(t, p, line, col, index)

	return v
}
