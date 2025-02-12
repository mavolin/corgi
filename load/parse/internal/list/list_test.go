package list

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestParenList(t *testing.T) {
	t.Parallel()
	testList(t, "paren list", '(', ')')
}

func TestBracketList(t *testing.T) {
	t.Parallel()
	testList(t, "bracket list", '[', ']')
}

func testList(t *testing.T, name string, open, close rune) {
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
		name   string
		elems  string
		expect []string
	}{
		{
			name:   "empty",
			elems:  "",
			expect: nil,
		}, {
			name:   "single",
			elems:  "foo",
			expect: []string{"foo"},
		}, {
			name:   "multiple",
			elems:  "foo, bar, baz",
			expect: []string{"foo", "bar", "baz"},
		}, {
			name:   "trailing comma",
			elems:  "foo,bar,baz,\n",
			expect: []string{"foo", "bar", "baz"},
		}, {
			name:   "multiline",
			elems:  "\n  \nfoo\t ,\nbar  ,\n\nbaz",
			expect: []string{"foo", "bar", "baz"},
		},
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		for _, c := range successCases {
			t.Run(c.name, func(t *testing.T) {
				t.Parallel()

				in := fmt.Sprintf("%c%s%c", open, c.elems, close)
				expect := &List[string]{
					Open: &ast.Position{Line: 1, Col: 1},
					Close: &ast.Position{
						Line: strings.Count(in, "\n") + 1,
						Col:  len(string(open)) + len(c.elems) + 1,
					},
					Elems: c.expect,
				}
				if expect.Close.Line > 1 {
					expect.Close.Col = len(in[strings.LastIndex(in, "\n")+1:])
				}

				actual := testutil.ParsesFully(t, in, list(name, open, close, elemFunc))
				assert.Equal(t, expect, actual)
			})
		}
	})

	recoverCases := []struct {
		name    string
		elems   string
		noClose bool
		expect  []string
	}{
		{
			name:   "empty elem",
			elems:  "foo, , baz",
			expect: []string{"foo", "", "baz"},
		}, {
			name:    "unclosed",
			elems:   "foo",
			noClose: true,
			expect:  []string{"foo"},
		}, {
			name:   "unexpected after elem",
			elems:  "foo 123, baz",
			expect: []string{"foo", "baz"},
		}, {
			name:   "no elem match",
			elems:  "foo, 123, baz",
			expect: []string{"foo", "", "baz"},
		}, {
			name:   "missing comma",
			elems:  "foo bar baz, qux",
			expect: []string{"foo", "bar", "baz", "qux"},
		},
	}

	t.Run("recover", func(t *testing.T) {
		t.Parallel()

		for _, c := range recoverCases {
			t.Run(c.name, func(t *testing.T) {
				t.Parallel()

				in := fmt.Sprintf("%c%s%c", open, c.elems, close)
				expect := &List[string]{
					Open:  &ast.Position{Line: 1, Col: 1},
					Close: &ast.Position{Line: 1, Col: len(string(open)) + len(c.elems) + 1},
					Elems: c.expect,
				}
				if c.noClose {
					in = in[:len(in)-1]
					expect.Close = nil
				}

				actualIn := in
				if !c.noClose {
					actualIn += " other"
				}
				p := testutil.NewParser(t, in)
				actual := testutil.AssertMatchesButError(t, p, list(name, open, close, elemFunc))
				if assert.Equal(t, expect, actual) {
					assert.Equal(t, p.Index(), len(in))
				}
			})
		}
	})
}

func TestCommaList(t *testing.T) {
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
		name   string
		in     string
		expect []string
	}{
		{
			name:   "single",
			in:     "foo",
			expect: []string{"foo"},
		}, {
			name:   "multiple",
			in:     "foo, bar, baz",
			expect: []string{"foo", "bar", "baz"},
		}, {
			name:   "multiline",
			in:     "foo\t ,\nbar  ,\n\nbaz",
			expect: []string{"foo", "bar", "baz"},
		},
	}

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		for _, c := range successCases {
			t.Run(c.name, func(t *testing.T) {
				t.Parallel()

				p := testutil.NewParser(t, c.in+" other")
				actual := testutil.AssertNoError(t, p, CommaList("a", "as", elemFunc))
				if assert.Equal(t, c.expect, actual) {
					assert.Equal(t, p.Index(), len(c.in))
				}
			})
		}
	})

	recoverCases := []struct {
		name   string
		in     string
		expect []string
	}{
		{
			name:   "empty elem",
			in:     "foo, , baz",
			expect: []string{"foo", "", "baz"},
		}, {
			name:   "no elems but comma",
			in:     ",",
			expect: []string{"", ""},
		},
	}

	t.Run("recover", func(t *testing.T) {
		t.Parallel()

		for _, c := range recoverCases {
			t.Run(c.name, func(t *testing.T) {
				t.Parallel()

				p := testutil.NewParser(t, c.in+" 123")
				actual := testutil.AssertMatchesButError(t, p, CommaList("a", "as", elemFunc))
				if assert.Equal(t, c.expect, actual) {
					assert.Equal(t, p.Index(), len(c.in))
				}
			})
		}
	})
}
