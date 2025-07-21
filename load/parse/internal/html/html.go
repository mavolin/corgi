// Package html implements a subset of rules from the HTML5 specification.
package html

import (
	"github.com/mavolin/corgi/v2/file/diagnostic"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/html/codepoint"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

func TagName() parser.Func[string] { // https://html.spec.whatwg.org/multipage/syntax.html#syntax-tag-name
	return func(p *parser.Parser) (string, *diagnostic.Diagnostic) {
		s := parser.TokenWhile(p, func() bool {
			return codepoint.MatchesAny(p, codepoint.ASCIIAlphanumeric)
		})
		if s == "" {
			return "", &diagnostic.Diagnostic{
				Message: "missing html tag name",
				Primary: quickanno.Expected(p, p.Pos(), "an html tag name"),
			}
		}
		return s, nil
	}
}

func AttributeName() parser.Func[string] { // https://html.spec.whatwg.org/multipage/syntax.html#syntax-attribute-name
	return func(p *parser.Parser) (string, *diagnostic.Diagnostic) {
		s := parser.TokenWhile(p, func() bool {
			return parser.Matches(p, AttributeNameRune())
		})
		if s == "" {
			return "", &diagnostic.Diagnostic{
				Message: "missing attribute name",
				Primary: quickanno.Expected(p, p.Pos(), "an attribute name"),
			}
		}
		return s, nil
	}
}

func AttributeNameRune() parser.Func[rune] { // https://html.spec.whatwg.org/multipage/syntax.html#syntax-attribute-name
	return func(p *parser.Parser) (rune, *diagnostic.Diagnostic) {
		r := parser.TryRunePredicate(p, func(rune) bool {
			return !codepoint.MatchesAny(p, codepoint.Control, codepoint.Noncharacter) &&
				!parser.MatchesAnyRune(p, ' ', '"', '\'', '>', '/', '=')
		})
		if r < 0 {
			return 0, &diagnostic.Diagnostic{
				Message: "missing attribute name",
				Primary: quickanno.Expected(p, p.Pos(), "an attribute name"),
			}
		}
		return r, nil
	}
}
