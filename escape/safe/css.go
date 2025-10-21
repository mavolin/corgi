package safe

import (
	"fmt"
	"image/color"
	"strings"

	_ "unsafe" // for go:linkname

	"github.com/mavolin/corgi/v2/internal/escapelite"
)

type (
	// CSSValue represents a known safe CSS3 value that can be used on the
	// right-hand side of a CSS property.
	//
	// It never contains a literal '<'.
	CSSValue struct{ val string }

	// CSSDeclarations encapsulates a known safe terminated CSS3 declaration
	// block.
	// Unless empty, it must be terminated by a semicolon.
	//
	// It never contains a literal '<'.
	CSSDeclarations struct{ val string }

	// A Stylesheet is a known safe CSS3 stylesheet.
	//
	// Stylesheets cannot be concatenated, as they might contain positional
	// at-rules, which are usually only allowed at the top of a stylesheet.
	//
	// A valid Stylesheet never contains a literal '<'.
	Stylesheet struct{ val string }
)

// ConstantCSSValue creates a new CSSValue wrapper from the given string
// constant.
//
// Before using this function, read the documentation of [CSSValue].
func ConstantCSSValue(c constant) CSSValue {
	s := string(c)
	if strings.ContainsRune(s, '<') {
		panic("CSSValue must not contain literal '<'")
	}
	return trustedCSSValue(s)
}

// CSSInt returns a CSSValue representing the given integer number.
func CSSInt(n int) CSSValue {
	return trustedCSSValue(escapelite.CSSInt(int64(n)))
}

// CSSFloat returns a CSSValue representing the given floating point number.
func CSSFloat(f float64) CSSValue {
	return trustedCSSValue(escapelite.CSSFloat(f))
}

// CSSColor returns a CSSValue representing the given color in CSS syntax.
//
// Before using this function, read the documentation of [CSSValue].
func CSSColor(c color.Color) CSSValue {
	return trustedCSSValue(escapelite.CSSColor(c))
}

// ConstantCSSDeclarations creates a new CSSDeclarations wrapper from the given
// string constant.
//
// Before using this function, read the documentation of [CSSDeclarations].
func ConstantCSSDeclarations(c constant) CSSDeclarations {
	s := string(c)
	if s != "" && s[len(s)-1] != ';' {
		panic("CSSDeclarations must be terminated by a semicolon")
	} else if strings.ContainsRune(s, '<') {
		panic("CSSDeclarations must not contain literal '<'")
	}
	return trustedCSSDeclarations(s)
}

// FormatCSSDeclaration formats the CSS property with the given constant name
// and the given CSSValue into a CSSDeclarations.
func FormatCSSDeclaration(name constant, value CSSValue) CSSDeclarations {
	return trustedCSSDeclarations(fmt.Sprintf("%s: %s;", string(name), value.Get()))
}

// ConcatCSSDeclarations concatenates multiple CSSDeclarations into one.
func ConcatCSSDeclarations(decls ...CSSDeclarations) CSSDeclarations {
	var n int
	for _, d := range decls {
		n += len(d.val)
	}
	var sb strings.Builder
	sb.Grow(n)
	for _, d := range decls {
		sb.WriteString(d.val)
	}
	return trustedCSSDeclarations(sb.String())
}

// ConstantStylesheet creates a new Stylesheet wrapper from the given string
// constant.
//
// Before using this function, read the documentation of [Stylesheet].
func ConstantStylesheet(c constant) Stylesheet {
	if strings.ContainsRune(string(c), '<') {
		panic("Stylesheet must not contain literal '<'")
	}
	return trustedStylesheet(string(c))
}

//go:linkname trustedCSSValue
func trustedCSSValue(s string) CSSValue { return CSSValue{val: s} }

//go:linkname trustedCSSDeclarations
func trustedCSSDeclarations(s string) CSSDeclarations { return CSSDeclarations{val: s} }

//go:linkname trustedStylesheet
func trustedStylesheet(s string) Stylesheet { return Stylesheet{val: s} }

func (cv CSSValue) Get() string         { return cv.val }
func (css CSSDeclarations) Get() string { return css.val }
func (css Stylesheet) Get() string      { return css.val }
