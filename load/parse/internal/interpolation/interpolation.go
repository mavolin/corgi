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
		var bi ast.BadInterpolation

		bi.From = p.Pos()
		if !parser.TryRune(p, '#') {
			return nil
		}
		bi.Until = p.Pos()

		return &bi
	}
}

func EscapedHash() parser.Func[*ast.EscapedHash] {
	return func(p *parser.Parser) *ast.EscapedHash {
		var eh ast.EscapedHash

		eh.Hash = parser.TryTokenAt(p, "##")
		if eh.Hash == nil {
			return nil
		}

		return &eh
	}
}

func HashSpace() parser.Func[*ast.HashSpace] {
	return func(p *parser.Parser) *ast.HashSpace {
		var hs ast.HashSpace

		hs.Hash = parser.TryTokenAt(p, "#_")
		if hs.Hash == nil {
			return nil
		}

		return &hs
	}
}

func EscapedRBracket() parser.Func[*ast.EscapedRBracket] {
	return func(p *parser.Parser) *ast.EscapedRBracket {
		var erb ast.EscapedRBracket

		erb.Hash = parser.TryTokenAt(p, "#]")
		if erb.Hash == nil {
			return nil
		}

		return &erb
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
		var r ast.CharacterReference

		r.Hash = parser.TryTokenAt(p, "#")
		if r.Hash == nil {
			return nil
		}

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
		var ei ast.ElementInterpolation

		ei.Hash = parser.TryRuneAt(p, '#')
		if ei.Hash == nil {
			return nil
		}

		var header *ast.ElementHeader
		p.DoInline(func() { header = parser.Try(p, elementHeader) })
		if header == nil {
			return nil
		}

		ei.Element = &ast.Element{Header: header}
		p.DoInline(func() {
			bt := parser.Try(p, body.BracketText())
			if bt != nil { // can't assign directly: any(nil) != (*ast.BracketText)(nil)
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
		var cci ast.ComponentCallInterpolation

		cci.Hash = parser.TryTokenAt(p, "#:")
		if cci.Hash == nil {
			return nil
		}
		cci.ComponentCall = new(ast.ComponentCall)
		cci.ComponentCall.Colon = new(ast.Position)
		*cci.ComponentCall.Colon = *cci.Hash
		cci.ComponentCall.Colon.Col++

		cci.ComponentCall.Header = parser.Try(p, componentCallHeader)
		if cci.ComponentCall.Header == nil {
			return nil
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
		var ei ast.ExpressionInterpolation

		ei.Hash = parser.TryRuneAt(p, '#')
		if ei.Hash == nil {
			return nil
		}

		ei.FormatDirective = parser.Try(p, formatDirective())

		ei.LBrace = parser.TryRuneAt(p, '{')
		if ei.LBrace == nil {
			if ei.FormatDirective != "" {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "expression interpolation: missing opening brace",
					Primary: quickanno.Expected(p, *ei.Hash, "an opening brace `{`"),
					Examples: []diagnostic.Example{
						{Example: "`#%" + ei.FormatDirective + "{...}`"},
					},
				})
				return &ei
			}
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

func formatDirective() parser.Func[string] {
	return func(p *parser.Parser) string {
		if !parser.TryRune(p, '%') {
			return ""
		}

		startIndex := p.Index()

		// flag
		for parser.TryAnyRune(p, '+', '-', '#', ' ', '0') > 0 { //nolint:revive
		}

		// width
		if parser.TryRunePredicate(p, isInRange('1', '9')) > 0 {
			for parser.TryRunePredicate(p, isInRange('0', '9')) > 0 { //nolint:revive
			}
		}

		// precision
		if parser.TryRune(p, '.') {
			for parser.TryRunePredicate(p, isInRange('0', '9')) > 0 { //nolint:revive
			}
		}

		// verb
		if parser.TryAnyRune(p, 'v', 'T', 't', 'b', 'c', 'd', 'o', 'O', 'x', 'X', 'U', 'e', 'E', 'f', 'F', 'g', 'G', 's', 'p') == 0 {
			if parser.TryRunePredicate(p, isInRange('a', 'z')) > 0 || parser.TryRunePredicate(p, isInRange('A', 'Z')) > 0 {
				return p.AST.Raw[startIndex:p.Index()]
			}

			return p.AST.Raw[startIndex:p.Index()]
		}

		return p.AST.Raw[startIndex:p.Index()]
	}
}

func isInRange(s, e rune) func(rune) bool {
	return func(r rune) bool { return r >= s && r <= e }
}
