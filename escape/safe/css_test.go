package safe

import (
	"image/color"
	"testing"

	"github.com/mavolin/corgi/v2/internal/test/should"
)

func TestConstantCSSValue(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		want := constant("12rem")
		got := ConstantCSSValue(want)
		should.Equal(t, got.Get(), string(want))
	})

	t.Run("failure", func(t *testing.T) {
		t.Parallel()

		t.Run("contains <", func(t *testing.T) {
			t.Parallel()
			should.Panic(t, func() {
				ConstantCSSValue(`"2 < 3"`)
			}, "CSSValue must not contain literal '<'")
		})
	})
}

func TestCSSInt(t *testing.T) {
	t.Parallel()

	want := "42"
	got := CSSInt(42).Get()
	should.Equal(t, want, got)
}

func TestCSSFloat(t *testing.T) {
	t.Parallel()

	want := "3.14"
	got := CSSFloat(3.14).Get()
	should.Equal(t, want, got)
}

func TestCSSColor(t *testing.T) {
	t.Parallel()

	t.Run("opaque", func(t *testing.T) {
		t.Parallel()

		c := color.NRGBA{R: 0x0c, G: 0x22, B: 0x38, A: 255}
		got := CSSColor(c).Get()
		should.Equal(t, got, "#0c2238")
	})

	t.Run("alpha", func(t *testing.T) {
		t.Parallel()

		c := color.NRGBA{R: 0x0c, G: 0x22, B: 0x38, A: 0x7f}
		got := CSSColor(c).Get()
		should.Equal(t, got, "#0c22387f")
	})
}

func TestConstantCSSDeclarations(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		want := constant("color: red; font-size: 12rem;")
		got := ConstantCSSDeclarations(want).Get()
		should.Equal(t, got, string(want))
	})

	t.Run("failure", func(t *testing.T) {
		t.Parallel()

		t.Run("missing ;", func(t *testing.T) {
			t.Parallel()
			should.Panic(t, func() {
				ConstantCSSDeclarations("color: red")
			}, "CSSDeclarations must be terminated by a semicolon")
		})

		t.Run("contains <", func(t *testing.T) {
			t.Parallel()
			should.Panic(t, func() {
				ConstantCSSDeclarations("content: '2 < 3';")
			}, "CSSDeclarations must not contain literal '<'")
		})
	})
}

func TestFormatCSSDeclaration(t *testing.T) {
	t.Parallel()

	val := ConstantCSSValue("red")

	got := FormatCSSDeclaration("color", val).Get()
	should.Equal(t, got, "color: red;")
}

func TestConcatCSSDeclarations(t *testing.T) {
	t.Parallel()

	dec1 := ConstantCSSDeclarations("color: red;")
	dec2 := ConstantCSSDeclarations("font-size: 12rem;")

	got := ConcatCSSDeclarations(dec1, dec2).Get()
	should.Equal(t, got, "color: red;font-size: 12rem;")
}

func TestConstantStylesheet(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		want := constant("body { color: red; }")
		got := ConstantStylesheet(want).Get()
		should.Equal(t, got, string(want))
	})

	t.Run("failure", func(t *testing.T) {
		t.Parallel()

		t.Run("contains <", func(t *testing.T) {
			t.Parallel()
			should.Panic(t, func() {
				ConstantStylesheet("body { content: '<'; }")
			}, "Stylesheet must not contain literal '<'")
		})
	})
}
