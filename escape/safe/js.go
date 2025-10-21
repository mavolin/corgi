package safe

import (
	"fmt"
	"regexp"

	_ "unsafe" // for go:linkname

	"github.com/mavolin/corgi/v2/internal/escapelite"
)

type (
	// JSLiteral encapsulates a known safe EcmaScript5 Literal, for example,
	// `5`, `"foo"`, or `{foo: "bar"}`.
	//
	// Using JSLiteral to include valid but untrusted JSON is not safe.
	// A safe alternative is to parse the JSON with json.Unmarshal and then
	// pass the resultant object into the template, where it will be
	// converted to sanitized JSON when presented in a JavaScript context.
	// More specifically, not all JSON is valid JavaScript.
	//
	// JSLiterals must not contain the case-insensitive sequences "<script",
	// "</script", and "<!--".
	// For simplicity, you might choose to always escape "<", to trivially
	// satisfy this requirement.
	JSLiteral struct{ val string }

	// A Script encapsulates a known safe EcmaScript5 script.
	//
	// Using Script to include valid but untrusted JSON is not safe.
	// See [JSLiteral] for more information.
	//
	// Concatenating Scripts is not safe, as the combined script might behave
	// differently than the separate scripts.
	//
	// Scripts must not contain the case-insensitive sequences "<script",
	// "</script", and "<!--".
	// For simplicity, you might choose to always escape "<", to trivially
	// satisfy this requirement.
	Script struct{ val string }
)

// JSLiteralFromData creates a new JSLiteral by encoding the given data as
// JSON.
// The usual json.Marshal rules apply.
//
// Before using this function, read the documentation of [JSLiteral].
func JSLiteralFromData(data any) (JSLiteral, error) {
	enc, err := escapelite.JSLiteral(data)
	if err != nil {
		return JSLiteral{}, err
	}
	return trustedJSLiteral(enc), nil
}

// ConstantScript creates a new Script wrapper from the given string
// constant.
//
// Before using this function, read the documentation of [Script].
func ConstantScript(c constant) Script {
	return trustedScript(string(c))
}

// jsIdentifierPattern matches strings that are valid Javascript identifiers.
//
// This pattern accepts only a subset of valid identifiers defined in
// https://tc39.github.io/ecma262/#sec-names-and-keywords. In particular,
// it does not match identifiers that contain non-ASCII letters, Unicode
// escape sequences, and the Unicode format-control characters
// \u200C (zero-width non-joiner) and \u200D (zero-width joiner).
var jsIdentifierPattern = regexp.MustCompile(`^[$_a-zA-Z][$_a-zA-Z0-9]+$`)

// ConstantScriptWithData formats the script as
//
//	var #{name} = #{data}; #{script}
//
// where #{name} is the given constant name, #{data} is the JSON encoding of
// the given data, and #{script} is the given script constant.
//
// Name must conform to the subset of valid JavaScript identifiers matched by
// the regular expression
//
//	[$_a-zA-Z][$_a-zA-Z0-9]+
//
// Before using this function, read the documentation of [Script].
func ConstantScriptWithData(name constant, data any, script constant) (Script, error) {
	if !jsIdentifierPattern.MatchString(string(name)) {
		return Script{}, fmt.Errorf("name %q is not a valid JavaScript identifier", string(name))
	}

	enc, err := escapelite.JSLiteral(data)
	if err != nil {
		return Script{}, err
	}
	s := "var " + string(name) + "=" + enc + ";" + string(script)
	return trustedScript(s), nil
}

//go:linkname trustedJSLiteral
func trustedJSLiteral(s string) JSLiteral {
	return JSLiteral{val: s}
}

//go:linkname trustedScript
func trustedScript(s string) Script {
	return Script{val: s}
}

func (js JSLiteral) Get() string { return js.val }
func (js Script) Get() string    { return js.val }
