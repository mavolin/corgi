package safe

type (
	// CSSValue encapsulates known safe content that matches any of the below
	// and is safe to be embedded in the body of a style attribute.
	//  1. The CSS3 stylesheet production, such as `p { color: purple }`.
	//  2. The CSS3 rule production, such as `a[href=~"https:"].foo#bar`.
	//  3. CSS3 declaration productions, such as `color: red; margin: 2px`.
	//  4. The CSS3 value production, such as `rgba(0, 0, 255, 127)`.
	//
	// A CSSValue must not contain the case-insensitive string "</style" to
	// prevent the premature end of the style element.
	//
	// See https://www.w3.org/TR/css3-syntax/#parsing and
	// https://web.archive.org/web/20090211114933/http://w3.org/TR/css3-syntax#style
	CSSValue struct{ val string }

	// CSSValueAttr encapsulates a known safe CSSValue that is safe to be
	// embedded in an attribute.
	//
	// It has the same requirements as CSSValue except that it may contain
	// "</style" without further escaping, but needs to escape double quotes.
	//
	// As it is used as an attribute, it may make use of ampersand escapes.
	//
	// Due to the difference in requirements between CSSValue and CSSValueAttr,
	// a CSSValueAttr cannot be interpolated in a style element.
	CSSValueAttr struct{ val string }
)

// TrustedCSSValue creates a new CSSValue from the given trusted string.
//
// Only use this function if you have read the package documentation and are
// sure that the passed string satisfies the requirements for a CSSValue.
func TrustedCSSValue(s string) CSSValue { return CSSValue{val: s} }

// TrustedCSSValueAttr creates a new CSSValueAttr from the given trusted string.
//
// Only use this function if you have read the package documentation and are
// sure that the passed string satisfies the requirements for a CSSValueAttr.
func TrustedCSSValueAttr(s string) CSSValueAttr { return CSSValueAttr{val: s} }

func (css CSSValue) Escaped() string   { return css.val }
func (a CSSValueAttr) Escaped() string { return a.val }
