package safe

type (
	// JSLiteral encapsulates a known safe EcmaScript5 Literal, for example,
	// `5`, `"foo"`, or `{foo: "bar"}`.
	//
	// A valid JSLiteral must not contain the sequence "</", to prevent the
	// premature end of the element containing it.
	// You should instead escape either character.
	// You must always adhere to this requirement, even if you only plan to use
	// the value in an attribute.
	// Corgi allows the use of a JSLiteral in custom elements marked as scripts,
	// therefore it is NOT enough to only escape the string "</script": you must
	// always escape "</" regardless of what follows!
	// It is also not enough to escape "</" only if it is eventually followed by
	// a ">": The HTML specification expressively forbids any occurrence of "</"
	// followed by the tag name, if it is followed by a rune from a character set
	// much bigger than just ">".
	//
	// Using JSLiteral to include valid but untrusted JSON is not safe.
	// A safe alternative is to parse the JSON with json.Unmarshal and then
	// pass the resultant object into the runtime, where it will be
	// converted to sanitized JSON when presented in a JavaScript context.
	//
	// See also:
	// https://html.spec.whatwg.org/multipage/scripting.html#restrictions-for-contents-of-script-elements
	JSLiteral struct{ val string }

	// A Script encapsulates a known safe EcmaScript5 script.
	//
	// A valid Script must not contain the sequence "</", to prevent the
	// premature end of the element containing it.
	// See [JSLiteral] for more information.
	//
	// Using Script to include valid but untrusted JSON is not safe.
	// See [JSLiteral] for more information.
	//
	// Concatenating Scripts is not safe, as the combined script might behave
	// differently than the separate scripts.
	Script struct{ val string }
)

// TrustedJSLiteral creates a new JSLiteral from the given trusted string.
//
// Only use this function if you have read the package documentation and are
// sure that the passed string satisfies the requirements for a JSLiteral.
func TrustedJSLiteral(s string) JSLiteral { return JSLiteral{val: s} }

func (j JSLiteral) Get() string { return j.val }
