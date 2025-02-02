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

	testutil.AssertAlsoFulfils(t, TextInterpolation(), testEscapedHash)
	testutil.AssertAlsoFulfils(t, TextInterpolation(), testHashSpace)
	testutil.AssertAlsoFulfils(t, TextInterpolation(), testExpressionInterpolation)
	testutil.AssertAlsoFulfils(t, TextInterpolation(), testComponentCallInterpolation)
	testutil.AssertAlsoFulfils(t, TextInterpolation(), testCharacterReference)
	testutil.AssertAlsoFulfils(t, TextInterpolation(), testExpressionInterpolation)
	testutil.AssertAlsoFulfils(t, TextInterpolation(), testElementInterpolation)
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
				Name:  c.name,
				Chars: c.chars,
				Hash:  ast.Position{Line: 1, Col: 1},
			}

			actual := testutil.ParsesFully(t, in, f)
			assert.Equal(t, expect, actual)
		})

	}
}

func TestElementInterpolation(t *testing.T) {
	t.Parallel()
	testElementInterpolation(t, ElementInterpolation())
}

func testElementInterpolation(t *testing.T, f parser.Func[*ast.ElementInterpolation]) {
	testCases := []struct {
		name   string
		in     string
		expect *ast.ElementInterpolation
	}{
		{
			name: "no body",
			in:   "#br",
			expect: &ast.ElementInterpolation{
				Element: &ast.Element{
					Header: ast.ElementHeader{
						Name: ast.ElementName{
							Name:     "br",
							Position: ast.Position{Line: 1, Col: 2},
						},
					},
				},
				Hash: ast.Position{Line: 1, Col: 1},
			},
		}, {
			name: "body",
			in:   "#strong[woof]",
			expect: &ast.ElementInterpolation{
				Element: &ast.Element{
					Header: ast.ElementHeader{
						Name: ast.ElementName{
							Name:     "strong",
							Position: ast.Position{Line: 1, Col: 2},
						},
					},
					Body: &ast.BracketText{
						LBracket: ast.Position{Line: 1, Col: 8},
						Lines: ast.TextBlock{
							ast.TextLine{
								&ast.Text{
									Text:     "woof",
									Position: ast.Position{Line: 1, Col: 9},
								},
							},
						},
						RBracket: &ast.Position{Line: 1, Col: 13},
					},
				},
				Hash: ast.Position{Line: 1, Col: 1},
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

func TestExpressionInterpolation(t *testing.T) {
	t.Parallel()
	testExpressionInterpolation(t, ExpressionInterpolation())
}

func testExpressionInterpolation(t *testing.T, f parser.Func[*ast.ExpressionInterpolation]) {
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
						Code: ast.Code{
							&ast.GoCode{
								Code:     "1 + 1",
								Position: ast.Position{Line: 1, Col: 3},
							},
						},
					},
					RBrace: &ast.Position{Line: 1, Col: 8},
					Hash:   ast.Position{Line: 1, Col: 1},
				},
			}, {
				name: "format directive",
				in:   "#%1.2f{2.3}",
				expect: &ast.ExpressionInterpolation{
					FormatDirective: "1.2f",
					LBrace:          &ast.Position{Line: 1, Col: 7},
					Expression: &ast.Expression{
						Code: ast.Code{
							&ast.GoCode{
								Code:     "2.3",
								Position: ast.Position{Line: 1, Col: 8},
							},
						},
					},
					RBrace: &ast.Position{Line: 1, Col: 11},
					Hash:   ast.Position{Line: 1, Col: 1},
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
				LBrace: &ast.Position{Line: 1, Col: 2},
				RBrace: &ast.Position{Line: 1, Col: 3},
				Hash:   ast.Position{Line: 1, Col: 1},
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
	testCases := []struct {
		name   string
		in     string
		expect *ast.ComponentCallInterpolation
	}{
		{
			name: "local",
			in:   "#:component()",
			expect: &ast.ComponentCallInterpolation{
				ComponentCall: &ast.ComponentCall{
					Header: ast.ComponentCallHeader{
						Colon: ast.Position{Line: 1, Col: 2},
						Name: &ast.Ident{
							Ident:    "component",
							Position: ast.Position{Line: 1, Col: 3},
						},
						Args: &ast.Arguments{
							LParen: ast.Position{Line: 1, Col: 12},
							RParen: &ast.Position{Line: 1, Col: 13},
						},
					},
				},
				Hash: ast.Position{Line: 1, Col: 1},
			},
		}, {
			name: "underscore block",
			in:   "#:component()[foo]",
			expect: &ast.ComponentCallInterpolation{
				ComponentCall: &ast.ComponentCall{
					Header: ast.ComponentCallHeader{
						Colon: ast.Position{Line: 1, Col: 2},
						Name: &ast.Ident{
							Ident:    "component",
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
							RBracket: &ast.Position{Line: 1, Col: 18},
						},
						Position: ast.Position{Line: 1, Col: 14},
					},
				},
				Hash: ast.Position{Line: 1, Col: 1},
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
