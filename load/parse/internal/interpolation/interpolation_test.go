package interpolation

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/should"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
	"github.com/mavolin/corgi/v2/load/parse/internal/whitespace"
)

func TestTextInterpolation(t *testing.T) {
	t.Parallel()

	parsetest.AlsoFulfils(t, TextInterpolation(), testTextCharacterEscape)
	parsetest.AlsoFulfils(t, TextInterpolation(), testExpressionInterpolation)
	parsetest.AlsoFulfils(t, TextInterpolation(), testCharacterReference)
	parsetest.AlsoFulfils(t, TextInterpolation(), testExpressionInterpolation)
}

func TestStringInterpolation(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		parsetest.AlsoFulfils(t, StringInterpolation(), testStringCharacterEscape)
		parsetest.AlsoFulfils(t, StringInterpolation(), testExpressionInterpolation)
		parsetest.AlsoFulfils(t, StringInterpolation(), testComponentCallInterpolation)
		parsetest.AlsoFulfils(t, StringInterpolation(), testCharacterReference)
	})

	t.Run("failure", func(t *testing.T) {
		t.Parallel()

		t.Run("hash space", func(t *testing.T) {
			t.Parallel()
			want := &ast.BadInterpolation{
				From:  ast.Position{Line: 1, Col: 1},
				Until: ast.Position{Line: 1, Col: 3},
			}
			wantError := "string interpolation: cannot use hash space here"

			got := parsetest.ParsesExact(t, "#_", StringInterpolation(), parsetest.WantErrors(wantError))
			should.Equal(t, got, ast.StringInterpolation(want))
		})

		t.Run("bad interpolation", func(t *testing.T) {
			t.Parallel()
			want := &ast.BadInterpolation{
				From:  ast.Position{Line: 1, Col: 1},
				Until: ast.Position{Line: 1, Col: 2},
			}
			wantError := "bad interpolation"

			got := parsetest.ParsesUntilExtra(t, "#", "@", StringInterpolation(), parsetest.WantErrors(wantError))
			should.Equal(t, got, ast.StringInterpolation(want))
		})
	})
}

func TestBadInterpolation(t *testing.T) {
	t.Parallel()

	want := &ast.BadInterpolation{
		From:  ast.Position{Line: 1, Col: 1},
		Until: ast.Position{Line: 1, Col: 2},
	}

	got := parsetest.ParsesUntilExtra(t, "#", "@", BadInterpolation())
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

			got := parsetest.ParsesExact(t, in, f)
			should.Equal(t, got, want)
		})
	}
}

func TestTextCharacterEscape(t *testing.T) {
	t.Parallel()
	testTextCharacterEscape(t, TextCharacterEscape())
}

func testTextCharacterEscape(t *testing.T, f parser.Func[*ast.CharacterEscape]) {
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

			got := parsetest.ParsesExact(t, in, f)
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

			got := parsetest.ParsesExact(t, in, VerbatimTextCharacterEscape())
			should.Equal(t, got, want)
		})
	}
}

func TestUnambiguousHash(t *testing.T) {
	t.Parallel()

	for _, c := range whitespace.Runes {
		t.Run(string(c), func(t *testing.T) {
			t.Parallel()

			got := parsetest.ParsesUntilExtra(t, "#", string(c)+"other stuff", UnambiguousHash())
			should.True(t, got)
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

			got := parsetest.ParsesExact(t, in, f)
			should.Equal(t, got, want)
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

		got := parsetest.ParsesExact(t, in, f)
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
			wantError := "expression interpolation: missing expression"

			got := parsetest.ParsesExact(t, "#{}", f, parsetest.WantErrors(wantError))
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
					Colon: &ast.Position{Line: 1, Col: ast.Col(1 + len("#"))},
					Header: &ast.ComponentCallHeader{
						Name: &ast.Identifier{
							Name:     "component",
							Position: &ast.Position{Line: 1, Col: ast.Col(1 + len("#:"))},
						},
						Arguments: &ast.Arguments{
							LParen: &ast.Position{Line: 1, Col: ast.Col(1 + len("#:component"))},
							RParen: &ast.Position{Line: 1, Col: ast.Col(1 + len("#:component("))},
						},
					},
				},
				Hash: &ast.Position{Line: 1, Col: 1},
			},
		}, {
			name: "default block",
			in:   "#:component() {}",
			want: &ast.ComponentCallInterpolation{
				ComponentCall: &ast.ComponentCall{
					Colon: &ast.Position{Line: 1, Col: ast.Col(1 + len("#"))},
					Header: &ast.ComponentCallHeader{
						Name: &ast.Identifier{
							Name:     "component",
							Position: &ast.Position{Line: 1, Col: ast.Col(1 + len("#:"))},
						},
						Arguments: &ast.Arguments{
							LParen: &ast.Position{Line: 1, Col: ast.Col(1 + len("#:component"))},
							RParen: &ast.Position{Line: 1, Col: ast.Col(1 + len("#:component("))},
						},
					},
					Body: &ast.Scope{
						LBrace: &ast.Position{Line: 1, Col: ast.Col(1 + len("#:component() "))},
						RBrace: &ast.Position{Line: 1, Col: ast.Col(1 + len("#:component() {"))},
					},
				},
				Hash: &ast.Position{Line: 1, Col: 1},
			},
		},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			got := parsetest.ParsesExact(t, c.in, f)
			should.Equal(t, got, c.want)
		})
	}
}
