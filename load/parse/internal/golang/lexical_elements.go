package golang

import (
	"fmt"

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
	return func(p *parser.Parser) (string, *diagnostic.Diagnostic) {
		if !parser.TryRune(p, '\'') {
			return "", &diagnostic.Diagnostic{
				Message: "missing rune literal",
				Primary: quickanno.Expected(p, p.Pos(), "a rune literal"),
			}
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

		return "'" + v + "'", nil
	}
}

func UnicodeValue(term rune) parser.Func[string] {
	return func(p *parser.Parser) (string, *diagnostic.Diagnostic) {
		if parser.MatchesAnyRune(p, '\\') {
			s := parser.TryInOrder(p, LittleUValue(), BigUValue(), EscapedChar(term))
			if s != "" {
				return s, nil
			}
		} else {
			r := parser.Try(p, UnicodeChar(term))
			if r != 0 {
				return string(r), nil
			}
		}

		return "", &diagnostic.Diagnostic{
			Message: "missing unicode value",
			Primary: quickanno.Expected(p, p.Pos(), "a unicode value"),
		}
	}
}

func ByteValue() parser.Func[string] {
	return func(p *parser.Parser) (string, *diagnostic.Diagnostic) {
		s := parser.TryInOrder(p, OctalByteValue(), HexByteValue())
		if s == "" {
			return "", &diagnostic.Diagnostic{
				Message:  "missing byte value",
				Primary:  quickanno.Expected(p, p.Pos(), "a byte value"),
				Examples: []diagnostic.Example{{Example: "`\\x12` or `\\123`"}},
			}
		}
		return s, nil
	}
}

func OctalByteValue() parser.Func[string] {
	return func(p *parser.Parser) (string, *diagnostic.Diagnostic) {
		if !parser.TryRune(p, '\\') {
			return "", &diagnostic.Diagnostic{
				Message: "missing octal byte value",
				Primary: quickanno.Expected(p, p.Pos(), "an octal byte value"),
			}
		}

		octalStart := p.Pos()
		var i int
		s := `\` + parser.TokenWhile(p, func() bool {
			i++
			return i <= 3 && parser.MatchesRunePredicate(p, Octal_Digit)
		})
		if len(s) < 4 {
			return "", &diagnostic.Diagnostic{
				Message:  "octal byte value: missing digits",
				Primary:  quickanno.Expected(p, octalStart, fmt.Sprint("3 octal digits, found ", len(s)-1)),
				Examples: []diagnostic.Example{{Example: `\123`}},
			}
		}
		return s, nil
	}
}

func HexByteValue() parser.Func[string] {
	return func(p *parser.Parser) (string, *diagnostic.Diagnostic) {
		if !parser.TryToken(p, `\x`) {
			return "", &diagnostic.Diagnostic{
				Message: "missing hex byte value",
				Primary: quickanno.Expected(p, p.Pos(), "a hex byte value"),
			}
		}

		hexStart := p.Pos()
		var i int
		s := `\x` + parser.TokenWhile(p, func() bool {
			i++
			return i <= 2 && parser.MatchesRunePredicate(p, Hex_Digit)
		})
		if len(s) < 4 {
			return "", &diagnostic.Diagnostic{
				Message:  "hex byte value: missing digits",
				Primary:  quickanno.Expected(p, hexStart, fmt.Sprint("2 hexadecimal digits, found ", len(s)-2)),
				Examples: []diagnostic.Example{{Example: `\x12`}},
			}
		}

		return s, nil
	}
}

func UnicodeChar(except rune) parser.Func[rune] {
	return func(p *parser.Parser) (rune, *diagnostic.Diagnostic) {
		r := parser.TryRunePredicate(p, func(r rune) bool {
			return r != except && Unicode_Char(r)
		})
		if r < 0 {
			return 0, &diagnostic.Diagnostic{
				Message: "missing unicode character, except newline",
				Primary: quickanno.Expected(p, p.Pos(), "a unicode character except a newline"),
			}
		}
		return r, nil
	}
}

func LittleUValue() parser.Func[string] {
	return func(p *parser.Parser) (string, *diagnostic.Diagnostic) {
		if !parser.TryToken(p, `\u`) {
			return "", &diagnostic.Diagnostic{
				Message: "missing little u value",
				Primary: quickanno.Expected(p, p.Pos(), "a little u value"),
			}
		}

		hexStart := p.Pos()
		var i int
		s := `\u` + parser.TokenWhile(p, func() bool {
			i++
			return i <= 4 && parser.MatchesRunePredicate(p, Hex_Digit)
		})
		if len(s) < 6 {
			return "", &diagnostic.Diagnostic{
				Message:  "little u value: missing digits",
				Primary:  quickanno.Expected(p, hexStart, fmt.Sprint("4 hexadecimal digits, found ", len(s)-2)),
				Examples: []diagnostic.Example{{Example: `\u1234`}},
			}
		}
		return s, nil
	}
}

func BigUValue() parser.Func[string] {
	return func(p *parser.Parser) (string, *diagnostic.Diagnostic) {
		if !parser.TryToken(p, `\U`) {
			return "", &diagnostic.Diagnostic{
				Message: "missing big u value",
				Primary: quickanno.Expected(p, p.Pos(), "a big u value"),
			}
		}

		hexStart := p.Pos()
		var i int
		s := `\U` + parser.TokenWhile(p, func() bool {
			i++
			return i <= 8 && parser.MatchesRunePredicate(p, Hex_Digit)
		})
		if len(s) < 10 {
			return "", &diagnostic.Diagnostic{
				Message:  "big u value: missing digits",
				Primary:  quickanno.Expected(p, hexStart, fmt.Sprint("8 hexadecimal digits, found ", len(s)-2)),
				Examples: []diagnostic.Example{{Example: `\U12345678`}},
			}
		}
		return s, nil
	}
}

func EscapedChar(term rune) parser.Func[string] {
	return func(p *parser.Parser) (string, *diagnostic.Diagnostic) {
		if !parser.TryRune(p, '\\') {
			return "", &diagnostic.Diagnostic{
				Message:  "missing escaped character",
				Primary:  quickanno.Expected(p, p.Pos(), "an escaped character"),
				Examples: []diagnostic.Example{{Example: "`\\n` or `\\t`"}},
			}
		}

		r := parser.TryAnyRune(p, 'a', 'b', 'f', 'n', 'r', 't', 'v', '\\', term)
		if r < 0 {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "invalid escaped character",
				Primary: quickanno.Expected(p, p.Pos(), "a valid escaped character"),
			})
			return `\`, nil
		}

		return `\` + string(r), nil
	}
}

// ============================================================================
// String literals
// ======================================================================================

func StringLit() parser.Func[*ast.StaticString] {
	return func(p *parser.Parser) (*ast.StaticString, *diagnostic.Diagnostic) {
		s := parser.TryInOrder(p, RawStringLit(), InterpretedStringLit())
		if s == nil {
			return nil, &diagnostic.Diagnostic{
				Message:  "missing string literal",
				Primary:  quickanno.Expected(p, p.Pos(), "a string literal"),
				Examples: []diagnostic.Example{{Example: "`\"woof\"`"}},
			}
		}
		return s, nil
	}
}

func RawStringLit() parser.Func[*ast.StaticString] {
	return func(p *parser.Parser) (*ast.StaticString, *diagnostic.Diagnostic) {
		var s ast.StaticString
		s.Quote = '`'

		s.Open = parser.TryRuneAt(p, '`')
		if s.Open == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "missing raw string literal",
				Primary: quickanno.Expected(p, p.Pos(), "a raw string literal"),
			}
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
		return &s, nil
	}
}

func InterpretedStringLit() parser.Func[*ast.StaticString] {
	return func(p *parser.Parser) (*ast.StaticString, *diagnostic.Diagnostic) {
		var s ast.StaticString
		s.Quote = '"'

		s.Open = parser.TryRuneAt(p, '"')
		if s.Open == nil {
			return nil, &diagnostic.Diagnostic{
				Message:  "missing interpreted string literal",
				Primary:  quickanno.Expected(p, p.Pos(), "an interpreted string literal"),
				Examples: []diagnostic.Example{{Example: "`\"woof\"`"}},
			}
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

		return &s, nil
	}
}
