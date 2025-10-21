package escapelite

import (
	"image/color"
	"testing"

	"github.com/mavolin/corgi/v2/internal/should"
)

func TestCSSInt(t *testing.T) {
	t.Parallel()

	want := "42"
	got := CSSInt(42)
	should.Equal(t, want, got)
}

func TestCSSFloat(t *testing.T) {
	t.Parallel()

	want := "3.14"
	got := CSSFloat(3.14)
	should.Equal(t, want, got)
}

func TestCSSColor(t *testing.T) {
	t.Parallel()

	t.Run("opaque", func(t *testing.T) {
		t.Parallel()

		c := color.NRGBA{R: 0x0c, G: 0x22, B: 0x38, A: 255}
		got := CSSColor(c)
		should.Equal(t, got, "#0c2238")
	})

	t.Run("alpha", func(t *testing.T) {
		t.Parallel()

		c := color.NRGBA{R: 0x0c, G: 0x22, B: 0x38, A: 0x7f}
		got := CSSColor(c)
		should.Equal(t, got, "#0c22387f")
	})
}

func TestCSSString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want string
	}{
		{`hello world`, `"hello world"`},
		{`he said "hi"`, `"he said \"hi\""`},
		{"line1\nline2", `"line1\00000Aline2"`},
		{`back\slash`, `"back\\slash"`},
		{`control` + string(rune(0x01)) + `char`, `"control\000001char"`},
	}

	for _, c := range tests {
		t.Run(c.in, func(t *testing.T) {
			t.Parallel()

			got := CSSString(c.in)
			should.Equal(t, got, c.want)
		})
	}
}
