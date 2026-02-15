package code

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/should"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

func TestConstantString(t *testing.T) {
	t.Parallel()
	parsetest.AlsoFulfils(t, ConstantString("tests"), testConstantInterpretedString())
	parsetest.AlsoFulfils(t, ConstantString("tests"), testConstantRawString())
}

func TestConstantInterpretedString(t *testing.T) {
	t.Parallel()
	testConstantInterpretedString()(t, ConstantInterpretedString("tests"))
}

func testConstantInterpretedString() func(t *testing.T, f parser.Func[*ast.InterpretedString]) {
	return func(t *testing.T, f parser.Func[*ast.InterpretedString]) {
		t.Run("success", func(t *testing.T) {
			t.Parallel()

			tests := []struct {
				name string
				in   string
				want *ast.InterpretedString
			}{
				{
					name: "text",
					in:   `"foo"`,
					want: &ast.InterpretedString{
						Open: &ast.Position{Line: 1, Col: 1},
						Contents: []ast.InterpretedStringNode{
							&ast.InterpretedStringText{
								Text:     "foo",
								Position: &ast.Position{Line: 1, Col: ast.Col(1 + len(`"`))},
							},
						},
						Close: &ast.Position{Line: 1, Col: ast.Col(1 + len(`"foo`))},
					},
				}, {
					name: "character reference",
					in:   `"#mdash;"`,
					want: &ast.InterpretedString{
						Open: &ast.Position{Line: 1, Col: 1},
						Contents: []ast.InterpretedStringNode{
							&ast.CharacterReference{
								Hash:  &ast.Position{Line: 1, Col: ast.Col(1 + len(`"`))},
								Name:  "mdash",
								Chars: "—",
							},
						},
						Close: &ast.Position{Line: 1, Col: ast.Col(1 + len(`"#mdash;`))},
					},
				}, {
					name: "character escape",
					in:   `"##"`,
					want: &ast.InterpretedString{
						Open: &ast.Position{Line: 1, Col: 1},
						Contents: []ast.InterpretedStringNode{
							&ast.CharacterEscape{
								Hash:   &ast.Position{Line: 1, Col: ast.Col(1 + len(`"`))},
								Symbol: '#',
								Rune:   '#',
							},
						},
						Close: &ast.Position{Line: 1, Col: ast.Col(1 + len(`"##`))},
					},
				}, {
					name: "combination",
					in:   `"foo #mdash;##"`,
					want: &ast.InterpretedString{
						Open: &ast.Position{Line: 1, Col: 1},
						Contents: []ast.InterpretedStringNode{
							&ast.InterpretedStringText{
								Text:     "foo ",
								Position: &ast.Position{Line: 1, Col: ast.Col(1 + len(`"`))},
							}, &ast.CharacterReference{
								Hash:  &ast.Position{Line: 1, Col: ast.Col(1 + len(`"foo `))},
								Name:  "mdash",
								Chars: "—",
							}, &ast.CharacterEscape{
								Hash:   &ast.Position{Line: 1, Col: ast.Col(1 + len(`"foo #mdash;`))},
								Symbol: '#',
								Rune:   '#',
							},
						},
						Close: &ast.Position{Line: 1, Col: ast.Col(1 + len(`"foo #mdash;##`))},
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

			const wantError = "constant string: use of non-constant expression"

			tests := []struct {
				name string
				in   string
				want *ast.InterpretedString
			}{
				{
					name: "expression interpolation",
					in:   `"#{woof}"`,
					want: &ast.InterpretedString{
						Open: &ast.Position{Line: 1, Col: 1},
						Contents: []ast.InterpretedStringNode{
							&ast.ExpressionInterpolation{
								Hash:   &ast.Position{Line: 1, Col: ast.Col(1 + len(`"`))},
								LBrace: &ast.Position{Line: 1, Col: ast.Col(1 + len(`"#`))},
								Expression: &ast.Expression{
									Nodes: ast.Code{
										&ast.GoCode{
											Code:     "woof",
											Position: &ast.Position{Line: 1, Col: ast.Col(1 + len(`"#{`))},
										},
									},
								},
								RBrace: &ast.Position{Line: 1, Col: ast.Col(1 + len(`"#{woof`))},
							},
						},
						Close: &ast.Position{Line: 1, Col: ast.Col(1 + len(`"#{woof}`))},
					},
				}, {
					name: "component call interpolation",
					in:   `"#:woof()"`,
					want: &ast.InterpretedString{
						Open: &ast.Position{Line: 1, Col: 1},
						Contents: []ast.InterpretedStringNode{
							&ast.ComponentCallInterpolation{
								Hash: &ast.Position{Line: 1, Col: ast.Col(1 + len(`"`))},
								ComponentCall: &ast.ComponentCall{
									Colon: &ast.Position{Line: 1, Col: ast.Col(1 + len(`"#`))},
									Header: &ast.ComponentCallHeader{
										Name: &ast.Identifier{
											Name:     "woof",
											Position: &ast.Position{Line: 1, Col: ast.Col(1 + len(`"#:`))},
										},
										Arguments: &ast.Arguments{
											LParen: &ast.Position{Line: 1, Col: ast.Col(1 + len(`"#:woof`))},
											RParen: &ast.Position{Line: 1, Col: ast.Col(1 + len(`"#:woof(`))},
										},
									},
								},
							},
						},
						Close: &ast.Position{
							Line: 1, Col: ast.Col(1 + len(`"#:woof()`)),
						},
					},
				},
			}

			for _, c := range tests {
				t.Run(c.name, func(t *testing.T) {
					t.Parallel()

					got := parsetest.ParsesUntilExtra(t, c.in, "", f, parsetest.WantErrors(wantError))
					should.Equal(t, got, c.want)
				})
			}
		})
	}
}

func TestConstantRawString(t *testing.T) {
	t.Parallel()
	testConstantRawString()(t, ConstantRawString("tests"))
}

func testConstantRawString() func(t *testing.T, f parser.Func[*ast.RawString]) {
	return func(t *testing.T, f parser.Func[*ast.RawString]) {
		t.Run("success", func(t *testing.T) {
			t.Parallel()

			tests := []struct {
				name string
				in   string
				want *ast.RawString
			}{
				{
					name: "text",
					in:   "`foo`",
					want: &ast.RawString{
						Open: &ast.Position{Line: 1, Col: 1},
						Contents: []ast.RawStringNode{
							&ast.RawStringText{
								Text:     "foo",
								Position: &ast.Position{Line: 1, Col: ast.Col(1 + len("`"))},
							},
						},
						Close: &ast.Position{Line: 1, Col: ast.Col(1 + len("`foo"))},
					},
				}, {
					name: "character reference",
					in:   "`#mdash;`",
					want: &ast.RawString{
						Open: &ast.Position{Line: 1, Col: 1},
						Contents: []ast.RawStringNode{
							&ast.CharacterReference{
								Hash:  &ast.Position{Line: 1, Col: ast.Col(1 + len("`"))},
								Name:  "mdash",
								Chars: "—",
							},
						},
						Close: &ast.Position{Line: 1, Col: ast.Col(1 + len("`#mdash;"))},
					},
				}, {
					name: "character escape",
					in:   "`##`",
					want: &ast.RawString{
						Open: &ast.Position{Line: 1, Col: 1},
						Contents: []ast.RawStringNode{
							&ast.CharacterEscape{
								Hash:   &ast.Position{Line: 1, Col: ast.Col(1 + len("`"))},
								Symbol: '#',
								Rune:   '#',
							},
						},
						Close: &ast.Position{Line: 1, Col: ast.Col(1 + len("`##"))},
					},
				}, {
					name: "combination",
					in:   "`foo #mdash;##`",
					want: &ast.RawString{
						Open: &ast.Position{Line: 1, Col: 1},
						Contents: []ast.RawStringNode{
							&ast.RawStringText{
								Text:     "foo ",
								Position: &ast.Position{Line: 1, Col: ast.Col(1 + len("`"))},
							}, &ast.CharacterReference{
								Hash:  &ast.Position{Line: 1, Col: ast.Col(1 + len("`foo "))},
								Name:  "mdash",
								Chars: "—",
							}, &ast.CharacterEscape{
								Hash:   &ast.Position{Line: 1, Col: ast.Col(1 + len("`foo #mdash;"))},
								Symbol: '#',
								Rune:   '#',
							},
						},
						Close: &ast.Position{Line: 1, Col: ast.Col(1 + len("`foo #mdash;##"))},
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

			const wantError = "constant string: use of non-constant expression"

			tests := []struct {
				name string
				in   string
				want *ast.RawString
			}{
				{
					name: "expression interpolation",
					in:   "`#{woof}`",
					want: &ast.RawString{
						Open: &ast.Position{Line: 1, Col: 1},
						Contents: []ast.RawStringNode{
							&ast.ExpressionInterpolation{
								Hash:   &ast.Position{Line: 1, Col: ast.Col(1 + len("`"))},
								LBrace: &ast.Position{Line: 1, Col: ast.Col(1 + len("`#"))},
								Expression: &ast.Expression{
									Nodes: ast.Code{
										&ast.GoCode{
											Code:     "woof",
											Position: &ast.Position{Line: 1, Col: ast.Col(1 + len("`#{"))},
										},
									},
								},
								RBrace: &ast.Position{Line: 1, Col: ast.Col(1 + len("`#{woof"))},
							},
						},
						Close: &ast.Position{Line: 1, Col: ast.Col(1 + len("`#{woof}"))},
					},
				}, {
					name: "component call interpolation",
					in:   "`#:woof()`",
					want: &ast.RawString{
						Open: &ast.Position{Line: 1, Col: 1},
						Contents: []ast.RawStringNode{
							&ast.ComponentCallInterpolation{
								Hash: &ast.Position{Line: 1, Col: ast.Col(1 + len("`"))},
								ComponentCall: &ast.ComponentCall{
									Colon: &ast.Position{Line: 1, Col: ast.Col(1 + len("`#"))},
									Header: &ast.ComponentCallHeader{
										Name: &ast.Identifier{
											Name:     "woof",
											Position: &ast.Position{Line: 1, Col: ast.Col(1 + len("`#:"))},
										},
										Arguments: &ast.Arguments{
											LParen: &ast.Position{Line: 1, Col: ast.Col(1 + len("`#:woof"))},
											RParen: &ast.Position{Line: 1, Col: ast.Col(1 + len("`#:woof("))},
										},
									},
								},
							},
						},
						Close: &ast.Position{
							Line: 1, Col: ast.Col(1 + len("`#:woof()")),
						},
					},
				},
			}

			for _, c := range tests {
				t.Run(c.name, func(t *testing.T) {
					t.Parallel()

					got := parsetest.ParsesUntilExtra(t, c.in, "", f, parsetest.WantErrors(wantError))
					should.Equal(t, got, c.want)
				})
			}
		})
	}
}

func TestString(t *testing.T) {
	t.Parallel()
	parsetest.AlsoFulfils(t, String(), testInterpretedString())
	parsetest.AlsoFulfils(t, String(), testRawString())
}

func TestInterpreted(t *testing.T) {
	t.Parallel()
	testInterpretedString()(t, InterpretedString())
}

func testInterpretedString() func(t *testing.T, f parser.Func[*ast.InterpretedString]) {
	return func(t *testing.T, f parser.Func[*ast.InterpretedString]) {
		t.Run("success", func(t *testing.T) {
			t.Parallel()

			tests := []struct {
				name string
				in   string
				want *ast.InterpretedString
			}{
				{
					name: "simple interpreted string",
					in:   `"foo"`,
					want: &ast.InterpretedString{
						Open: &ast.Position{Line: 1, Col: 1},
						Contents: []ast.InterpretedStringNode{
							&ast.InterpretedStringText{
								Text:     "foo",
								Position: &ast.Position{Line: 1, Col: 2},
							},
						},
						Close: &ast.Position{Line: 1, Col: 5},
					},
				}, {
					name: "with interpolation",
					in:   `"foo #{bar} baz"`,
					want: &ast.InterpretedString{
						Open: &ast.Position{Line: 1, Col: 1},
						Contents: []ast.InterpretedStringNode{
							&ast.InterpretedStringText{
								Text:     "foo ",
								Position: &ast.Position{Line: 1, Col: ast.Col(1 + len(`"`))},
							}, &ast.ExpressionInterpolation{
								Hash:   &ast.Position{Line: 1, Col: ast.Col(1 + len(`"foo `))},
								LBrace: &ast.Position{Line: 1, Col: ast.Col(1 + len(`"foo #`))},
								Expression: &ast.Expression{
									Nodes: ast.Code{
										&ast.GoCode{
											Code:     "bar",
											Position: &ast.Position{Line: 1, Col: ast.Col(1 + len(`"foo #{`))},
										},
									},
								},
								RBrace: &ast.Position{Line: 1, Col: ast.Col(1 + len(`"foo #{bar`))},
							}, &ast.InterpretedStringText{
								Text:     " baz",
								Position: &ast.Position{Line: 1, Col: ast.Col(1 + len(`"foo #{bar}`))},
							},
						},
						Close: &ast.Position{Line: 1, Col: ast.Col(1 + len(`"foo #{bar} baz`))},
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
				want      *ast.InterpretedString
				wantError string
			}{
				{
					name: "missing closing quote",
					in:   `"foo`,
					want: &ast.InterpretedString{
						Open: &ast.Position{Line: 1, Col: 1},
						Contents: []ast.InterpretedStringNode{
							&ast.InterpretedStringText{
								Text:     "foo",
								Position: &ast.Position{Line: 1, Col: 2},
							},
						},
					},
					wantError: "string: missing closing quote",
				},
			}

			for _, c := range tests {
				t.Run(c.name, func(t *testing.T) {
					t.Parallel()

					got := parsetest.ParsesUntilExtra(t, c.in, "", f, parsetest.WantErrors(c.wantError))
					should.Equal(t, got, c.want)
				})
			}
		})
	}
}

func TestRawString(t *testing.T) {
	t.Parallel()
	testRawString()(t, RawString())
}

func testRawString() func(t *testing.T, f parser.Func[*ast.RawString]) {
	return func(t *testing.T, f parser.Func[*ast.RawString]) {
		t.Run("success", func(t *testing.T) {
			t.Parallel()

			tests := []struct {
				name string
				in   string
				want *ast.RawString
			}{
				{
					name: "simple",
					in:   "`bar`",
					want: &ast.RawString{
						Open: &ast.Position{Line: 1, Col: 1},
						Contents: []ast.RawStringNode{
							&ast.RawStringText{
								Text:     "bar",
								Position: &ast.Position{Line: 1, Col: 2},
							},
						},
						Close: &ast.Position{Line: 1, Col: 5},
					},
				}, {
					name: "with interpolation",
					in:   "`foo #{bar} baz`",
					want: &ast.RawString{
						Open: &ast.Position{Line: 1, Col: 1},
						Contents: []ast.RawStringNode{
							&ast.RawStringText{
								Text:     "foo ",
								Position: &ast.Position{Line: 1, Col: ast.Col(1 + len("`"))},
							}, &ast.ExpressionInterpolation{
								Hash:   &ast.Position{Line: 1, Col: ast.Col(1 + len("`foo "))},
								LBrace: &ast.Position{Line: 1, Col: ast.Col(1 + len("`foo #"))},
								Expression: &ast.Expression{
									Nodes: ast.Code{
										&ast.GoCode{
											Code:     "bar",
											Position: &ast.Position{Line: 1, Col: ast.Col(1 + len("`foo #{"))},
										},
									},
								},
								RBrace: &ast.Position{Line: 1, Col: ast.Col(1 + len("`foo #{bar"))},
							}, &ast.RawStringText{
								Text:     " baz",
								Position: &ast.Position{Line: 1, Col: ast.Col(1 + len("`foo #{bar}"))},
							},
						},
						Close: &ast.Position{Line: 1, Col: ast.Col(1 + len("`foo #{bar} baz"))},
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
				want      *ast.RawString
				wantError string
			}{
				{
					name: "missing closing quote",
					in:   "`foo",
					want: &ast.RawString{
						Open: &ast.Position{Line: 1, Col: 1},
						Contents: []ast.RawStringNode{
							&ast.RawStringText{
								Text:     "foo",
								Position: &ast.Position{Line: 1, Col: 2},
							},
						},
					},
					wantError: "string: missing closing quote",
				},
			}

			for _, c := range tests {
				t.Run(c.name, func(t *testing.T) {
					t.Parallel()

					got := parsetest.ParsesUntilExtra(t, c.in, "", f, parsetest.WantErrors(c.wantError))
					should.Equal(t, got, c.want)
				})
			}
		})
	}
}
