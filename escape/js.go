package escape

import (
	"encoding/json"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/mavolin/corgi/escape/safe"
)

// JSify converts the passed value to a JavaScript value.
//
// It is safe to embed into HTML without further escaping.
func JSify(val any) (safe.JS, error) {
	switch t := val.(type) {
	case safe.JS:
		return t, nil
	case json.Marshaler:
		// Do not treat as a Stringer.
	case fmt.Stringer:
		val = t.String()
	}

	jsonVal, err := json.Marshal(val)
	if err != nil {
		return safe.JS{}, err
	}

	if len(jsonVal) == 0 {
		// In, `x=y/{{.}}*z` a json.Marshaler that produces "" should
		// not cause the output `x=y/*z`.
		return safe.TrustedJS(" null "), nil
	}
	first, _ := utf8.DecodeRune(jsonVal)
	last, _ := utf8.DecodeLastRune(jsonVal)
	var buf strings.Builder
	// Prevent IdentifierNames and NumericLiterals from running into
	// keywords: in, instanceof, typeof, void
	pad := isJSIdentPart(first) || isJSIdentPart(last)
	if pad {
		buf.WriteByte(' ')
	}
	written := 0
	// Make sure that json.Marshal escapes codepoints U+2028 & U+2029
	// so it falls within the subset of JSON which is valid JS.
	for i := 0; i < len(jsonVal); {
		r, n := utf8.DecodeRune(jsonVal[i:])
		repl := ""
		if r == 0x2028 {
			repl = `\u2028`
		} else if r == 0x2029 {
			repl = `\u2029`
		}
		if repl != "" {
			buf.Write(jsonVal[written:i])
			buf.WriteString(repl)
			written = i + n
		}
		i += n
	}
	if buf.Len() != 0 {
		buf.Write(jsonVal[written:])
		if pad {
			buf.WriteByte(' ')
		}
		return safe.TrustedJS(buf.String()), nil
	}
	return safe.TrustedJS(string(jsonVal)), nil
}

func JSAttrify(val any) (safe.JSAttr, error) {
	attrVal, ok := val.(safe.JSAttr)
	if ok {
		return attrVal, nil
	}

	s, err := JSify(val)
	if err != nil {
		return safe.JSAttr{}, err
	}

	return safe.TrustedJSAttr(plainAttrEscaper.Replace(s.Escaped())), nil
}

// isJSIdentPart reports whether the given rune is a JS identifier part.
// It does not handle all the non-Latin letters, joiners, and combining marks,
// but it does handle every codepoint that can occur in a numeric literal or
// a keyword.
func isJSIdentPart(r rune) bool {
	switch {
	case r == '$':
		return true
	case '0' <= r && r <= '9':
		return true
	case 'A' <= r && r <= 'Z':
		return true
	case r == '_':
		return true
	case 'a' <= r && r <= 'z':
		return true
	}
	return false
}
