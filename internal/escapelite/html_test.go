package escapelite

import (
	"testing"

	"github.com/mavolin/corgi/v2/internal/should"
)

func TestHTMLInt(t *testing.T) {
	t.Parallel()

	want := "-42"
	got := HTMLInt(-42)
	should.Equal(t, want, got)
}

func TestHTMLUint(t *testing.T) {
	t.Parallel()

	want := "42"
	got := HTMLUint(42)
	should.Equal(t, want, got)
}

func TestHTMLFloat(t *testing.T) {
	t.Parallel()

	want := "3.14"
	got := HTMLFloat(3.14)
	should.Equal(t, want, got)
}

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
