// Package interpolation implements generic interpolation parsers.
package interpolation

import (
	"github.com/mavolin/corgi/v2/escape/charref"
	"github.com/mavolin/corgi/v2/fancyerr"
	"github.com/mavolin/corgi/v2/fancyerr/anno"
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/body"
	"github.com/mavolin/corgi/v2/load/parse/internal/html/codepoint"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/whitespace"
)

func TextInterpolation() parser.Func[ast.TextInterpolation] {
	return func(p *parser.Parser) (ast.TextInterpolation, *fancyerr.Error) {
		if err := ensureInterpolation(p); err != nil {
			return nil, err
		}

		if eh := parser.Try(p, EscapedHash()); eh != nil {
			return eh, nil
		} else if hs := parser.Try(p, HashSpace()); hs != nil {
			return hs, nil
		} else if hs := parser.Try(p, EscapedRBracket()); hs != nil {
			return hs, nil
		} else if ei := parser.Try(p, ExpressionInterpolation()); ei != nil {
			return ei, nil
		} else if cc := parser.Try(p, ComponentCallInterpolation()); cc != nil {
			return cc, nil
		} else if cr := parser.Try(p, CharacterReference()); cr != nil {
			return cr, nil
		} else if ei := parser.Try(p, ElementInterpolation()); ei != nil {
			return ei, nil
		}

		p.CaptureError(&fancyerr.Error{
			Message: "bad interpolation",
			Primary: []fancyerr.Annotation{
				anno.Position(p.File, p.Pos(), "expected a valid interpolation, but found this"),
			},
			Hints: []fancyerr.Hint{
				{Hint: "If you just wanted to write hash, you need to escape it.", Example: "`##`"},
			},
			Examples: []fancyerr.Example{
				{Title: "escaped hash", Example: "`##`"},
				{Title: "hash space", Example: "`#_`"},
				{Title: "escaped right bracket", Example: "`#]`"},
				{Title: "expression interpolation", Example: "`#{1 + 1}`"},
				{Title: "component call interpolation", Example: "`#:fmt.Number(val: 21_000)`"},
				{Title: "character reference", Example: "`#amp;`"},
				{Title: "element interpolation", Example: "`#br` or #strong[foo]"},
			},
		})
		return parser.TryErr(p, BadInterpolation())
	}
}

func StringInterpolation() parser.Func[ast.StringInterpolation] {
	return func(p *parser.Parser) (ast.StringInterpolation, *fancyerr.Error) {
		if err := ensureInterpolation(p); err != nil {
			return nil, err
		}

		if eh := parser.Try(p, EscapedHash()); eh != nil {
			return eh, nil
		} else if ei := parser.Try(p, ExpressionInterpolation()); ei != nil {
			return ei, nil
		} else if cc := parser.Try(p, ComponentCallInterpolation()); cc != nil {
			return cc, nil
		} else if cr := parser.Try(p, CharacterReference()); cr != nil {
			return cr, nil
		}

		// TryErr other kinds of interpolation, that aren't allowed inside a string
		if hs := parser.Try(p, HashSpace()); hs != nil {
			p.CaptureError(&fancyerr.Error{
				Message: "string interpolation: cannot use hash space here",
				Primary: []fancyerr.Annotation{
					anno.Range(p.File, hs.Start(), hs.End(),
						"there is no point in using a hash space, you can just write a space instead"),
				},
				Explanation: "A hash space is used to insert a trailing space in text blocks. " +
					"This isn't necessary in strings, and you can just as well write a regular space instead.",
			})
			return &ast.BadInterpolation{From: hs.Start(), Until: hs.End()}, nil
		} else if hr := parser.Try(p, EscapedRBracket()); hr != nil {
			p.CaptureError(&fancyerr.Error{
				Message: "string interpolation: cannot use escaped right bracket here",
				Primary: []fancyerr.Annotation{
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

		p.CaptureError(&fancyerr.Error{
			Message: "bad interpolation",
			Primary: []fancyerr.Annotation{
				anno.NChars(p.File, p.Pos(), 1, "expected a valid interpolation, but found this"),
			},
			Hints: []fancyerr.Hint{
				{Hint: "If you just wanted to use hash, you need to escape it.", Example: "`##`"},
			},
			Examples: []fancyerr.Example{
				{Title: "escaped hash", Example: "`##`"},
				{Title: "expression interpolation", Example: "`#{1 + 1}`"},
				{Title: "component call interpolation", Example: "`#:fmt.Number(val: 21_000)`"},
				{Title: "character reference", Example: "`#amp;`"},
			},
		})
		return parser.TryErr(p, BadInterpolation())
	}
}

func ensureInterpolation(p *parser.Parser) *fancyerr.Error {
	if !parser.MatchesToken(p, "#") {
		return &fancyerr.Error{
			Message: "missing interpolation",
			Primary: quickanno.Expected(p, p.Pos(), "a hash, starting an interpolation"),
		}
	}
	return nil
}

// BadInterpolation consumes any single hash as a bad interpolation.
// It is the caller's responsibility to ensure that the hash does not actually
// represent a valid interpolation.
// The returned BadInterpolation will always be a single rune long.
//
// Likewise, it is the caller's responsibility to capture an error indicating
// that a valid interpolation was expected.
func BadInterpolation() parser.Func[*ast.BadInterpolation] {
	return func(p *parser.Parser) (*ast.BadInterpolation, *fancyerr.Error) {
		var bi ast.BadInterpolation

		bi.From = p.Pos()
		if !parser.TryRune(p, '#') {
			return nil, &fancyerr.Error{
				Message: "missing bad interpolation",
			}
		}
		bi.Until = p.Pos()

		return &bi, nil
	}
}

func EscapedHash() parser.Func[*ast.EscapedHash] {
	return func(p *parser.Parser) (*ast.EscapedHash, *fancyerr.Error) {
		var eh ast.EscapedHash

		eh.Hash = parser.TryTokenAt(p, "##")
		if eh.Hash == nil {
			return nil, &fancyerr.Error{
				Message: "missing escaped hash",
				Primary: quickanno.Expected(p, p.Pos(), "an escaped hash (`##`)"),
			}
		}

		return &eh, nil
	}
}

func HashSpace() parser.Func[*ast.HashSpace] {
	return func(p *parser.Parser) (*ast.HashSpace, *fancyerr.Error) {
		var hs ast.HashSpace

		hs.Hash = parser.TryTokenAt(p, "#_")
		if hs.Hash == nil {
			return nil, &fancyerr.Error{
				Message: "missing hash space",
				Primary: quickanno.Expected(p, p.Pos(), "a hash followed by an underline (`#_`)"),
			}
		}

		return &hs, nil
	}
}

func EscapedRBracket() parser.Func[*ast.EscapedRBracket] {
	return func(p *parser.Parser) (*ast.EscapedRBracket, *fancyerr.Error) {
		var erb ast.EscapedRBracket

		erb.Hash = parser.TryTokenAt(p, "#]")
		if erb.Hash == nil {
			return nil, &fancyerr.Error{
				Message: "missing hash right bracket",
				Primary: quickanno.Expected(p, p.Pos(), "a hash followed by a right bracket (`#]`)"),
			}
		}

		return &erb, nil
	}
}

func UnambiguousHash() parser.Func[struct{}] {
	return func(p *parser.Parser) (struct{}, *fancyerr.Error) {
		if !parser.TryRune(p, '#') {
			return struct{}{}, &fancyerr.Error{
				Message: "missing unambiguous hash",
				Primary: quickanno.Expected(p, p.Pos(), "an unambiguous hash, i.e. one followed by whitespace"),
			}
		}

		if (p.Inline() && parser.TryAnyRune(p, whitespace.HorizontalRunes...) < 0) ||
			(!p.Inline() && parser.TryAnyRune(p, whitespace.Runes...) < 0) {
			return struct{}{}, &fancyerr.Error{
				Message: "ambiguous hash",
				Primary: quickanno.Expected(p, p.Pos(), "an unambiguous hash, i.e. one followed by whitespace"),
			}
		}
		return struct{}{}, nil
	}
}

func CharacterReference() parser.Func[*ast.CharacterReference] {
	return func(p *parser.Parser) (*ast.CharacterReference, *fancyerr.Error) {
		var r ast.CharacterReference

		r.Hash = parser.TryTokenAt(p, "#")
		if r.Hash == nil {
			return nil, &fancyerr.Error{
				Message: "missing character reference",
				Primary: quickanno.Expected(p, p.Pos(), "a character reference"),
				Examples: []fancyerr.Example{
					{Example: "`#amp;` or `#mdash;`"},
				},
			}
		}

		r.Name = parser.TokenWhile(p, func() bool {
			return parser.MatchesRunePredicate(p, codepoint.ASCIIAlphanumeric)
		})
		if !parser.TryRune(p, ';') {
			return nil, &fancyerr.Error{
				Message: "character reference: missing semicolon",
				Primary: quickanno.Expected(p, p.Pos(), "a semicolon to terminate the character reference"),
				Hints: []fancyerr.Hint{
					{Hint: "If you just wanted to write hash, you need to escape it.", Example: "`##`"},
				},
			}
		}
		if r.Name == "" {
			p.CaptureError(&fancyerr.Error{
				Message:  "character reference: missing name",
				Primary:  quickanno.Expected(p, p.Pos(), "a character reference name"),
				Examples: []fancyerr.Example{{Example: "`#amp;` or `#mdash;`"}},
				Hints: []fancyerr.Hint{
					{Hint: "If you just wanted to write hash, you need to escape it.", Example: "`##`"},
				},
			})
			return &r, nil
		}

		r.Chars = charref.Chars(r.Name)
		if r.Chars == "" {
			p.CaptureError(&fancyerr.Error{
				Message:  "character reference: unknown name",
				Primary:  quickanno.Expected(p, p.Pos(), "a valid character reference name"),
				Examples: []fancyerr.Example{{Example: "`#amp;` or `#mdash;`"}},
				Hints: []fancyerr.Hint{
					{Hint: "If you just wanted to write hash, you need to escape it.", Example: "`##`"},
				},
			})
		}

		return &r, nil
	}
}

var elementHeader parser.Func[*ast.ElementHeader]

func SetElementHeader(f parser.Func[*ast.ElementHeader]) {
	elementHeader = f
}
func ElementInterpolation() parser.Func[*ast.ElementInterpolation] {
	return func(p *parser.Parser) (*ast.ElementInterpolation, *fancyerr.Error) {
		var ei ast.ElementInterpolation

		ei.Hash = parser.TryRuneAt(p, '#')
		if ei.Hash == nil {
			return nil, &fancyerr.Error{
				Message: "missing element interpolation",
				Primary: quickanno.Expected(p, p.Pos(), "an element interpolation"),
			}
		}

		var header *ast.ElementHeader
		p.DoInline(func() { header = parser.Try(p, elementHeader) })
		if header == nil {
			return nil, &fancyerr.Error{
				Message: "missing element interpolation: missing element name",
				Primary: quickanno.Expected(p, p.Pos(), "an element name"),
				Examples: []fancyerr.Example{
					{Example: "`#br` or `#strong[foo]`"},
				},
			}
		}

		ei.Element = &ast.Element{Header: header}
		p.DoInline(func() {
			bt := parser.Try(p, body.BracketText())
			if bt != nil { // can't assign directly: any(nil) != (*ast.BracketText)(nil)
				ei.Element.Body = bt
			}
		})
		return &ei, nil
	}
}

var componentCallHeader parser.Func[*ast.ComponentCallHeader]

func SetComponentCallHeader(f parser.Func[*ast.ComponentCallHeader]) {
	componentCallHeader = f
}

func ComponentCallInterpolation() parser.Func[*ast.ComponentCallInterpolation] {
	return func(p *parser.Parser) (*ast.ComponentCallInterpolation, *fancyerr.Error) {
		var cci ast.ComponentCallInterpolation

		cci.Hash = parser.TryTokenAt(p, "#:")
		if cci.Hash == nil {
			return nil, &fancyerr.Error{
				Message: "missing component call interpolation",
				Primary: quickanno.Expected(p, p.Pos(), "a component call interpolation"),
			}
		}
		cci.ComponentCall = new(ast.ComponentCall)
		cci.ComponentCall.Colon = new(ast.Position)
		*cci.ComponentCall.Colon = *cci.Hash
		cci.ComponentCall.Colon.Col++

		cci.ComponentCall.Header = parser.Try(p, componentCallHeader)
		if cci.ComponentCall.Header == nil {
			return nil, &fancyerr.Error{
				Message: "missing component call interpolation: missing header",
				Primary: quickanno.Expected(p, p.Pos(), "a component call header"),
			}
		}

		p.DoInline(func() {
			bt := parser.Try(p, body.BracketText())
			if bt != nil {
				cci.ComponentCall.Body = &ast.UnderscoreBlockShorthand{
					Implicit: true,
					Body:     bt,
					Position: bt.LBracket,
				}
			}
		})

		return &cci, nil
	}
}

var expression parser.Func[*ast.Expression]

func SetExpression(f parser.Func[*ast.Expression]) {
	expression = f
}

func ExpressionInterpolation() parser.Func[*ast.ExpressionInterpolation] {
	return func(p *parser.Parser) (*ast.ExpressionInterpolation, *fancyerr.Error) {
		var ei ast.ExpressionInterpolation

		ei.Hash = parser.TryRuneAt(p, '#')
		if ei.Hash == nil {
			return nil, &fancyerr.Error{
				Message: "missing expression interpolation",
				Primary: quickanno.Expected(p, p.Pos(), "an expression interpolation"),
			}
		}

		ei.FormatDirective = parser.Try(p, formatDirective())

		ei.LBrace = parser.TryRuneAt(p, '{')
		if ei.LBrace == nil {
			if ei.FormatDirective != "" {
				p.CaptureError(&fancyerr.Error{
					Message: "expression interpolation: missing opening brace",
					Primary: quickanno.Expected(p, *ei.Hash, "an opening brace `{`"),
					Examples: []fancyerr.Example{
						{Example: "`#%" + ei.FormatDirective + "{...}`"},
					},
				})
				return &ei, nil
			}
			return nil, &fancyerr.Error{
				Message: "expression interpolation: missing opening brace",
				Primary: quickanno.Expected(p, *ei.Hash, "an opening brace `{`"),
				Examples: []fancyerr.Example{
					{Example: "`#{...}`"},
				},
			}
		}

		parser.TrySkip(p, whitespace.Horizontal())
		ei.Expression = parser.Must(p, expression)
		parser.TrySkip(p, whitespace.Horizontal())

		ei.RBrace = parser.TryRuneAt(p, '}')
		if ei.RBrace == nil {
			p.CaptureError(&fancyerr.Error{
				Message: "expression interpolation: missing closing brace",
				Primary: quickanno.Expected(p, *ei.Hash, "a closing brace `}`"),
			})
			return &ei, nil
		}

		return &ei, nil
	}
}

func formatDirective() parser.Func[string] {
	return func(p *parser.Parser) (string, *fancyerr.Error) {
		if !parser.TryRune(p, '%') {
			return "", &fancyerr.Error{
				Message: "missing format directive",
				Primary: quickanno.Expected(p, p.Pos(), "a format directive"),
				Examples: []fancyerr.Example{
					{Example: "`%d`"},
				},
			}
		}

		startIndex := p.Index()

		// flag
		for parser.TryAnyRune(p, '+', '-', '#', ' ', '0') > 0 {
		}

		// width
		if parser.TryRunePredicate(p, isInRange('1', '9')) > 0 {
			for parser.TryRunePredicate(p, isInRange('0', '9')) > 0 {
			}
		}

		// precision
		if parser.TryRune(p, '.') {
			for parser.TryRunePredicate(p, isInRange('0', '9')) > 0 {
			}
		}

		// verb
		if parser.TryAnyRune(p, 'v', 'T', 't', 'b', 'c', 'd', 'o', 'O', 'x', 'X', 'U', 'e', 'E', 'f', 'F', 'g', 'G', 's', 'p') < 0 {
			if parser.TryRunePredicate(p, isInRange('a', 'z')) > 0 || parser.TryRunePredicate(p, isInRange('A', 'Z')) > 0 {
				return p.Raw[startIndex:p.Index()], &fancyerr.Error{
					Message: "invalid format verb",
					Primary: quickanno.Expected(p, p.Pos(), "a valid format verb"),
					Examples: []fancyerr.Example{
						{Example: "`%d`"},
					},
					Explanation: "This is not a format verb according to the Go documentation. " +
						"Consult the documentation of the Go built-in package `fmt` for a list of valid format verbs.",
				}
			}

			return p.Raw[startIndex:p.Index()], &fancyerr.Error{
				Message: "missing format verb",
				Primary: quickanno.Expected(p, p.Pos(), "a format verb"),
				Examples: []fancyerr.Example{
					{Example: "`%d`"},
				},
			}
		}

		return p.Raw[startIndex:p.Index()], nil
	}
}

func isInRange(s, e rune) func(rune) bool {
	return func(r rune) bool { return r >= s && r <= e }
}
