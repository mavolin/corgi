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

		if eh := parser.Try(p, EscapedHash()); eh != nil {
			return eh
		} else if hs := parser.Try(p, HashSpace()); hs != nil {
			return hs
		} else if hs := parser.Try(p, EscapedRBracket()); hs != nil {
			return hs
		} else if ei := parser.Try(p, ExpressionInterpolation()); ei != nil {
			return ei
		} else if cc := parser.Try(p, ComponentCallInterpolation()); cc != nil {
			return cc
		} else if cr := parser.Try(p, CharacterReference()); cr != nil {
			return cr
		} else if ei := parser.Try(p, ElementInterpolation()); ei != nil {
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

		if eh := parser.Try(p, EscapedHash()); eh != nil {
			return eh
		} else if ei := parser.Try(p, ExpressionInterpolation()); ei != nil {
			return ei
		} else if cc := parser.Try(p, ComponentCallInterpolation()); cc != nil {
			return cc
		} else if cr := parser.Try(p, CharacterReference()); cr != nil {
			return cr
		}

		// TryErr other kinds of interpolation, that aren't allowed inside a string
		if hs := parser.Try(p, HashSpace()); hs != nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "string interpolation: cannot use hash space here",
				Primary: []diagnostic.Annotation{
					anno.Range(p.File, hs.Start(), hs.End(),
						"there is no point in using a hash space, you can just write a space instead"),
				},
				Explanation: "A hash space is used to insert a trailing space in text blocks. " +
					"This isn't necessary in strings, and you can just as well write a regular space instead.",
			})
			return &ast.BadInterpolation{From: hs.Start(), Until: hs.End()}
		} else if hr := parser.Try(p, EscapedRBracket()); hr != nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "string interpolation: cannot use escaped right bracket here",
				Primary: []diagnostic.Annotation{
					anno.Range(p.File, hr.Start(), hr.End(),
						"there is no point in using an escaped right bracket, you can just write a right bracket instead"),
				},
				Explanation: "An escaped right bracket is an escape sequence available in bracket text " +
					"so that one can write a `]` without terminating the bracket text." +
					"Since `]` is not a control character in strings, " +
					"you can just as well write a regular right bracket instead.",
			})
		}
		// I don't see a reason why someone would use an element interpolation
		// in a string, so don't bother checking, especially since "#mdash foo"
		// is a valid element interpolation, but is, far more likely, supposed
		// to be a character reference lacking a semicolon.

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

func EscapedHash() parser.Func[*ast.EscapedHash] {
	return func(p *parser.Parser) *ast.EscapedHash {
		hash := parser.TryTokenAt(p, "##")
		if hash == nil {
			return nil
		}

		return &ast.EscapedHash{Hash: hash}
	}
}

func HashSpace() parser.Func[*ast.HashSpace] {
	return func(p *parser.Parser) *ast.HashSpace {
		hs := parser.TryTokenAt(p, "#_")
		if hs == nil {
			return nil
		}

		return &ast.HashSpace{Hash: hs}
	}
}

func EscapedRBracket() parser.Func[*ast.EscapedRBracket] {
	return func(p *parser.Parser) *ast.EscapedRBracket {
		erb := parser.TryTokenAt(p, "#]")
		if erb == nil {
			return nil
		}

		return &ast.EscapedRBracket{Hash: erb}
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

var elementHeader parser.Func[*ast.ElementHeader]

func SetElementHeader(f parser.Func[*ast.ElementHeader]) {
	elementHeader = f
}

func ElementInterpolation() parser.Func[*ast.ElementInterpolation] {
	return func(p *parser.Parser) *ast.ElementInterpolation {
		hash := parser.TryRuneAt(p, '#')
		if hash == nil {
			return nil
		}

		var ei ast.ElementInterpolation
		ei.Hash = hash

		var header *ast.ElementHeader
		p.DoInline(func() { header = parser.Try(p, elementHeader) })
		if header == nil {
			return nil
		}

		ei.Element = &ast.Element{Header: header}
		p.DoInline(func() {
			bt := parser.Try(p, body.BracketText())
			if bt != nil {
				ei.Element.Body = bt
			}
		})
		return &ei
	}
}

var componentCallHeader parser.Func[*ast.ComponentCallHeader]

func SetComponentCallHeader(f parser.Func[*ast.ComponentCallHeader]) {
	componentCallHeader = f
}

func ComponentCallInterpolation() parser.Func[*ast.ComponentCallInterpolation] {
	return func(p *parser.Parser) *ast.ComponentCallInterpolation {
		hash := parser.TryRuneAt(p, '#')
		if hash == nil {
			return nil
		}

		colon := parser.TryRuneAt(p, ':')

		header := parser.Try(p, componentCallHeader)
		if header == nil {
			return nil
		}

		var cci ast.ComponentCallInterpolation
		cci.Hash = hash
		cci.ComponentCall = &ast.ComponentCall{
			Colon:  colon,
			Header: header,
		}

		p.DoInline(func() {
			bt := parser.Try(p, body.BracketText())
			if bt != nil {
				cci.ComponentCall.Body = &ast.DefaultBlockShorthand{
					Implicit: true,
					Body:     bt,
					Position: bt.LBracket,
				}
			}
		})

		return &cci
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

func isInRange(s, e rune) func(rune) bool {
	return func(r rune) bool { return r >= s && r <= e }
}
