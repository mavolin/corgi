package code

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/should"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

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
				want      *ast.BlockFunction
				wantError string
			}{
				{
					name: "missing closing parenthesis",
					in:   "block(",
					want: &ast.BlockFunction{
						LParen: &ast.Position{Line: 1, Col: 6},
						Block:  &ast.Position{Line: 1, Col: 1},
					},
					wantError: "block function: unclosed arguments",
				}, {
					name: "too many arguments",
					in:   "block(foo, bar)",
					want: &ast.BlockFunction{
						LParen:    &ast.Position{Line: 1, Col: 6},
						BlockName: &ast.Identifier{Name: "foo", Position: &ast.Position{Line: 1, Col: 7}},
						RParen:    &ast.Position{Line: 1, Col: 15},
						Block:     &ast.Position{Line: 1, Col: 1},
					},
					wantError: "block function: too many arguments",
				},
			}

			for _, c := range tests {
				t.Run(c.name, func(t *testing.T) {
					t.Parallel()

					got := parsetest.ParsesUntilExtra(t, c.in, " ", f, parsetest.WantErrors(c.wantError))
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

			got := parsetest.ParsesUntilEOS(t, in, f)
			should.Equal(t, got, want)
		})

		t.Run("failure", func(t *testing.T) {
			t.Parallel()

			tests := []struct {
				name      string
				in        string
				want      *ast.Ternary
				wantError string
			}{
				{
					name: "no args",
					in:   "?()",
					want: &ast.Ternary{
						QuestionMark: &ast.Position{Line: 1, Col: 1},
						LParen:       &ast.Position{Line: 1, Col: 2},
						RParen:       &ast.Position{Line: 1, Col: 3},
					},
					wantError: "ternary function: missing arguments",
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
					wantError: "ternary function: missing if-true and if-false values",
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
					wantError: "ternary function: missing if-false value",
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
					wantError: "ternary function: too many arguments",
				},
			}

			for _, c := range tests {
				t.Run(c.name, func(t *testing.T) {
					t.Parallel()

					got := parsetest.ParsesUntilEOS(t, c.in, f, parsetest.WantErrors(c.wantError))
					should.Equal(t, got, c.want)
				})
			}
		})
	}
}
