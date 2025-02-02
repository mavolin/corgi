package golang

import (
	"fmt"

	"github.com/mavolin/corgi/v2/fancyerr"
	"github.com/mavolin/corgi/v2/file/ast"
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
	return func(p *parser.Parser) (string, *fancyerr.Error) {
		if !parser.TryRune(p, '\'') {
			return "", &fancyerr.Error{
				Message: "missing rune literal",
				Primary: quickanno.Expected(p, p.Pos(), "a rune literal"),
			}
		}

		v, ok := parser.TryInOrder(p, ByteValue(), UnicodeValue('\''))
		if !ok {
			p.CaptureError(&fancyerr.Error{
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
			p.CaptureError(&fancyerr.Error{
				Message: "rune literal: missing closing quote",
				Primary: quickanno.Expected(p, p.Pos(), "a closing quote for the opening quote here"),
			})
		}

		return "'" + v + "'", nil
	}
}

func UnicodeValue(term rune) parser.Func[string] {
	return func(p *parser.Parser) (string, *fancyerr.Error) {
		if parser.MatchesAnyRune(p, '\\') {
			s, ok := parser.TryInOrder(p, LittleUValue(), BigUValue(), EscapedChar(term))
			if ok {
				return s, nil
			}
		} else {
			r, ok := parser.TryOk(p, UnicodeChar(term))
			if ok {
				return string(r), nil
			}
		}

		return "", &fancyerr.Error{
			Message: "missing unicode value",
			Primary: quickanno.Expected(p, p.Pos(), "a unicode value"),
		}
	}
}

func ByteValue() parser.Func[string] {
	return func(p *parser.Parser) (string, *fancyerr.Error) {
		s, ok := parser.TryInOrder(p, OctalByteValue(), HexByteValue())
		if !ok {
			return "", &fancyerr.Error{
				Message:  "missing byte value",
				Primary:  quickanno.Expected(p, p.Pos(), "a byte value"),
				Examples: []fancyerr.Example{{Example: "`\\x12` or `\\123`"}},
			}
		}
		return s, nil
	}
}

func OctalByteValue() parser.Func[string] {
	return func(p *parser.Parser) (string, *fancyerr.Error) {
		if !parser.TryRune(p, '\\') {
			return "", &fancyerr.Error{
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
			return "", &fancyerr.Error{
				Message:  "octal byte value: missing digits",
				Primary:  quickanno.Expected(p, octalStart, fmt.Sprint("3 octal digits, found ", len(s)-1)),
				Examples: []fancyerr.Example{{Example: `\123`}},
			}
		}
		return s, nil
	}
}

func HexByteValue() parser.Func[string] {
	return func(p *parser.Parser) (string, *fancyerr.Error) {
		if !parser.TryToken(p, `\x`) {
			return "", &fancyerr.Error{
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
			return "", &fancyerr.Error{
				Message:  "hex byte value: missing digits",
				Primary:  quickanno.Expected(p, hexStart, fmt.Sprint("2 hexadecimal digits, found ", len(s)-2)),
				Examples: []fancyerr.Example{{Example: `\x12`}},
			}
		}

		return s, nil
	}
}

func UnicodeChar(except rune) parser.Func[rune] {
	return func(p *parser.Parser) (rune, *fancyerr.Error) {
		r := parser.TryRunePredicate(p, func(r rune) bool {
			return r != except && Unicode_Char(r)
		})
		if r < 0 {
			return 0, &fancyerr.Error{
				Message: "missing unicode character, except newline",
				Primary: quickanno.Expected(p, p.Pos(), "a unicode character except a newline"),
			}
		}
		return r, nil
	}
}

func LittleUValue() parser.Func[string] {
	return func(p *parser.Parser) (string, *fancyerr.Error) {
		if !parser.TryToken(p, `\u`) {
			return "", &fancyerr.Error{
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
			return "", &fancyerr.Error{
				Message:  "little u value: missing digits",
				Primary:  quickanno.Expected(p, hexStart, fmt.Sprint("4 hexadecimal digits, found ", len(s)-2)),
				Examples: []fancyerr.Example{{Example: `\u1234`}},
			}
		}
		return s, nil
	}
}

func BigUValue() parser.Func[string] {
	return func(p *parser.Parser) (string, *fancyerr.Error) {
		if !parser.TryToken(p, `\U`) {
			return "", &fancyerr.Error{
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
			return "", &fancyerr.Error{
				Message:  "big u value: missing digits",
				Primary:  quickanno.Expected(p, hexStart, fmt.Sprint("8 hexadecimal digits, found ", len(s)-2)),
				Examples: []fancyerr.Example{{Example: `\U12345678`}},
			}
		}
		return s, nil
	}
}

func EscapedChar(term rune) parser.Func[string] {
	return func(p *parser.Parser) (string, *fancyerr.Error) {
		if !parser.TryRune(p, '\\') {
			return "", &fancyerr.Error{
				Message:  "missing escaped character",
				Primary:  quickanno.Expected(p, p.Pos(), "an escaped character"),
				Examples: []fancyerr.Example{{Example: "`\\n` or `\\t`"}},
			}
		}

		r := parser.TryAnyRune(p, 'a', 'b', 'f', 'n', 'r', 't', 'v', '\\', term)
		if r < 0 {
			p.CaptureError(&fancyerr.Error{
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
	return func(p *parser.Parser) (*ast.StaticString, *fancyerr.Error) {
		s, ok := parser.TryInOrder(p, RawStringLit(), InterpretedStringLit())
		if !ok {
			return nil, &fancyerr.Error{
				Message:  "missing string literal",
				Primary:  quickanno.Expected(p, p.Pos(), "a string literal"),
				Examples: []fancyerr.Example{{Example: "`\"woof\"`"}},
			}
		}
		return s, nil
	}
}

func RawStringLit() parser.Func[*ast.StaticString] {
	return func(p *parser.Parser) (*ast.StaticString, *fancyerr.Error) {
		s := &ast.StaticString{Open: p.Pos(), Quote: '`'}

		if !parser.TryRune(p, '`') {
			return nil, &fancyerr.Error{
				Message: "missing raw string literal",
				Primary: quickanno.Expected(p, p.Pos(), "a raw string literal"),
			}
		}

		s.Contents = parser.TokenWhile(p, func() bool {
			return !parser.MatchesAnyRune(p, '`')
		})

		s.Close = p.PosPtr()
		if !parser.TryRune(p, '`') {
			s.Close = nil
			p.CaptureError(&fancyerr.Error{
				Message: "raw string literal: missing closing backtick",
				Primary: quickanno.Expected(p, s.Open, "a closing backtick for the opening backtick here"),
			})
		}
		return s, nil
	}
}

func InterpretedStringLit() parser.Func[*ast.StaticString] {
	return func(p *parser.Parser) (*ast.StaticString, *fancyerr.Error) {
		s := &ast.StaticString{Open: p.Pos(), Quote: '"'}

		if !parser.TryRune(p, '"') {
			return nil, &fancyerr.Error{
				Message:  "missing interpreted string literal",
				Primary:  quickanno.Expected(p, p.Pos(), "an interpreted string literal"),
				Examples: []fancyerr.Example{{Example: "`\"woof\"`"}},
			}
		}

		index := p.Index()
		for {
			_, ok := parser.TryInOrder(p, ByteValue(), UnicodeValue('"'))
			if !ok {
				break
			}
		}
		s.Contents = p.File.Raw[index:p.Index()]

		s.Close = p.PosPtr()
		if !parser.TryRune(p, '"') {
			s.Close = nil
			p.CaptureError(&fancyerr.Error{
				Message: "interpreted string literal: missing closing quote",
				Primary: quickanno.Expected(p, s.Open, "a closing quote for the opening quote here"),
			})
		}

		return s, nil
	}
}
