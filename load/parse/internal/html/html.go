// Package html implements a subset of rules from the HTML5 specification.
package html

import (
	"github.com/mavolin/corgi/v2/fancyerr"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/html/codepoint"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

func TagName() parser.Func[string] { // https://html.spec.whatwg.org/multipage/syntax.html#syntax-tag-name
	return func(p *parser.Parser) (string, *fancyerr.Error) {
		s := parser.TokenWhile(p, func() bool {
			return codepoint.MatchesAny(p, codepoint.ASCIIAlphanumeric)
		})
		if s == "" {
			return "", &fancyerr.Error{
				Message: "missing html tag name",
				Primary: quickanno.Expected(p, p.Pos(), "an html tag name"),
			}
		}
		return s, nil
	}
}

func AttributeName() parser.Func[string] { // https://html.spec.whatwg.org/multipage/syntax.html#syntax-attribute-name
	return func(p *parser.Parser) (string, *fancyerr.Error) {
		s := parser.TokenWhile(p, func() bool {
			return parser.Matches(p, AttributeNameRune())
		})
		if s == "" {
			return "", &fancyerr.Error{
				Message: "missing attribute name",
				Primary: quickanno.Expected(p, p.Pos(), "an attribute name"),
			}
		}
		return s, nil
	}
}

func AttributeNameRune() parser.Func[rune] { // https://html.spec.whatwg.org/multipage/syntax.html#syntax-attribute-name
	return func(p *parser.Parser) (rune, *fancyerr.Error) {
		r := parser.TryRunePredicate(p, func(r rune) bool {
			return !codepoint.MatchesAny(p, codepoint.Control, codepoint.Noncharacter) &&
				!parser.MatchesAnyRune(p, ' ', '"', '\'', '>', '/', '=')
		})
		if r < 0 {
			return 0, &fancyerr.Error{
				Message: "missing attribute name",
				Primary: quickanno.Expected(p, p.Pos(), "an attribute name"),
			}
		}
		return r, nil
	}
}
