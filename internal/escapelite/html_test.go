package escapelite

import (
	"testing"

	"github.com/mavolin/corgi/v2/internal/test/should"
)

func TestContent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want string
	}{
		{"foo & bar < baz", "foo &amp; bar &lt; baz"},
		{"<script>", "&lt;script>"},
		{"&", "&amp;"},
		{"<", "&lt;"},
		{"plain text", "plain text"},
		{"", ""},
	}

	for _, c := range tests {
		t.Run(c.in, func(t *testing.T) {
			t.Parallel()

			got := Content(c.in)
			should.Equal(t, got, c.want)
		})
	}
}

func TestDoubleQuoted(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want string
	}{
		{"foo & \"bar\"", "foo &amp; &#34;bar&#34;"},
		{"&", "&amp;"},
		{"\"quoted\"", "&#34;quoted&#34;"},
		{"plain text", "plain text"},
		{"", ""},
	}

	for _, c := range tests {
		t.Run(c.in, func(t *testing.T) {
			t.Parallel()

			got := DoubleQuoted(c.in)
			should.Equal(t, got, c.want)
		})
	}
}

func TestFullAttr(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want string
	}{
		{"foo & bar", `"foo &amp; bar"`},
		{"&", "&amp;"},
		{"plain text", `"plain text"`},
		{"word", "word"},
	}

	for _, c := range tests {
		t.Run(c.in, func(t *testing.T) {
			t.Parallel()

			got := FullAttr(c.in)
			should.Equal(t, got, c.want)
		})
	}
}
