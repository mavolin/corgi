package list

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/internal/should"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
)

func TestParenList(t *testing.T) {
	t.Parallel()
	testList(t, '(', ')', ParenList)
}

func TestBracketList(t *testing.T) {
	t.Parallel()
	testList(t, '[', ']', BracketList)
}

func testList(t *testing.T, opening, closing rune, f func(belongsTo, singular, plural string, elemFunc parser.Func[string]) parser.Func[*List[string]]) {
	elemFunc := func(p *parser.Parser) string {
		return parser.TokenWhileRunePredicate(p, func(r rune) bool {
			return r >= 'a' && r <= 'z'
		})
	}

	successCases := []struct {
		name  string
		elems string
		want  []string
	}{
		{
			name:  "empty",
			elems: "",
			want:  nil,
		}, {
			name:  "single",
			elems: "foo",
			want:  []string{"foo"},
		}, {
			name:  "multiple",
			elems: "foo, bar, baz",
			want:  []string{"foo", "bar", "baz"},
		}, {
			name:  "trailing comma",
			elems: "foo,bar,baz,\n",
			want:  []string{"foo", "bar", "baz"},
		}, {
			name:  "multiline",
			elems: "\n  \nfoo\t ,\nbar  ,\n\nbaz",
			want:  []string{"foo", "bar", "baz"},
		},
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		for _, c := range successCases {
			t.Run(c.name, func(t *testing.T) {
				t.Parallel()

				in := fmt.Sprintf("%c%s%c", opening, c.elems, closing)
				want := &List[string]{
					Open: &ast.Position{Line: 1, Col: 1},
					Close: &ast.Position{
						Line: ast.Line(strings.Count(in, "\n") + 1),            //nolint:gosec
						Col:  ast.Col(len(string(opening)) + len(c.elems) + 1), //nolint:gosec
					},
					Elems: c.want,
				}
				if want.Close.Line > 1 {
					want.Close.Col = ast.Col(len(in[strings.LastIndex(in, "\n")+1:])) //nolint:gosec
				}

				got := parsetest.ParsesExact(t, in, f("belongsTo", "singular", "plural", elemFunc))
				should.Equal(t, got, want)
			})
		}
	})

	recoverCases := []struct {
		name       string
		elems      string
		noClose    bool
		want       []string
		wantErrors []string
	}{
		{
			name:       "empty elem",
			elems:      "foo, , baz",
			want:       []string{"foo", "", "baz"},
			wantErrors: []string{"belongsTo: missing singular"},
		}, {
			name:       "unclosed",
			elems:      "foo",
			noClose:    true,
			want:       []string{"foo"},
			wantErrors: []string{"belongsTo: unclosed plural"},
		}, {
			name:       "unexpected after elem",
			elems:      "foo 123, baz",
			want:       []string{"foo", "baz"},
			wantErrors: []string{"belongsTo: unexpected runes after singular"},
		}, {
			name:       "no elem match",
			elems:      "foo, 123, baz",
			want:       []string{"foo", "", "baz"},
			wantErrors: []string{"belongsTo: missing singular"},
		}, {
			name:  "missing comma",
			elems: "foo bar baz, qux",
			want:  []string{"foo", "bar", "baz", "qux"},
			wantErrors: []string{
				"belongsTo: missing comma",
				"belongsTo: missing comma",
			},
		},
	}

	t.Run("recover", func(t *testing.T) {
		t.Parallel()

		for _, c := range recoverCases {
			t.Run(c.name, func(t *testing.T) {
				t.Parallel()

				in := fmt.Sprintf("%c%s%c", opening, c.elems, closing)
				want := &List[string]{
					Open:  &ast.Position{Line: 1, Col: 1},
					Close: &ast.Position{Line: 1, Col: ast.Col(len(string(opening)) + len(c.elems) + 1)}, //nolint:gosec
					Elems: c.want,
				}
				if c.noClose {
					in = in[:len(in)-1]
					want.Close = nil
				}

				extra := " "
				if !c.noClose {
					extra = " other"
				}

				got := parsetest.ParsesUntilExtra(t, in, extra, list("belongsTo", "singular", "plural", opening, closing, elemFunc),
					parsetest.WantErrors(c.wantErrors...))
				should.Equal(t, got, want)
			})
		}
	})
}

func TestCommaList(t *testing.T) {
	t.Parallel()

	elemFunc := func(p *parser.Parser) string {
		return parser.TokenWhileRunePredicate(p, func(r rune) bool {
			return r >= 'a' && r <= 'z'
		})
	}

	successCases := []struct {
		name string
		in   string
		want []string
	}{
		{
			name: "single",
			in:   "foo",
			want: []string{"foo"},
		}, {
			name: "multiple",
			in:   "foo, bar, baz",
			want: []string{"foo", "bar", "baz"},
		}, {
			name: "multiline",
			in:   "foo\t ,\nbar  ,\n\nbaz",
			want: []string{"foo", "bar", "baz"},
		},
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		for _, c := range successCases {
			t.Run(c.name, func(t *testing.T) {
				t.Parallel()

				got := parsetest.ParsesExact(t, c.in, CommaList("belongsTo", "singular", "plural", elemFunc))
				should.Equal(t, got, c.want)
			})
		}
	})

	recoverCases := []struct {
		name       string
		in         string
		want       []string
		wantErrors []string
	}{
		{
			name:       "empty elem",
			in:         "foo, , baz",
			want:       []string{"foo", "", "baz"},
			wantErrors: []string{"belongsTo: missing singular"},
		}, {
			name:       "no elems but comma",
			in:         ",",
			want:       []string{"", ""},
			wantErrors: []string{"belongsTo: missing singular", "belongsTo: missing singular"},
		},
	}

	t.Run("recover", func(t *testing.T) {
		t.Parallel()

		for _, c := range recoverCases {
			t.Run(c.name, func(t *testing.T) {
				t.Parallel()

				got := parsetest.ParsesUntilEOS(t, c.in, CommaList("belongsTo", "singular", "plural", elemFunc),
					parsetest.WantErrors(c.wantErrors...))
				should.Equal(t, got, c.want)
			})
		}
	})
}
