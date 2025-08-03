package golang

import (
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/unexpected"
	"github.com/mavolin/corgi/v2/load/parse/internal/whitespace"
)

// https://go.dev/ref/spec#Lexical_elements

// ============================================================================
// Rune literals
// ======================================================================================

func RuneLit() parser.Func[string] {
	return func(p *parser.Parser) string {
		if !parser.TryRune(p, '\'') {
			return ""
		}

		v := parser.TryInOrder(p, ByteValue(), UnicodeValue('\''))
		if v == "" {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "rune literal: missing rune",
				Primary: quickanno.Expected(p, p.Pos(), "a rune"),
			})
		}

		state := p.CloneState()
		err := unexpected.UntilAnyRune(p, whitespace.Horizontal(), '\'')
		if err != nil {
			err.Message = "unexpected runes after rune literal"
			p.CaptureError(err)
		}

		if !parser.TryRune(p, '\'') {
			p.RestoreState(state)
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "rune literal: missing closing quote",
				Primary: quickanno.Expected(p, p.Pos(), "a closing quote for the opening quote here"),
			})
		}

		return "'" + v + "'"
	}
}

func UnicodeValue(term rune) parser.Func[string] {
	return func(p *parser.Parser) string {
		if parser.MatchesAnyRune(p, '\\') {
			return parser.TryInOrder(p, LittleUValue(), BigUValue(), EscapedChar(term))
		}

		r := parser.Try(p, UnicodeChar(term))
		if r != 0 {
			return string(r)
		}
		return ""
	}
}

func ByteValue() parser.Func[string] {
	return func(p *parser.Parser) string {
		return parser.TryInOrder(p, OctalByteValue(), HexByteValue())
	}
}

func OctalByteValue() parser.Func[string] {
	return func(p *parser.Parser) string {
		if !parser.TryRune(p, '\\') {
			return ""
		}

		var i int
		s := `\` + parser.TokenWhile(p, func() bool {
			i++
			return i <= 3 && parser.MatchesRunePredicate(p, Octal_Digit)
		})
		if len(s) < 4 {
			return ""
		}
		return s
	}
}

func HexByteValue() parser.Func[string] {
	return func(p *parser.Parser) string {
		if !parser.TryToken(p, `\x`) {
			return ""
		}

		var i int
		s := `\x` + parser.TokenWhile(p, func() bool {
			i++
			return i <= 2 && parser.MatchesRunePredicate(p, Hex_Digit)
		})
		if len(s) < 4 {
			return ""
		}

		return s
	}
}

func UnicodeChar(except rune) parser.Func[rune] {
	return func(p *parser.Parser) rune {
		return parser.TryRunePredicate(p, func(r rune) bool {
			return r != except && Unicode_Char(r)
		})
	}
}

func LittleUValue() parser.Func[string] {
	return func(p *parser.Parser) string {
		if !parser.TryToken(p, `\u`) {
			return ""
		}

		var i int
		s := `\u` + parser.TokenWhile(p, func() bool {
			i++
			return i <= 4 && parser.MatchesRunePredicate(p, Hex_Digit)
		})
		if len(s) < 6 {
			return ""
		}
		return s
	}
}

func BigUValue() parser.Func[string] {
	return func(p *parser.Parser) string {
		if !parser.TryToken(p, `\U`) {
			return ""
		}

		var i int
		s := `\U` + parser.TokenWhile(p, func() bool {
			i++
			return i <= 8 && parser.MatchesRunePredicate(p, Hex_Digit)
		})
		if len(s) < 10 {
			return ""
		}
		return s
	}
}

func EscapedChar(term rune) parser.Func[string] {
	return func(p *parser.Parser) string {
		if !parser.TryRune(p, '\\') {
			return ""
		}

		r := parser.TryAnyRune(p, 'a', 'b', 'f', 'n', 'r', 't', 'v', '\\', term)
		if r == 0 {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "invalid escaped character",
				Primary: quickanno.Expected(p, p.Pos(), "a valid escaped character"),
			})
			return `\`
		}

		return `\` + string(r)
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
		var s ast.StaticString
		s.Quote = '`'

		s.Open = parser.TryRuneAt(p, '`')
		if s.Open == nil {
			return nil
		}

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
		var s ast.StaticString
		s.Quote = '"'

		s.Open = parser.TryRuneAt(p, '"')
		if s.Open == nil {
			return nil
		}

		index := p.Index()
		//nolint:revive
		for parser.TryInOrder(p, ByteValue(), UnicodeValue('"')) != "" {
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
