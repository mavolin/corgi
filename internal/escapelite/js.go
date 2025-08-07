package escapelite

import (
	"encoding/json"
	"regexp"
	"strings"
	"unicode/utf8"
)

var jsContentEscapeMatcher = regexp.MustCompile("<([^ \t\r\n])")

// JSContent escapes the passed string for use in a JS element, more
// specifically so that it passes the "Restrictions for contents of script
// elements".
//
// This function is not needed for the contents of js attributes.
//
// See:
// https://html.spec.whatwg.org/multipage/scripting.html#restrictions-for-contents-of-script-elements
func JSContent(s unescaped) string {
	return jsContentEscapeMatcher.ReplaceAllString(s, `\x3c$1`)
}

func JS(val any) (string, error) {
	b, err := json.Marshal(val)
	if err != nil {
		return "", err
	}

	if len(b) == 0 {
		// In, `x=y/{{.}}*z` a json.Marshaler that produces "" should
		// not cause the output `x=y/*z`.
		return "null", nil
	}

	var buf strings.Builder
	written := 0
	// Make sure that json.Marshal escapes codepoints U+2028 & U+2029
	// so it falls within the subset of JSON which is valid JS.
	for i := 0; i < len(b); {
		r, n := utf8.DecodeRune(b[i:])
		repl := ""
		switch r {
		case 0x2028:
			repl = `\u2028`
		case 0x2029:
			repl = `\u2029`
		}
		if repl != "" {
			buf.Write(b[written:i])
			buf.WriteString(repl)
			written = i + n
		}
		i += n
	}
	if buf.Len() != 0 {
		buf.Write(b[written:])
		return buf.String(), nil
	}
	return string(b), nil
}
