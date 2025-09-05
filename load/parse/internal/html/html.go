// Package html implements a subset of rules from the HTML5 specification.
package html

import (
	"github.com/mavolin/corgi/v2/load/parse/internal/html/codepoint"
)

func TagNameStartRune(r rune) bool { // https://html.spec.whatwg.org/multipage/parsing.html#tag-open-state
	return codepoint.ASCIIAlpha(r)
}

func TagNameRemainingRune(r rune) bool { // https://html.spec.whatwg.org/multipage/parsing.html#tag-name-state
	switch r {
	case '\t', '\n', '\f', ' ':
		return false
	case '/', '>':
		return false
	default:
		return true
	}
}

func AttributeNameRune(r rune) bool { // https://html.spec.whatwg.org/multipage/syntax.html#syntax-attribute-name
	if codepoint.Control(r) || codepoint.Noncharacter(r) {
		return false
	}
	switch r {
	case ' ', '"', '\'', '>', '/', '=':
		return false
	}
	return true
}
