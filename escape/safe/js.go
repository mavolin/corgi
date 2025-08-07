package safe

type (
	// JSLiteral encapsulates a known safe EcmaScript5 Literal, for example,
	// `5`, `"foo"`, or `{foo: "bar"}`.
	//
	// Using JSLiteral to include valid but untrusted JSON is not safe.
	// A safe alternative is to parse the JSON with json.Unmarshal and then
	// pass the resultant object into the template, where it will be
	// converted to sanitized JSON when presented in a JavaScript context.
	JSLiteral struct{ val string }

	// A Script encapsulates a known safe EcmaScript5 script.
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

// TrustedScript creates a new Script from the given trusted string.
//
// Only use this function if you have read the package documentation and are
// sure that the passed string satisfies the requirements for a Script.
func TrustedScript(s string) Script { return Script{val: s} }

func (js JSLiteral) Get() string { return js.val }
func (js Script) Get() string    { return js.val }
