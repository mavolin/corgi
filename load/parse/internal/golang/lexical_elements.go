package golang

import (
	"github.com/mavolin/corgi/v2/file/ast"
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
			return !parser.MatchesAnyRune(p, '\'') &&
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
		if parser.MatchesAnyRune(p, '\\') {
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

// ============================================================================
// String literals
// ======================================================================================

func StringLit() parser.Func[*ast.StaticString] {
	return func(p *parser.Parser) *ast.StaticString {
		s := parser.TryInOrder(p, RawStringLit(), InterpretedStringLit())
		if s == nil {
			return nil
		}
		return s
	}
}

func RawStringLit() parser.Func[*ast.StaticString] {
	return func(p *parser.Parser) *ast.StaticString {
		open := parser.TryRuneAt(p, '`')
		if open == nil {
			return nil
		}

		var s ast.StaticString
		s.Quote = '`'
		s.Open = open

		s.Contents = parser.TokenWhile(p, func() bool {
			return !parser.MatchesAnyRune(p, '`')
		})

		s.Close = parser.TryRuneAt(p, '`')
		if s.Close == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "raw string literal: missing closing backtick",
				Primary: quickanno.Expected(p, *s.Open, "a closing backtick for the opening backtick here"),
			})
		}
		return &s
	}
}

func InterpretedStringLit() parser.Func[*ast.StaticString] {
	return func(p *parser.Parser) *ast.StaticString {
		open := parser.TryRuneAt(p, '"')
		if open == nil {
			return nil
		}

		var s ast.StaticString
		s.Quote = '"'
		s.Open = open

		index := p.Index()
		for parser.TryInOrder(p, ByteValue(), UnicodeValue('"')) {
		}
		s.Contents = p.AST.Raw[index:p.Index()]

		s.Close = parser.TryRuneAt(p, '"')
		if s.Close == nil {
			s.Close = nil
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "interpreted string literal: missing closing quote",
				Primary: quickanno.Expected(p, *s.Open, "a closing quote for the opening quote here"),
			})
		}

		return &s
	}
}
