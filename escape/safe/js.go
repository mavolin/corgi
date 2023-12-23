package safe

type (
	// JS encapsulates a known safe EcmaScript5 Expression, for example,
	// `(x + y * z())`.
	// Template authors are responsible for ensuring that typed expressions
	// do not break the intended precedence and that there is no
	// statement/expression ambiguity as when passing an expression like
	// "{ foo: bar() }\n['foo']()", which is both a valid Expression and a
	// valid Program with a very different meaning.
	//
	// A JS value must not contain the case-insensitive string "</script" to
	// prevent the premature end of the script element.
	//
	// Using JS to include valid but untrusted JSON is not safe.
	// A safe alternative is to parse the JSON with json.Unmarshal and then
	// pass the resultant object into the template, where it will be
	// converted to sanitized JSON when presented in a JavaScript context.
	JS struct{ val string }

	// JSAttr is a js attribute safe to be embedded in double quotes and used
	// as an attribute value.
	//
	// It differs from JS only in that it may contain the case-insensitive
	// string "</script" without further escaping and that it must escape
	// double quotes.
	//
	// Due to the difference in requirements between JS and JSAttr, a JSAttr
	// value cannot be interpolated in a style element.
	//
	// Using JS to include valid but untrusted JSON is not safe.
	// A safe alternative is to parse the JSON with json.Unmarshal and then
	// pass the resultant object into the template, where it will be
	// converted to sanitized JSON when presented in a JavaScript context.
	JSAttr struct{ val string }
)

// TrustedJS creates a new JS from the given trusted string.
//
// Only use this function if you have read the package documentation and are
// sure that the passed string satisfies the requirements for a JS.
func TrustedJS(s string) JS { return JS{val: s} }

// TrustedJSAttr creates a new JSAttr from the given trusted string.
//
// Only use this function if you have read the package documentation and are
// sure that the passed string satisfies the requirements for a JSAttr.
func TrustedJSAttr(s string) JSAttr { return JSAttr{val: s} }

func (j JS) Escaped() string     { return j.val }
func (a JSAttr) Escaped() string { return a.val }
