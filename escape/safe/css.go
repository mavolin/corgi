package safe

type (
	// CSSValue encapsulates known safe CSS3 value that can safely be embedded
	// in a style element.
	//
	// See https://www.w3.org/TR/css3-syntax/#parsing and
	// https://web.archive.org/web/20090211114933/http://w3.org/TR/css3-syntax#style
	CSSValue struct{ val string }

	// CSSDeclarations encapsulates a known safe terminated CSS3 declaration
	// block.
	// Unless empty, it must be terminated by a semicolon.
	CSSDeclarations struct{ val string }

	// A CSSRuleSet encapsulates a known safe CSS3 ruleset or other top-level,
	// positionally independent CSS constructs, e.g. media queries or
	// animations.
	//
	// Unlike [Stylesheet], multiple [CSSRuleSet] values must be safe to
	// concatenate.
	CSSRuleSet struct{ val string }

	// A Stylesheet is a known safe CSS3 stylesheet.
	//
	// Stylesheets cannot be concatenated, as they might contain positional
	// at-rules, which are usually only allowed at the top of a stylesheet.
	Stylesheet struct{ val string }
)

// TrustedCSSValue creates a new CSSValue from the given trusted string.
//
// Only use this function if you have read the package documentation and are
// sure that the passed string satisfies the requirements for a CSSValue.
func TrustedCSSValue(s string) CSSValue { return CSSValue{val: s} }

// TrustedCSSDeclarations creates a new CSSDeclarations from the given trusted
// string.
//
// Only use this function if you have read the package documentation and are
// sure that the passed string satisfies the requirements for CSSDeclarations.
func TrustedCSSDeclarations(s string) CSSDeclarations { return CSSDeclarations{val: s} }

// TrustedCSSRuleSet creates a new CSSRuleSet from the given trusted string.
//
// Only use this function if you have read the package documentation and are
// sure that the passed string satisfies the requirements for a CSSRuleSet.
func TrustedCSSRuleSet(s string) CSSRuleSet { return CSSRuleSet{val: s} }

// TrustedStylesheet creates a new Stylesheet from the given trusted string.
//
// Only use this function if you have read the package documentation and are
// sure that the passed string satisfies the requirements for a Stylesheet.
func TrustedStylesheet(s string) Stylesheet { return Stylesheet{val: s} }

func (css CSSValue) Get() string        { return css.val }
func (css CSSDeclarations) Get() string { return css.val }
func (css CSSRuleSet) Get() string      { return css.val }
func (css Stylesheet) Get() string      { return css.val }
