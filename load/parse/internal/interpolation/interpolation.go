// Package interpolation implements generic interpolation parsers.
package interpolation

import (
	"github.com/mavolin/corgi/v2/escape/charref"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/body"
	"github.com/mavolin/corgi/v2/load/parse/internal/html/codepoint"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/whitespace"
)

func TextInterpolation() parser.Func[ast.TextInterpolation] {
	return func(p *parser.Parser) ast.TextInterpolation {
		if !parser.MatchesAnyRune(p, '#') {
			return nil
		}

		if ce := parser.Try(p, TextCharacterEscape()); ce != nil {
			return ce
		} else if ei := parser.Try(p, ExpressionInterpolation()); ei != nil {
			return ei
		} else if cr := parser.Try(p, CharacterReference()); cr != nil {
			return cr
		} else if ei := parser.Try(p, ModeSwitch()); ei != nil {
			return ei
		}

		p.CaptureError(&diagnostic.Diagnostic{
			Message: "bad interpolation",
			Primary: []diagnostic.Annotation{
				anno.Position(p.File, p.Pos(), "expected a valid interpolation, but found this"),
			},
			Hints: []diagnostic.Hint{
				{Hint: "If you just wanted to write hash, you need to escape it.", Example: "`##`"},
			},
			Examples: []diagnostic.Example{
				{Title: "escaped hash", Example: "`##`"},
				{Title: "hash space", Example: "`#_`"},
				{Title: "escaped right bracket", Example: "`#]`"},
				{Title: "expression interpolation", Example: "`#{1 + 1}`"},
				{Title: "component call interpolation", Example: "`#:fmt.Number(val: 21_000)`"},
				{Title: "character reference", Example: "`#amp;`"},
				{Title: "element interpolation", Example: "`#br` or #strong[foo]"},
			},
		})
		bi := parser.Try(p, BadInterpolation())
		if bi == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Type:    diagnostic.InternalError,
				Message: "string interpolation: bad interpolation captured nothing",
				Primary: quickanno.Expected(p, p.Pos(), "a bad interpolation"),
			})
		}
		return bi
	}
}

func StringInterpolation() parser.Func[ast.StringInterpolation] {
	return func(p *parser.Parser) ast.StringInterpolation {
		if !parser.MatchesAnyRune(p, '#') {
			return nil
		}

		if ce := parser.Try(p, StringCharacterEscape()); ce != nil {
			return ce
		} else if ei := parser.Try(p, ExpressionInterpolation()); ei != nil {
			return ei
		} else if cr := parser.Try(p, CharacterReference()); cr != nil {
			return cr
		} else if cc := parser.Try(p, ComponentCallInterpolation()); cc != nil {
			return cc
		}

		// TryErr other kinds of interpolation, that aren't allowed inside a string
		if ce := parser.Try(p, TextCharacterEscape()); ce != nil {
			switch ce.Symbol {
			case '_':
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "string interpolation: cannot use hash space here",
					Primary: []diagnostic.Annotation{
						anno.Range(p.File, ce.Start(), ce.End(),
							"there is no point in using a hash space, you can just write a space instead"),
					},
					Explanation: "A hash space is used to insert a trailing space in text blocks. " +
						"This isn't necessary in strings, and you can just as well write a regular space instead.",
				})
			case ']':
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "string interpolation: cannot use escaped right bracket here",
					Primary: []diagnostic.Annotation{
						anno.Range(p.File, ce.Start(), ce.End(),
							"there is no point in using an escaped right bracket, you can just write a right bracket instead"),
					},
					Explanation: "An escaped right bracket is an escape sequence available in bracket text " +
						"so that one can write a `]` without terminating the bracket text." +
						"Since `]` is not a control character in strings, " +
						"you can just as well write a regular right bracket instead.",
				})
			default:
				p.CaptureError(&diagnostic.Diagnostic{
					Type:    diagnostic.InternalError,
					Message: "string interpolation: recovered text character escape: unknown symbol",
					Primary: []diagnostic.Annotation{
						anno.Range(p.File, ce.Start(), ce.End(), "could not identify symbol"),
					},
					Explanation: "Could not identify the symbol to give you a better error message. " +
						"This is a bug in the parser, please report it.",
				})
			}
			return &ast.BadInterpolation{From: ce.Start(), Until: ce.End()}
		}

		p.CaptureError(&diagnostic.Diagnostic{
			Message: "bad interpolation",
			Primary: []diagnostic.Annotation{
				anno.NRunes(p.File, p.Pos(), 1, "expected a valid interpolation, but found this"),
			},
			Hints: []diagnostic.Hint{
				{Hint: "If you just wanted to use hash, you need to escape it.", Example: "`##`"},
			},
			Examples: []diagnostic.Example{
				{Title: "escaped hash", Example: "`##`"},
				{Title: "expression interpolation", Example: "`#{1 + 1}`"},
				{Title: "component call interpolation", Example: "`#:fmt.Number(val: 21_000)`"},
				{Title: "character reference", Example: "`#amp;`"},
			},
		})
		bi := parser.Try(p, BadInterpolation())
		if bi == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Type:    diagnostic.InternalError,
				Message: "string interpolation: bad interpolation captured nothing",
				Primary: quickanno.Expected(p, p.Pos(), "a bad interpolation"),
			})
		}
		return bi
	}
}

// BadInterpolation consumes any single hash as a bad interpolation.
// It is the caller's responsibility to ensure that the hash does not actually
// represent a valid interpolation.
// The returned BadInterpolation will always be a single rune long.
//
// Likewise, it is the caller's responsibility to capture an error indicating
// that a valid interpolation was expected.
func BadInterpolation() parser.Func[*ast.BadInterpolation] {
	return func(p *parser.Parser) *ast.BadInterpolation {
		from := parser.TryRuneAt(p, '#')
		if from == nil {
			return nil
		}

		return &ast.BadInterpolation{From: *from, Until: p.Pos()}
	}
}

func StringCharacterEscape() parser.Func[*ast.CharacterEscape] {
	return func(p *parser.Parser) *ast.CharacterEscape {
		hash := parser.TryRuneAt(p, '#')
		if hash == nil {
			return nil
		}

		symbol := parser.TryAnyRune(p, '#')
		if symbol == 0 {
			return nil
		}

		return &ast.CharacterEscape{Hash: hash, Symbol: symbol, Rune: symbol}
	}
}

func TextCharacterEscape() parser.Func[*ast.CharacterEscape] {
	return func(p *parser.Parser) *ast.CharacterEscape {
		hash := parser.TryRuneAt(p, '#')
		if hash == nil {
			return nil
		}

		symbol := parser.TryAnyRune(p, '#', '_', ']')
		if symbol == 0 {
			return nil
		}

		var r rune
		switch symbol {
		case '_':
			r = ' '
		default:
			r = symbol
		}

		return &ast.CharacterEscape{Hash: hash, Symbol: symbol, Rune: r}
	}
}

func VerbatimTextCharacterEscape() parser.Func[*ast.CharacterEscape] {
	return func(p *parser.Parser) *ast.CharacterEscape {
		hash := parser.TryRuneAt(p, '#')
		if hash == nil {
			return nil
		}

		symbol := parser.TryAnyRune(p, ']')
		if symbol == 0 {
			return nil
		}

		return &ast.CharacterEscape{Hash: hash, Symbol: symbol, Rune: symbol}
	}
}

func UnambiguousHash() parser.Func[bool] {
	return func(p *parser.Parser) bool {
		if !parser.TryRune(p, '#') {
			return false
		}

		if (p.Inline() && parser.TryAnyRune(p, whitespace.HorizontalRunes...) == 0) ||
			(!p.Inline() && parser.TryAnyRune(p, whitespace.Runes...) == 0) {
			return false
		}
		return true
	}
}

func CharacterReference() parser.Func[*ast.CharacterReference] {
	return func(p *parser.Parser) *ast.CharacterReference {
		hash := parser.TryTokenAt(p, "#")
		if hash == nil {
			return nil
		}

		var r ast.CharacterReference
		r.Hash = hash

		r.Name = parser.TokenWhile(p, func() bool {
			return parser.MatchesRunePredicate(p, codepoint.ASCIIAlphanumeric)
		})
		if !parser.TryRune(p, ';') {
			return nil
		}
		if r.Name == "" {
			p.CaptureError(&diagnostic.Diagnostic{
				Message:  "character reference: missing name",
				Primary:  quickanno.Expected(p, p.Pos(), "a character reference name"),
				Examples: []diagnostic.Example{{Example: "`#amp;` or `#mdash;`"}},
				Hints: []diagnostic.Hint{
					{Hint: "If you just wanted to write hash, you need to escape it.", Example: "`##`"},
				},
			})
			return &r
		}

		r.Chars = charref.Chars(r.Name)
		if r.Chars == "" {
			p.CaptureError(&diagnostic.Diagnostic{
				Message:  "character reference: unknown name",
				Primary:  quickanno.Expected(p, p.Pos(), "a valid character reference name"),
				Examples: []diagnostic.Example{{Example: "`#amp;` or `#mdash;`"}},
				Hints: []diagnostic.Hint{
					{Hint: "If you just wanted to write hash, you need to escape it.", Example: "`##`"},
				},
			})
		}

		return &r
	}
}

func ModeSwitch() parser.Func[*ast.ModeSwitch] {
	return func(p *parser.Parser) *ast.ModeSwitch {
		hash := parser.TryRuneAt(p, '#')
		if hash == nil {
			return nil
		}

		node := parser.Try(p, body.ScopeNode())
		if node == nil {
			return nil
		} else if _, ok := node.(*ast.ArrowBlock); ok {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "mode switch: useless switch to arrow block",
				Primary: []diagnostic.Annotation{
					anno.Node(p.File, node, "no point in switching to an arrow block here"),
				},
				Explanation: "You are already in tex mode, there is no point in using an arrow block here.",
			})
		}

		return &ast.ModeSwitch{Hash: hash, Node: node}
	}
}

var componentCall parser.Func[*ast.ComponentCall]

func SetComponentCall(f parser.Func[*ast.ComponentCall]) {
	componentCall = f
}

func ComponentCallInterpolation() parser.Func[*ast.ComponentCallInterpolation] {
	return func(p *parser.Parser) *ast.ComponentCallInterpolation {
		hash := parser.TryRuneAt(p, '#')
		if hash == nil {
			return nil
		}

		cc := parser.Try(p, componentCall)
		if cc == nil {
			return nil
		}

		return &ast.ComponentCallInterpolation{Hash: hash, ComponentCall: cc}
	}
}

var expression parser.Func[*ast.Expression]

func SetExpression(f parser.Func[*ast.Expression]) {
	expression = f
}

func ExpressionInterpolation() parser.Func[*ast.ExpressionInterpolation] {
	return func(p *parser.Parser) *ast.ExpressionInterpolation {
		hash := parser.TryRuneAt(p, '#')
		if hash == nil {
			return nil
		}

		var ei ast.ExpressionInterpolation
		ei.Hash = hash

		ei.LBrace = parser.TryRuneAt(p, '{')
		if ei.LBrace == nil {
			return nil
		}

		parser.TrySkip(p, whitespace.Horizontal())
		ei.Expression = parser.Try(p, expression)
		if ei.Expression == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "expression interpolation: missing expression",
				Primary: quickanno.Expected(p, p.Pos(), "an expression"),
				Examples: []diagnostic.Example{
					{Example: "`#{1 + 1}`"},
				},
			})
		}
		parser.TrySkip(p, whitespace.Horizontal())

		ei.RBrace = parser.TryRuneAt(p, '}')
		if ei.RBrace == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "expression interpolation: missing closing brace",
				Primary: quickanno.Expected(p, ei.End(), "a closing brace `}`"),
			})
			return &ei
		}

		return &ei
	}
}
