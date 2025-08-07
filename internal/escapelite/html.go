package escapelite

import "strings"

var contentEscaper = strings.NewReplacer(
	"&", "&amp;",
	"<", "&lt;",
)

// Content replaces [&<] with escape sequences.
func Content(s unescaped) string {
	return contentEscaper.Replace(s)
}

var quotedEscaper = strings.NewReplacer(
	"&", "&amp;",
	`"`, "&#34;",
)

// DoubleQuoted escapes s by replacing [&"] so that it can safely be placed
// between double quotes as an HTML attribute value.
//
// The returned content is not suitable for use in single or unquoted
// attributes.
func DoubleQuoted(s unescaped) string {
	return quotedEscaper.Replace(s)
}

var unquotedEscaper = strings.NewReplacer("&", "&amp;")

// FullAttr escapes s to be placed behind the equal sign of an HTML attribute.
//
// If possible, it will not quote the value to save on the otherwise two
// additional needed quotes.
// If it must, it will return a double-quoted value.
//
// This analysis comes at the cost of performance, needing two instead of a
// single pass over the string, first to determine whether quotes are needed,
// and then to escape accordingly.
func FullAttr(s unescaped) string {
	if needQuotes(s) {
		return `"` + DoubleQuoted(s) + `"`
	}
	return unquotedEscaper.Replace(s)
}

// needQuotes checks whether the given string would need to be quoted if it were
// to be used as an HTML attribute value.
func needQuotes(s unescaped) bool {
	if s == "" {
		return true
	}

	// https://html.spec.whatwg.org/multipage/syntax.html#attributes-2
	for _, r := range s {
		switch r {
		case '\t', '\n', '\f', '\r', ' ': // ASCII whitespace
			fallthrough
		case '"', '\'', '=', '<', '>', '`':
			return true
		}
	}

	return false
}
