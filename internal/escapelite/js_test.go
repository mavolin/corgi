package escapelite

import (
	"testing"

	"github.com/mavolin/corgi/v2/internal/test/should"
)

func TestJSContent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty string", "", ""},
		{"plain text", "hello world", "hello world"},
		{"no escape needed", "var x = 42;", "var x = 42;"},
		{"single opening tag", "<script>", `\x3cscript>`},
		{"multiple opening tags", "<div><span>", `\x3cdiv>\x3cspan>`},
		{"closing tags", "</div></span>", `\x3c/div>\x3c/span>`},
		{"tag with whitespace", "< div>", "< div>"},
		{"lt sign followed by non-tag", "<42", `\x3c42`},
		{"script tag with content", "<script>alert('xss')</script>", `\x3cscript>alert('xss')\x3c/script>`},
		{"mixed content", "text <b>bold</b> text", `text \x3cb>bold\x3c/b> text`},
		{"comment-like", "<!-- comment -->", `\x3c!-- comment -->`},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got := JSContent(c.in)
			should.Equal(t, got, c.want)
		})
	}
}

func TestJS(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   any
		want string
	}{
		{"null", nil, "null"},
		{"empty string", "", `""`},
		{"string", "hello", `"hello"`},
		{"number", 42, "42"},
		{"float", 3.14, "3.14"},
		{"bool true", true, "true"},
		{"bool false", false, "false"},
		{"array", []string{"a", "b"}, `["a","b"]`},
		{"object", map[string]int{"a": 1, "b": 2}, `{"a":1,"b":2}`},
		{"escape quotes", `"quoted"`, `"\"quoted\""`},
		{"escape backslash", `a\b`, `"a\\b"`},
		{"escape control chars", "a\nb\tc", `"a\nb\tc"`},
		{"unicode escape", string([]rune{0x2028}), `"\u2028"`},
		{"unicode escape line separator", string([]rune{0x2029}), `"\u2029"`},
		{"mixed unicode escapes", "test" + string([]rune{0x2028, 0x2029}) + "end", `"test\u2028\u2029end"`},
		{"html special chars", `<script>alert("xss")</script>`, `"\u003cscript\u003ealert(\"xss\")\u003c/script\u003e"`},
		{"emoji", "😀", `"😀"`},
	}

	for _, c := range tests {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			got, err := JS(c.in)
			if should.NoError(t, err) {
				should.Equal(t, got, c.want)
			}
		})
	}
}
