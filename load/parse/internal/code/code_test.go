package code

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestGoCode(t *testing.T) {
	t.Parallel()
	testGoCode(false)(t, GoCode(false))
}

func testGoCode(withEOS bool) func(t *testing.T, f parser.Func[[]ast.CodeNode]) {
	return func(t *testing.T, f parser.Func[[]ast.CodeNode]) {
		testCases := []struct {
			name   string
			code   string
			expect []ast.CodeNode
		}{
			{
				name: "identifier",
				code: "woof",
			}, {
				name: "comma in parentheses",
				code: "(woof, bark)",
			}, {
				name: "comma in brackets",
				code: "[woof, bark]",
			}, {
				name: "comma in braces",
				code: "{woof, bark}",
			}, {
				name: "semicolon in braces",
				code: "{woof; bark}",
			}, {
				name: "rune literal",
				code: "';'",
			}, {
				name: "mix",
				code: "foo(bar, baz) + string(';')",
			}, {
				name: "string in parentheses",
				code: "(\"foo\")",
				expect: []ast.CodeNode{
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

		for _, c := range testCases {
			t.Run(c.name, func(t *testing.T) {
				t.Parallel()

				expect := c.expect
				if expect == nil {
					expect = []ast.CodeNode{&ast.GoCode{Code: c.code, Position: &ast.Position{Line: 1, Col: 1}}}
				}

				if withEOS {
					c.code += ";"
				}
				actual := parsesCodeNodeFully(t, c.code, f)
				assert.Equal(t, expect, actual)
			})
		}
	}
}

func TestBlockFunction(t *testing.T) {
	t.Parallel()
	testBlockFunction(false)(t, BlockFunction())
}

func testBlockFunction(withEOS bool) func(t *testing.T, f parser.Func[*ast.BlockFunction]) {
	return func(t *testing.T, f parser.Func[*ast.BlockFunction]) {
		t.Run("success", func(t *testing.T) {
			t.Parallel()

			in := "block(foo)"
			if withEOS {
				in += ";"
			}
			expect := &ast.BlockFunction{
				LParen:    &ast.Position{Line: 1, Col: 6},
				BlockName: &ast.Ident{Ident: "foo", Position: &ast.Position{Line: 1, Col: 7}},
				RParen:    &ast.Position{Line: 1, Col: 10},
				Block:     &ast.Position{Line: 1, Col: 1},
			}

			actual := parsesCodeNodeFully(t, in, f)
			assert.Equal(t, expect, actual)
		})
		t.Run("failure", func(t *testing.T) {
			t.Parallel()

			testCases := []struct {
				name    string
				in      string
				expect  *ast.BlockFunction
				needEOS bool
			}{
				{
					name:    "missing block name",
					in:      "block()",
					needEOS: true,
					expect: &ast.BlockFunction{
						LParen: &ast.Position{Line: 1, Col: 6},
						RParen: &ast.Position{Line: 1, Col: 7},
						Block:  &ast.Position{Line: 1, Col: 1},
					},
				}, {
					name: "missing closing parenthesis",
					in:   "block(",
					expect: &ast.BlockFunction{
						LParen: &ast.Position{Line: 1, Col: 6},
						Block:  &ast.Position{Line: 1, Col: 1},
					},
				}, {
					name:    "too many arguments",
					in:      "block(foo, bar)",
					needEOS: true,
					expect: &ast.BlockFunction{
						LParen:    &ast.Position{Line: 1, Col: 6},
						BlockName: &ast.Ident{Ident: "foo", Position: &ast.Position{Line: 1, Col: 7}},
						RParen:    &ast.Position{Line: 1, Col: 15},
						Block:     &ast.Position{Line: 1, Col: 1},
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

func TestTernary(t *testing.T) {
	t.Parallel()
	testTernary(false)(t, Ternary())
}

func testTernary(withEOS bool) func(t *testing.T, f parser.Func[*ast.Ternary]) {
	return func(t *testing.T, f parser.Func[*ast.Ternary]) {
		t.Run("success", func(t *testing.T) {
			t.Parallel()

			in := "?(condition, ifTrue, ifFalse)"
			if withEOS {
				in += ";"
			}
			expect := &ast.Ternary{
				QuestionMark: &ast.Position{Line: 1, Col: 1},
				LParen:       &ast.Position{Line: 1, Col: 2},
				Condition: &ast.Expression{
					Code: ast.Code{
						&ast.GoCode{Code: "condition", Position: &ast.Position{Line: 1, Col: 3}},
					},
				},
				TrueVal: &ast.Expression{
					Code: ast.Code{
						&ast.GoCode{Code: "ifTrue", Position: &ast.Position{Line: 1, Col: 14}},
					},
				},
				FalseVal: &ast.Expression{
					Code: ast.Code{
						&ast.GoCode{Code: "ifFalse", Position: &ast.Position{Line: 1, Col: 22}},
					},
				},
				RParen: &ast.Position{Line: 1, Col: 29},
			}

			actual := parsesCodeNodeFully(t, in, f)
			assert.Equal(t, expect, actual)
		})

		t.Run("failure", func(t *testing.T) {
			t.Parallel()

			testCases := []struct {
				name    string
				in      string
				expect  *ast.Ternary
				needEOS bool
			}{
				{
					name:    "no args",
					in:      "?()",
					needEOS: true,
					expect: &ast.Ternary{
						QuestionMark: &ast.Position{Line: 1, Col: 1},
						LParen:       &ast.Position{Line: 1, Col: 2},
						RParen:       &ast.Position{Line: 1, Col: 3},
					},
				}, {
					name:    "only condition",
					in:      "?(condition)",
					needEOS: true,
					expect: &ast.Ternary{
						QuestionMark: &ast.Position{Line: 1, Col: 1},
						LParen:       &ast.Position{Line: 1, Col: 2},
						Condition: &ast.Expression{
							Code: ast.Code{
								&ast.GoCode{Code: "condition", Position: &ast.Position{Line: 1, Col: 3}},
							},
						},
						RParen: &ast.Position{Line: 1, Col: 12},
					},
				}, {
					name:    "missing ifFalse",
					in:      "?(condition, ifTrue)",
					needEOS: true,
					expect: &ast.Ternary{
						QuestionMark: &ast.Position{Line: 1, Col: 1},
						LParen:       &ast.Position{Line: 1, Col: 2},
						Condition: &ast.Expression{
							Code: ast.Code{
								&ast.GoCode{Code: "condition", Position: &ast.Position{Line: 1, Col: 3}},
							},
						},
						TrueVal: &ast.Expression{
							Code: ast.Code{
								&ast.GoCode{Code: "ifTrue", Position: &ast.Position{Line: 1, Col: 14}},
							},
						},
						RParen: &ast.Position{Line: 1, Col: 20},
					},
				}, {
					name:    "too many args",
					in:      "?(condition, ifTrue, ifFalse, foo)",
					needEOS: true,
					expect: &ast.Ternary{
						QuestionMark: &ast.Position{Line: 1, Col: 1},
						LParen:       &ast.Position{Line: 1, Col: 2},
						Condition: &ast.Expression{
							Code: ast.Code{
								&ast.GoCode{Code: "condition", Position: &ast.Position{Line: 1, Col: 3}},
							},
						},
						TrueVal: &ast.Expression{
							Code: ast.Code{
								&ast.GoCode{Code: "ifTrue", Position: &ast.Position{Line: 1, Col: 14}},
							},
						},
						FalseVal: &ast.Expression{
							Code: ast.Code{
								&ast.GoCode{Code: "ifFalse", Position: &ast.Position{Line: 1, Col: 22}},
							},
						},
						RParen: &ast.Position{Line: 1, Col: 34},
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

func parsesCodeNodeFully[T any](t *testing.T, input string, f parser.Func[T]) T {
	t.Helper()

	p := testutil.NewParser(t, input+", 1other stuff")
	v := testutil.AssertNoError(t, p, f)

	line, col, index := testutil.CalcEnd(1, 1, 0, input)
	testutil.AssertPosition(t, p, line, col, index)

	return v
}
