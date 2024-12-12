// Package interpolation implements generic interpolation parsers.
package interpolation

import (
	"github.com/mavolin/corgi/v2/escape/charref"
	"github.com/mavolin/corgi/v2/fancyerr"
	"github.com/mavolin/corgi/v2/fancyerr/anno"
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/html/codepoint"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/whitespace"

	_ "unsafe"
)

func TextInterpolation() parser.Func[ast.TextInterpolation] {
	return func(p *parser.Parser) (ast.TextInterpolation, *fancyerr.Error) {
		if err := ensureInterpolation(p); err != nil {
			return nil, err
		}

		if eh, ok := parser.TryOk(p, EscapedHash()); ok {
			return eh, nil
		} else if hs, ok := parser.TryOk(p, HashSpace()); ok {
			return hs, nil
		} else if ei, ok := parser.TryOk(p, ExpressionInterpolation()); ok {
			return ei, nil
		} else if cc, ok := parser.TryOk(p, ComponentCallInterpolation()); ok {
			return cc, nil
		} else if cr, ok := parser.TryOk(p, CharacterReference()); ok {
			return cr, nil
		} else if ei, ok := parser.TryOk(p, ElementInterpolation()); ok {
			return ei, nil
		}

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
				{Title: "hash space", Example: "`#_`"},
				{Title: "expression interpolation", Example: "`#{1 + 1}`"},
				{Title: "component call interpolation", Example: "`#:fmt.Number(val: 21_000)`"},
				{Title: "character reference", Example: "`#amp;`"},
				{Title: "element interpolation", Example: "`#br` or #strong[foo]"},
			},
		})
		return parser.Try(p, BadInterpolation())
	}
}

func StringInterpolation() parser.Func[ast.StringInterpolation] {
	return func(p *parser.Parser) (ast.StringInterpolation, *fancyerr.Error) {
		if err := ensureInterpolation(p); err != nil {
			return nil, err
		}

		if eh, ok := parser.TryOk(p, EscapedHash()); ok {
			return eh, nil
		} else if ei, ok := parser.TryOk(p, ExpressionInterpolation()); ok {
			return ei, nil
		} else if cc, ok := parser.TryOk(p, ComponentCallInterpolation()); ok {
			return cc, nil
		} else if cr, ok := parser.TryOk(p, CharacterReference()); ok {
			return cr, nil
		}

		// Try other kinds of interpolation, that aren't allowed inside a string
		pos := p.Pos()
		if hs, ok := parser.TryOk(p, HashSpace()); ok {
			p.CaptureError(&fancyerr.Error{
				Message: "cannot use hash space in string interpolation",
				Primary: []fancyerr.Annotation{
					anno.Range(p.File, hs.Pos(), hs.End(),
						"there is no point in using a hash space, you can just write a space instead"),
				},
				Explanation: "A hash space is used to insert a trailing space in text blocks. " +
					"This isn't necessary in strings, and you can just as well write a regular space instead.",
			})
			return &ast.BadInterpolation{Start: pos, Until: hs.Position}, nil
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
		return parser.Try(p, BadInterpolation())
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
		pos := p.Pos()
		if parser.TryRune(p, '#') {
			return &ast.BadInterpolation{Start: pos, Until: p.Pos()}, nil
		}

		return nil, &fancyerr.Error{
			Message: "missing bad interpolation",
		}
	}
}

func EscapedHash() parser.Func[*ast.EscapedHash] {
	return func(p *parser.Parser) (*ast.EscapedHash, *fancyerr.Error) {
		pos := p.Pos()
		if !parser.TryToken(p, "##") {
			return nil, &fancyerr.Error{
				Message: "missing escaped hash",
				Primary: quickanno.Expected(p, pos, "an escaped hash (`##`)"),
			}
		}

		return &ast.EscapedHash{Position: pos}, nil
	}
}

func HashSpace() parser.Func[*ast.HashSpace] {
	return func(p *parser.Parser) (*ast.HashSpace, *fancyerr.Error) {
		pos := p.Pos()
		if !parser.TryToken(p, "#_") {
			return nil, &fancyerr.Error{
				Message: "missing hash space",
				Primary: quickanno.Expected(p, pos, "a hash followed by an underline (`#_`)"),
			}
		}

		return &ast.HashSpace{Position: pos}, nil
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
		r := &ast.CharacterReference{Position: p.Pos()}

		if !parser.TryRune(p, '#') {
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
			}
		}
		if r.Name == "" {
			p.CaptureError(&fancyerr.Error{
				Message:  "character reference: missing name",
				Primary:  quickanno.Expected(p, p.Pos(), "a character reference name"),
				Examples: []fancyerr.Example{{Example: "`#amp;` or `#mdash;`"}},
			})
			return r, nil
		}

		r.Chars = charref.Chars(r.Name)
		if r.Chars == "" {
			p.CaptureError(&fancyerr.Error{
				Message:  "character reference: unknown name",
				Primary:  quickanno.Expected(p, p.Pos(), "a valid character reference name"),
				Examples: []fancyerr.Example{{Example: "`#amp;` or `#mdash;`"}},
			})
		}

		return r, nil
	}
}

//go:linkname ElementInterpolation
func ElementInterpolation() parser.Func[*ast.ElementInterpolation]

//go:linkname ComponentCallInterpolation
func ComponentCallInterpolation() parser.Func[*ast.ComponentCallInterpolation]

//go:linkname ExpressionInterpolation
func ExpressionInterpolation() parser.Func[*ast.ExpressionInterpolation]
