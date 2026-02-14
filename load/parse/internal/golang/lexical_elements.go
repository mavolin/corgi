package golang

import (
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/whitespace"
)

// https://go.dev/ref/spec#Lexical_elements

// ============================================================================
// Rune literals
// ======================================================================================

func RuneLit() parser.Func[bool] {
	return func(p *parser.Parser) bool {
		if !parser.TryRune(p, '\'') {
			return false
		}

		if !parser.TryInOrder(p, ByteValue(), UnicodeValue('\'')) {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "rune literal: missing rune",
				Primary: quickanno.Expected(p, p.Pos(), "a rune"),
			})
		}

		startPos := p.Pos()
		unexpected := parser.TokenWhile(p, func() bool {
			return !parser.MatchesRune(p, '\'') &&
				(parser.MatchesWS(p, whitespace.Horizontal()) || !parser.Matches(p, comment.AndEOS()))
		})
		if !parser.TryRune(p, '\'') {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "rune literal: missing closing quote",
				Primary: quickanno.Expected(p, p.Pos(), "a closing quote for the opening quote here"),
			})
		} else if unexpected != "" {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "unexpected runes in rune literal",
				Primary: []diagnostic.Annotation{
					anno.Range(p.File, startPos, p.Pos(), "too many runes"),
				},
			})
		}

		return true
	}
}

func UnicodeValue(term rune) parser.Func[bool] {
	return func(p *parser.Parser) bool {
		if parser.MatchesRune(p, '\\') {
			return parser.TryInOrder(p, LittleUValue(), BigUValue(), EscapedChar(term))
		}

		return parser.Try(p, UnicodeChar(term))
	}
}

func ByteValue() parser.Func[bool] {
	return func(p *parser.Parser) bool {
		return parser.TryInOrder(p, OctalByteValue(), HexByteValue())
	}
}

func OctalByteValue() parser.Func[bool] {
	return func(p *parser.Parser) bool {
		if !parser.TryRune(p, '\\') {
			return false
		}

		for range 3 {
			if parser.TryRunePredicate(p, Octal_Digit) == 0 {
				return false
			}
		}
		return true
	}
}

func HexByteValue() parser.Func[bool] {
	return func(p *parser.Parser) bool {
		if !parser.TryToken(p, `\x`) {
			return false
		}

		for range 2 {
			if parser.TryRunePredicate(p, Hex_Digit) == 0 {
				return false
			}
		}
		return true
	}
}

func UnicodeChar(except rune) parser.Func[bool] {
	return func(p *parser.Parser) bool {
		return parser.TryRunePredicate(p, func(r rune) bool {
			return r != except && Unicode_Char(r)
		}) != 0
	}
}

func LittleUValue() parser.Func[bool] {
	return func(p *parser.Parser) bool {
		if !parser.TryToken(p, `\u`) {
			return false
		}

		for range 4 {
			if parser.TryRunePredicate(p, Hex_Digit) == 0 {
				return false
			}
		}
		return true
	}
}

func BigUValue() parser.Func[bool] {
	return func(p *parser.Parser) bool {
		if !parser.TryToken(p, `\U`) {
			return false
		}

		for range 8 {
			if parser.TryRunePredicate(p, Hex_Digit) == 0 {
				return false
			}
		}
		return true
	}
}

func EscapedChar(term rune) parser.Func[bool] {
	return func(p *parser.Parser) bool {
		if !parser.TryRune(p, '\\') {
			return false
		}

		r := parser.TryAnyRune(p, 'a', 'b', 'f', 'n', 'r', 't', 'v', '\\', term)
		if r == 0 {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "invalid escaped character",
				Primary: quickanno.Expected(p, p.Pos(), "a valid escaped character"),
			})
		}
		return true
	}
}
