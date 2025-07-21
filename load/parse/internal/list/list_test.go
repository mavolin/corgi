package list

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/internal/test/should"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/parsetest"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

func TestParenList(t *testing.T) {
	t.Parallel()
	testList(t, "paren list", '(', ')')
}

func TestBracketList(t *testing.T) {
	t.Parallel()
	testList(t, "bracket list", '[', ']')
}

func testList(t *testing.T, name string, opening, closing rune) {
	elemFunc := func(p *parser.Parser) (string, *diagnostic.Diagnostic) {
		s := parser.TokenWhile(p, func() bool {
			return parser.MatchesRunePredicate(p, func(r rune) bool {
				return r >= 'a' && r <= 'z'
			})
		})
		if s == "" {
			return "", &diagnostic.Diagnostic{
				Message: "missing test element",
				Primary: quickanno.Expected(p, p.Pos(), "a test element"),
			}
		}
		return s, nil
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
						Line: strings.Count(in, "\n") + 1,
						Col:  len(string(opening)) + len(c.elems) + 1,
					},
					Elems: c.want,
				}
				if want.Close.Line > 1 {
					want.Close.Col = len(in[strings.LastIndex(in, "\n")+1:])
				}

				got := parsetest.ParsesFully(t, in, list(name, opening, closing, elemFunc))
				should.Equal(t, want, got)
			})
		}
	})

	recoverCases := []struct {
		name    string
		elems   string
		noClose bool
		want    []string
	}{
		{
			name:  "empty elem",
			elems: "foo, , baz",
			want:  []string{"foo", "", "baz"},
		}, {
			name:    "unclosed",
			elems:   "foo",
			noClose: true,
			want:    []string{"foo"},
		}, {
			name:  "unexpected after elem",
			elems: "foo 123, baz",
			want:  []string{"foo", "baz"},
		}, {
			name:  "no elem match",
			elems: "foo, 123, baz",
			want:  []string{"foo", "", "baz"},
		}, {
			name:  "missing comma",
			elems: "foo bar baz, qux",
			want:  []string{"foo", "bar", "baz", "qux"},
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
					Close: &ast.Position{Line: 1, Col: len(string(opening)) + len(c.elems) + 1},
					Elems: c.want,
				}
				if c.noClose {
					in = in[:len(in)-1]
					want.Close = nil
				}

				gotIn := in
				if !c.noClose {
					gotIn += " other"
				}
				p := parsetest.NewParser(t, gotIn)
				got := parsetest.AssertMatchesButError(t, p, list(name, opening, closing, elemFunc))
				if should.Equal(t, want, got) {
					should.Equal(t, p.Index(), len(in))
				}
			})
		}
	})
}

func TestCommaList(t *testing.T) {
	t.Parallel()

	elemFunc := func(p *parser.Parser) (string, *diagnostic.Diagnostic) {
		s := parser.TokenWhile(p, func() bool {
			return parser.MatchesRunePredicate(p, func(r rune) bool {
				return r >= 'a' && r <= 'z'
			})
		})
		if s == "" {
			return "", &diagnostic.Diagnostic{
				Message: "missing test element",
				Primary: quickanno.Expected(p, p.Pos(), "a test element"),
			}
		}
		return s, nil
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

				p := parsetest.NewParser(t, c.in+" other")
				got := parsetest.AssertNoError(t, p, CommaList("a", "as", elemFunc))
				if should.Equal(t, c.want, got) {
					should.Equal(t, len(c.in), p.Index())
				}
			})
		}
	})

	recoverCases := []struct {
		name string
		in   string
		want []string
	}{
		{
			name: "empty elem",
			in:   "foo, , baz",
			want: []string{"foo", "", "baz"},
		}, {
			name: "no elems but comma",
			in:   ",",
			want: []string{"", ""},
		},
	}

	t.Run("recover", func(t *testing.T) {
		t.Parallel()

		for _, c := range recoverCases {
			t.Run(c.name, func(t *testing.T) {
				t.Parallel()

				p := parsetest.NewParser(t, c.in+" 123")
				got := parsetest.AssertMatchesButError(t, p, CommaList("a", "as", elemFunc))
				if should.Equal(t, c.want, got) {
					should.Equal(t, len(c.in), p.Index())
				}
			})
		}
	})
}
