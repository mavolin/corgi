package interpolation

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestTextInterpolation(t *testing.T) {
	t.Parallel()

	testutil.AssertAlsoFulfils(t, StringInterpolation(), testEscapedHash)
	testutil.AssertAlsoFulfils(t, StringInterpolation(), testHashSpace)
	testutil.AssertAlsoFulfils(t, StringInterpolation(), testExpressionInterpolation)
	testutil.AssertAlsoFulfils(t, StringInterpolation(), testComponentCallInterpolation)
	testutil.AssertAlsoFulfils(t, StringInterpolation(), testCharacterReference)
	testutil.AssertAlsoFulfils(t, StringInterpolation(), testExpressionInterpolation)
}

func TestStringInterpolation(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		testutil.AssertAlsoFulfils(t, StringInterpolation(), testEscapedHash)
		testutil.AssertAlsoFulfils(t, StringInterpolation(), testExpressionInterpolation)
		testutil.AssertAlsoFulfils(t, StringInterpolation(), testComponentCallInterpolation)
		testutil.AssertAlsoFulfils(t, StringInterpolation(), testCharacterReference)
	})

	t.Run("failure", func(t *testing.T) {
		t.Parallel()

		t.Run("hash space", func(t *testing.T) {
			t.Parallel()
			expect := &ast.BadInterpolation{
				Start: ast.Position{Line: 1, Col: 1},
				Until: ast.Position{Line: 1, Col: 3},
			}

			si := testutil.MatchesButError(t, "#_", StringInterpolation())
			assert.Equal(t, expect, si)
		})

		t.Run("bad interpolation", func(t *testing.T) {
			t.Parallel()
			expect := &ast.BadInterpolation{
				Start: ast.Position{Line: 1, Col: 1},
				Until: ast.Position{Line: 1, Col: 2},
			}

			si := testutil.MatchesButError(t, "#@", StringInterpolation())
			assert.Equal(t, expect, si)
		})
	})
}

func TestBadInterpolation(t *testing.T) {
	t.Parallel()

	in := "#@"
	expect := &ast.BadInterpolation{
		Start: ast.Position{Line: 1, Col: 1},
		Until: ast.Position{Line: 1, Col: 2},
	}

	p := testutil.NewParser(t, in+" 1other stuff")
	actual := testutil.AssertNoError(t, p, BadInterpolation())

	testutil.AssertPosition(t, p, 1, 2, 1)
	assert.Equal(t, expect, actual)
}

func TestEscapedHash(t *testing.T) {
	t.Parallel()
	testEscapedHash(t, EscapedHash())
}

func testEscapedHash(t *testing.T, f parser.Func[*ast.EscapedHash]) {
	in := "##"
	expect := &ast.EscapedHash{Position: ast.Position{Line: 1, Col: 1}}

	actual := testutil.ParsesFully(t, in, f)
	assert.Equal(t, expect, actual)
}

func TestHashSpace(t *testing.T) {
	t.Parallel()
	testHashSpace(t, HashSpace())
}

func testHashSpace(t *testing.T, f parser.Func[*ast.HashSpace]) {
	in := "#_"
	expect := &ast.HashSpace{Position: ast.Position{Line: 1, Col: 1}}

	actual := testutil.ParsesFully(t, in, f)
	assert.Equal(t, expect, actual)
}

func TestUnambiguousHash(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		char     rune
		noInline bool
	}{
		{char: ' ', noInline: false},
		{char: '\t', noInline: false},
		{char: '\n', noInline: true},
		{char: '\r', noInline: true},
	}

	for _, c := range testCases {
		t.Run(string(c.char), func(t *testing.T) {
			t.Parallel()

			in := "#" + string(c.char)

			if !c.noInline {
				t.Run("inline", func(t *testing.T) {
					t.Parallel()

					p := testutil.NewParser(t, in)
					p.DoInline(func() {
						testutil.AssertNoError(t, p, UnambiguousHash())
						line, col, index := testutil.CalcEnd(1, 1, 0, in)
						testutil.AssertPosition(t, p, line, col, index)
					})
				})
			}
			t.Run("no inline", func(t *testing.T) {
				testutil.ParsesFully(t, in, UnambiguousHash())
			})
		})
	}
}

func TestCharacterReference(t *testing.T) {
	t.Parallel()
	testCharacterReference(t, CharacterReference())
}

func testCharacterReference(t *testing.T, f parser.Func[*ast.CharacterReference]) {
	testCases := []struct {
		name  string
		chars string
	}{
		{name: "amp", chars: "&"},
		{name: "mdash", chars: "—"},
		{name: "euro", chars: "€"},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			in := "#" + c.name + ";"
			expect := &ast.CharacterReference{
				Name:     c.name,
				Chars:    c.chars,
				Position: ast.Position{Line: 1, Col: 1},
			}

			actual := testutil.ParsesFully(t, in, f)
			assert.Equal(t, expect, actual)
		})

	}
}

func TestExpressionInterpolation(t *testing.T) {
	t.Parallel()
	testExpressionInterpolation(t, ExpressionInterpolation())
}

func testExpressionInterpolation(t *testing.T, f parser.Func[*ast.ExpressionInterpolation]) {
	t.Parallel()
	t.Run("success", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			name   string
			in     string
			expect *ast.ExpressionInterpolation
		}{
			{
				name: "simple",
				in:   "#{1 + 1}",
				expect: &ast.ExpressionInterpolation{
					LBrace: &ast.Position{Line: 1, Col: 2},
					Expression: &ast.Expression{
						Nodes: []ast.ExpressionNode{
							&ast.GoCode{
								Code:     "1 + 1",
								Position: ast.Position{Line: 1, Col: 3},
							},
						},
					},
					RBrace:   &ast.Position{Line: 1, Col: 9},
					Position: ast.Position{Line: 1, Col: 1},
				},
			}, {
				name: "closing brace in string",
				in:   "#{`}`}",
				expect: &ast.ExpressionInterpolation{
					LBrace: &ast.Position{Line: 1, Col: 2},
					Expression: &ast.Expression{
						Nodes: []ast.ExpressionNode{
							&ast.String{
								Open:  ast.Position{Line: 1, Col: 3},
								Quote: '`',
								Contents: []ast.StringNode{
									&ast.StringText{
										Text:     "}",
										Position: ast.Position{Line: 1, Col: 4},
									},
								},
								Close: &ast.Position{Line: 1, Col: 5},
							},
						},
					},
					RBrace:   &ast.Position{Line: 1, Col: 5},
					Position: ast.Position{Line: 1, Col: 1},
				},
			}, {
				name: "format directive",
				in:   "#%1.2f{2.3}",
				expect: &ast.ExpressionInterpolation{
					FormatDirective: "1.2f",
					LBrace:          &ast.Position{Line: 1, Col: 7},
					Expression: &ast.Expression{
						Nodes: []ast.ExpressionNode{
							&ast.GoCode{
								Code:     "2.3",
								Position: ast.Position{Line: 1, Col: 8},
							},
						},
					},
					RBrace:   &ast.Position{Line: 1, Col: 11},
					Position: ast.Position{Line: 1, Col: 1},
				},
			},
		}

		for _, c := range testCases {
			t.Run(c.name, func(t *testing.T) {
				t.Parallel()
				actual := testutil.ParsesFully(t, c.in, f)
				assert.Equal(t, c.expect, actual)
			})
		}
	})

	t.Run("failure", func(t *testing.T) {
		t.Parallel()
		t.Run("no expression", func(t *testing.T) {
			t.Parallel()
			expect := &ast.ExpressionInterpolation{
				LBrace:   &ast.Position{Line: 1, Col: 2},
				RBrace:   &ast.Position{Line: 1, Col: 3},
				Position: ast.Position{Line: 1, Col: 1},
			}

			actual := testutil.MatchesButError(t, "#{}", f)
			assert.Equal(t, expect, actual)
		})
	})
}

func TestComponentCallInterpolation(t *testing.T) {
	t.Parallel()
	testComponentCallInterpolation(t, ComponentCallInterpolation())
}

func testComponentCallInterpolation(t *testing.T, f parser.Func[*ast.ComponentCallInterpolation]) {
	t.Parallel()

	testCases := []struct {
		name   string
		in     string
		expect *ast.ComponentCallInterpolation
	}{
		{
			name: "local",
			in:   "#:Component()",
			expect: &ast.ComponentCallInterpolation{
				ComponentCall: &ast.ComponentCall{
					Header: &ast.ComponentCallHeader{
						Name: &ast.Ident{
							Ident:    "Component",
							Position: ast.Position{Line: 1, Col: 3},
						},
						Args: &ast.Arguments{
							LParen: ast.Position{Line: 1, Col: 12},
							RParen: &ast.Position{Line: 1, Col: 13},
						},
					},
					Position: ast.Position{Line: 1, Col: 2},
				},
				Position: ast.Position{Line: 1, Col: 1},
			},
		}, {
			name: "external",
			in:   "#:foo.Component()",
			expect: &ast.ComponentCallInterpolation{
				ComponentCall: &ast.ComponentCall{
					Header: &ast.ComponentCallHeader{
						Name: &ast.QualifiedIdent{
							Package: ast.Ident{
								Ident:    "foo",
								Position: ast.Position{Line: 1, Col: 3},
							},
							Name: &ast.Ident{
								Ident:    "Component",
								Position: ast.Position{Line: 1, Col: 7},
							},
						},
						Args: &ast.Arguments{
							LParen: ast.Position{Line: 1, Col: 12},
							RParen: &ast.Position{Line: 1, Col: 13},
						},
					},
					Position: ast.Position{Line: 1, Col: 2},
				},
				Position: ast.Position{Line: 1, Col: 1},
			},
		}, {
			name: "args",
			in:   "#:Component(foo: 1)",
			expect: &ast.ComponentCallInterpolation{
				ComponentCall: &ast.ComponentCall{
					Header: &ast.ComponentCallHeader{
						Name: &ast.Ident{
							Ident:    "Component",
							Position: ast.Position{Line: 1, Col: 3},
						},
						Args: &ast.Arguments{
							LParen: ast.Position{Line: 1, Col: 12},
							Args: []ast.Argument{
								&ast.ComponentArgument{
									Name: &ast.Ident{
										Ident:    "foo",
										Position: ast.Position{Line: 1, Col: 13},
									},
									Colon: &ast.Position{Line: 1, Col: 16},
									Value: &ast.Expression{
										Nodes: []ast.ExpressionNode{
											&ast.GoCode{
												Code:     "1",
												Position: ast.Position{Line: 1, Col: 18},
											},
										},
									},
									Position: ast.Position{Line: 1, Col: 13},
								},
							},
							RParen: &ast.Position{Line: 1, Col: 13},
						},
					},
					Position: ast.Position{Line: 1, Col: 2},
				},
				Position: ast.Position{Line: 1, Col: 1},
			},
		}, {
			name: "underscore block",
			in:   "#:Component()[foo]",
			expect: &ast.ComponentCallInterpolation{
				ComponentCall: &ast.ComponentCall{
					Header: &ast.ComponentCallHeader{
						Name: &ast.Ident{
							Ident:    "Component",
							Position: ast.Position{Line: 1, Col: 3},
						},
						Args: &ast.Arguments{
							LParen: ast.Position{Line: 1, Col: 12},
							RParen: &ast.Position{Line: 1, Col: 13},
						},
					},
					Body: &ast.UnderscoreBlockShorthand{
						Implicit: true,
						Body: &ast.BracketText{
							LBracket: ast.Position{Line: 1, Col: 14},
							Lines: ast.TextBlock{
								ast.TextLine{
									&ast.Text{
										Text:     "foo",
										Position: ast.Position{Line: 1, Col: 15},
									},
								},
							},
							RBracket: &ast.Position{Line: 1, Col: 15},
						},
					},
					Position: ast.Position{Line: 1, Col: 2},
				},
				Position: ast.Position{Line: 1, Col: 1},
			},
		},
	}

	for _, c := range testCases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			actual := testutil.ParsesFully(t, c.in, f)
			assert.Equal(t, c.expect, actual)
		})
	}
}
