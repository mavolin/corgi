package interpolation

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/test/should"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

func TestTextInterpolation(t *testing.T) {
	t.Parallel()

	parsetest.AssertAlsoFulfils(t, TextInterpolation(), testTextCharacterEscape)
	parsetest.AssertAlsoFulfils(t, TextInterpolation(), testExpressionInterpolation)
	parsetest.AssertAlsoFulfils(t, TextInterpolation(), testComponentCallInterpolation)
	parsetest.AssertAlsoFulfils(t, TextInterpolation(), testCharacterReference)
	parsetest.AssertAlsoFulfils(t, TextInterpolation(), testExpressionInterpolation)
	parsetest.AssertAlsoFulfils(t, TextInterpolation(), testElementInterpolation)
}

func TestStringInterpolation(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		parsetest.AssertAlsoFulfils(t, StringInterpolation(), testTextCharacterEscape)
		parsetest.AssertAlsoFulfils(t, StringInterpolation(), testExpressionInterpolation)
		parsetest.AssertAlsoFulfils(t, StringInterpolation(), testComponentCallInterpolation)
		parsetest.AssertAlsoFulfils(t, StringInterpolation(), testCharacterReference)
	})

	t.Run("failure", func(t *testing.T) {
		t.Parallel()

		t.Run("hash space", func(t *testing.T) {
			t.Parallel()
			want := &ast.BadInterpolation{
				From:  ast.Position{Line: 1, Col: 1},
				Until: ast.Position{Line: 1, Col: 3},
			}

			si := parsetest.MatchesButError(t, "#_", StringInterpolation())
			should.Equal(t, si, ast.StringInterpolation(want))
		})

		t.Run("bad interpolation", func(t *testing.T) {
			t.Parallel()
			want := &ast.BadInterpolation{
				From:  ast.Position{Line: 1, Col: 1},
				Until: ast.Position{Line: 1, Col: 2},
			}

			si := parsetest.MatchesButError(t, "#@", StringInterpolation())
			should.Equal(t, si, ast.StringInterpolation(want))
		})
	})
}

func TestBadInterpolation(t *testing.T) {
	t.Parallel()

	in := "#@"
	want := &ast.BadInterpolation{
		From:  ast.Position{Line: 1, Col: 1},
		Until: ast.Position{Line: 1, Col: 2},
	}

	p := parsetest.NewParser(t, in+" 1other stuff")
	got := parsetest.AssertNoError(t, p, BadInterpolation())

	parsetest.AssertPosition(t, p, 1, 2, 1)
	should.Equal(t, got, want)
}

func TestStringCharacterEscape(t *testing.T) {
	t.Parallel()
	testStringCharacterEscape(t, StringCharacterEscape())
}

func testStringCharacterEscape(t *testing.T, f parser.Func[*ast.CharacterEscape]) {
	tests := []struct {
		Symbol rune
		Rune   rune
	}{
		{'#', '#'},
	}

	for _, c := range tests {
		t.Run(string(c.Symbol), func(t *testing.T) {
			t.Parallel()
			in := "#" + string(c.Symbol)
			want := &ast.CharacterEscape{
				Hash:   &ast.Position{Line: 1, Col: 1},
				Symbol: c.Symbol,
				Rune:   c.Rune,
			}

			got := parsetest.ParsesFully(t, in, f)
			should.Equal(t, got, want)
		})
	}

}

func TestTextCharacterEscape(t *testing.T) {
	t.Parallel()
	testTextCharacterEscape(t, TextCharacterEscape())
}

func testTextCharacterEscape(t *testing.T, f parser.Func[*ast.CharacterEscape]) {
	t.Parallel()

	tests := []struct {
		Symbol rune
		Rune   rune
	}{
		{'#', '#'},
		{'_', ' '},
		{']', ']'},
	}

	for _, c := range tests {
		t.Run(string(c.Symbol), func(t *testing.T) {
			t.Parallel()
			in := "#" + string(c.Symbol)
			want := &ast.CharacterEscape{
				Hash:   &ast.Position{Line: 1, Col: 1},
				Symbol: c.Symbol,
				Rune:   c.Rune,
			}

			got := parsetest.ParsesFully(t, in, f)
			should.Equal(t, got, want)
		})
	}
}

func TestVerbatimTextCharacterEscape(t *testing.T) {
	t.Parallel()

	tests := []struct {
		Symbol rune
		Rune   rune
	}{
		{']', ']'},
	}

	for _, c := range tests {
		t.Run(string(c.Symbol), func(t *testing.T) {
			t.Parallel()
			in := "#" + string(c.Symbol)
			want := &ast.CharacterEscape{
				Hash:   &ast.Position{Line: 1, Col: 1},
				Symbol: c.Symbol,
				Rune:   c.Rune,
			}

			got := parsetest.ParsesFully(t, in, VerbatimTextCharacterEscape())
			should.Equal(t, got, want)
		})
	}
}

func TestUnambiguousHash(t *testing.T) {
	t.Parallel()

	tests := []struct {
		char     rune
		noInline bool
	}{
		{char: ' ', noInline: false},
		{char: '\t', noInline: false},
		{char: '\n', noInline: true},
		{char: '\r', noInline: true},
	}

	for _, c := range tests {
		t.Run(string(c.char), func(t *testing.T) {
			t.Parallel()

			in := "#" + string(c.char)

			if !c.noInline {
				t.Run("inline", func(t *testing.T) {
					t.Parallel()

					p := parsetest.NewParser(t, in)
					p.DoInline(func() {
						parsetest.AssertNoError(t, p, UnambiguousHash())
						line, col, index := parsetest.CalcEnd(1, 1, 0, in)
						parsetest.AssertPosition(t, p, line, col, index)
					})
				})
			}
			t.Run("no inline", func(t *testing.T) {
				parsetest.ParsesFully(t, in, UnambiguousHash())
			})
		})
	}
}

func TestCharacterReference(t *testing.T) {
	t.Parallel()
	testCharacterReference(t, CharacterReference())
}

func testCharacterReference(t *testing.T, f parser.Func[*ast.CharacterReference]) {
	tests := []struct {
		name  string
		chars string
	}{
		{name: "amp", chars: "&"},
		{name: "mdash", chars: "—"},
		{name: "euro", chars: "€"},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			in := "#" + c.name + ";"
			want := &ast.CharacterReference{
				Name:  c.name,
				Chars: c.chars,
				Hash:  &ast.Position{Line: 1, Col: 1},
			}

			got := parsetest.ParsesFully(t, in, f)
			should.Equal(t, got, want)
		})
	}
}

func TestElementInterpolation(t *testing.T) {
	t.Parallel()
	testElementInterpolation(t, ElementInterpolation())
}

func testElementInterpolation(t *testing.T, f parser.Func[*ast.ElementInterpolation]) {
	tests := []struct {
		name string
		in   string
		want *ast.ElementInterpolation
	}{
		{
			name: "no body",
			in:   "#br",
			want: &ast.ElementInterpolation{
				Element: &ast.Element{
					Header: &ast.ElementHeader{
						Name: &ast.ElementReference{
							Name: &ast.ElementName{
								Name:     "br",
								Position: &ast.Position{Line: 1, Col: 2},
							},
						},
					},
				},
				Hash: &ast.Position{Line: 1, Col: 1},
			},
		}, {
			name: "body",
			in:   "#strong[woof]",
			want: &ast.ElementInterpolation{
				Element: &ast.Element{
					Header: &ast.ElementHeader{
						Name: &ast.ElementReference{
							Name: &ast.ElementName{
								Name:     "strong",
								Position: &ast.Position{Line: 1, Col: 2},
							},
						},
					},
					Body: &ast.BracketText{
						LBracket: &ast.Position{Line: 1, Col: 8},
						Lines: ast.TextBlock{
							ast.TextLine{
								&ast.Text{
									Text:     "woof",
									Position: &ast.Position{Line: 1, Col: 9},
								},
							},
						},
						RBracket: &ast.Position{Line: 1, Col: 13},
					},
				},
				Hash: &ast.Position{Line: 1, Col: 1},
			},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got := parsetest.ParsesFully(t, c.in, f)
			should.Equal(t, got, c.want)
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

		in := "#{1 + 1}"
		want := &ast.ExpressionInterpolation{
			LBrace: &ast.Position{Line: 1, Col: 2},
			Expression: &ast.Expression{
				Nodes: ast.Code{
					&ast.GoCode{
						Code:     "1 + 1",
						Position: &ast.Position{Line: 1, Col: 3},
					},
				},
			},
			RBrace: &ast.Position{Line: 1, Col: 8},
			Hash:   &ast.Position{Line: 1, Col: 1},
		}

		got := parsetest.ParsesFully(t, in, f)
		should.Equal(t, got, want)
	})

	t.Run("failure", func(t *testing.T) {
		t.Parallel()
		t.Run("no expression", func(t *testing.T) {
			t.Parallel()
			want := &ast.ExpressionInterpolation{
				LBrace: &ast.Position{Line: 1, Col: 2},
				RBrace: &ast.Position{Line: 1, Col: 3},
				Hash:   &ast.Position{Line: 1, Col: 1},
			}

			got := parsetest.MatchesButError(t, "#{}", f)
			should.Equal(t, got, want)
		})
	})
}

func TestComponentCallInterpolation(t *testing.T) {
	t.Parallel()
	testComponentCallInterpolation(t, ComponentCallInterpolation())
}

func testComponentCallInterpolation(t *testing.T, f parser.Func[*ast.ComponentCallInterpolation]) {
	tests := []struct {
		name string
		in   string
		want *ast.ComponentCallInterpolation
	}{
		{
			name: "local",
			in:   "#:component()",
			want: &ast.ComponentCallInterpolation{
				ComponentCall: &ast.ComponentCall{
					Colon: &ast.Position{Line: 1, Col: 2},
					Header: &ast.ComponentCallHeader{
						Name: &ast.Identifier{
							Name:     "component",
							Position: &ast.Position{Line: 1, Col: 3},
						},
						Arguments: &ast.Arguments{
							LParen: &ast.Position{Line: 1, Col: 12},
							RParen: &ast.Position{Line: 1, Col: 13},
						},
					},
				},
				Hash: &ast.Position{Line: 1, Col: 1},
			},
		}, {
			name: "default block",
			in:   "#:component()[foo]",
			want: &ast.ComponentCallInterpolation{
				ComponentCall: &ast.ComponentCall{
					Colon: &ast.Position{Line: 1, Col: 2},
					Header: &ast.ComponentCallHeader{
						Name: &ast.Identifier{
							Name:     "component",
							Position: &ast.Position{Line: 1, Col: 3},
						},
						Arguments: &ast.Arguments{
							LParen: &ast.Position{Line: 1, Col: 12},
							RParen: &ast.Position{Line: 1, Col: 13},
						},
					},
					Body: &ast.DefaultBlockShorthand{
						Implicit: true,
						Body: &ast.BracketText{
							LBracket: &ast.Position{Line: 1, Col: 14},
							Lines: ast.TextBlock{
								ast.TextLine{
									&ast.Text{
										Text:     "foo",
										Position: &ast.Position{Line: 1, Col: 15},
									},
								},
							},
							RBracket: &ast.Position{Line: 1, Col: 18},
						},
						Position: &ast.Position{Line: 1, Col: 14},
					},
				},
				Hash: &ast.Position{Line: 1, Col: 1},
			},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got := parsetest.ParsesFully(t, c.in, f)
			should.Equal(t, got, c.want)
		})
	}
}
