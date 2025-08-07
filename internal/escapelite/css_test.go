// filepath: /home/mavolin/Code/github.com/mavolin/corgi/internal/escapelite/css test.go
package escapelite

import (
	"testing"

	"github.com/mavolin/corgi/v2/internal/test/should"
)

func TestCSSValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{"simple text", "woof bark", "woof bark"},
		{"empty string", "", ""},
		{"special chars", `& < > " '`, `\26  \3c  \3e  \22  \27 `},
		{"semicolon", "a;b", `a\3b b`},
		{"backlash", `a\b`, `a\\b`},
		{"braces", "a{b}c", `a\7b b\7d c`},
		{"colon", "a:b", `a\3a b`},
		{"forward slash", "a/b", `a\2f b`},
		{"plus", "a+b", `a\2b b`},
		{"parentheses", "a(b)c", `a\28 b\29 c`},
		{"color", "#abc", `#abc`},
		{"non-ascii", "café", "café"},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got := CSSValue(c.in)
			should.Equal(t, got, c.want)
		})
	}
}

func TestFilterCSSValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{"simple identifier", "simple", "simple"},
		{"empty string", "", ""},
		{"color keyword", "blue", "blue"},
		{"pixel dimension", "10px", "10px"},
		{"percentage value", "25%", "25%"},
		{"hex color", "#abc", "#abc"},
		{"rgba with parentheses", "rgba(0,0,0,0.5)", Replacement}, // Contains parentheses, should be replaced
		{"javascript url", `url("javascript:alert(1)")`, Replacement},
		{"css expression", "expression(alert(1))", Replacement},
		{"moz binding", "mozbinding:url(alert)", Replacement},
		{"prefixed moz binding", "-moz-binding:url(alert)", Replacement},
		{"script injection", "</style><script>alert(1)</script>", Replacement},
		{"html comment", "<!-- comment -->", Replacement},
		{"hex escape sequence", `\61\62\63`, "abc"},      // \61 = 'a', \62 = 'b', \63 = 'c'
		{"escaped quotes", `\22 quoted\22`, Replacement}, // Contains quotes, should be replaced
		{"unmatched brace", `{unmatched`, Replacement},
		{"unmatched quote", `"unmatched`, Replacement},
		{"unmatched parenthesis", `(unmatched`, Replacement},
		{"unmatched bracket", `[unmatched`, Replacement},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got := FilterCSSValue(c.in)
			should.Equal(t, got, c.want)
		})
	}
}
