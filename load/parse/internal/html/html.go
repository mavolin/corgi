// Package html implements a subset of rules from the HTML5 specification.
package html

import (
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/html/codepoint"
)

func TagName() parser.Func[string] { // https://html.spec.whatwg.org/multipage/syntax.html#syntax-tag-name
	return func(p *parser.Parser) string {
		s := parser.TokenWhile(p, func() bool {
			return codepoint.MatchesAny(p, codepoint.ASCIIAlphanumeric)
		})
		if s == "" {
			return ""
		}
		return s
	}
}

func AttributeName() parser.Func[string] { // https://html.spec.whatwg.org/multipage/syntax.html#syntax-attribute-name
	return func(p *parser.Parser) string {
		s := parser.TokenWhile(p, func() bool {
			return parser.Matches(p, AttributeNameRune())
		})
		if s == "" {
			return ""
		}
		return s
	}
}

func AttributeNameRune() parser.Func[rune] { // https://html.spec.whatwg.org/multipage/syntax.html#syntax-attribute-name
	return func(p *parser.Parser) rune {
		r := parser.TryRunePredicate(p, func(rune) bool {
			return !codepoint.MatchesAny(p, codepoint.Control, codepoint.Noncharacter) &&
				!parser.MatchesAnyRune(p, ' ', '"', '\'', '>', '/', '=')
		})
		if r == 0 {
			return 0
		}
		return r
	}
}
