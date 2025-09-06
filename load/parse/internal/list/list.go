package list

import (
	"slices"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/unexpected"
	"github.com/mavolin/corgi/v2/load/parse/internal/whitespace"
)

type List[T any] struct {
	Open  *ast.Position
	Elems []T
	Close *ast.Position
}

// ParenList parses a list of zero or more elements enclosed in parentheses
// and separated by commas.
//
// Missing elements are reported and represented by the zero value of T.
func ParenList[T comparable](singular, plural string, elemFunc parser.Func[T]) parser.Func[*List[T]] {
	return list(singular, plural, '(', ')', elemFunc)
}

// BracketList parses a list of zero or more elements enclosed in brackets
// and separated by commas.
//
// Missing elements are reported and represented by the zero value of T.
func BracketList[T comparable](singular, plural string, elemFunc parser.Func[T]) parser.Func[*List[T]] {
	return list(singular, plural, '[', ']', elemFunc)
}

func list[T comparable](singular, plural string, opening, closing rune, elemFunc parser.Func[T]) parser.Func[*List[T]] {
	return func(p *parser.Parser) *List[T] {
		open := parser.TryRuneAt(p, opening)
		if open == nil {
			return nil
		}

		var l List[T]
		l.Open = open

		var zero T

		for {
			lastPos := p.Pos()
			parser.TrySkip(p, comment.OrAnyWhitespace())

			pos := p.Pos()
			if parser.TryOptionalRune(p, closing, nil) {
				l.Close = &pos
				break
			} else if parser.TryOptionalRune(p, parser.EOF, nil) ||
				p.Inline() && parser.MatchesAnyRune(p, whitespace.VerticalRunes...) {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "unclosed " + plural,
					Primary: quickanno.Expected(p, lastPos, "a `"+string(closing)+"`"),
					Secondary: []diagnostic.Annotation{
						anno.Position(p.File, *l.Open, "for the opening `"+string(opening)+"` here"),
					},
				})
				break
			}
			elem := parser.TryOptional(p, elemFunc, nil)
			l.Elems = append(l.Elems, elem)
			if elem == zero {
				if parser.MatchesToken(p, ",") { // missing elem
					p.CaptureError(&diagnostic.Diagnostic{
						Message: "missing " + singular,
						Primary: quickanno.Expected(p, pos, plural),
					})
				} else {
					err := unexpected.UntilAnyRune(p, comment.OrAnyWhitespace(), ',', closing)
					if err != nil {
						err.Message = "missing " + plural
						err.Primary[0].Annotation = "found these unexpected runes instead"
						p.CaptureError(err)
					}
				}
			}

			parser.TrySkip(p, comment.OrHorizontalWhitespace())

			if parser.MatchesAnyRune(p, closing, parser.EOF) || parser.TryRune(p, ',') {
				continue
			} else if p.Inline() && parser.MatchesAnyRune(p, whitespace.VerticalRunes...) {
				continue
			}

			// let's be hyper-forgiving and allow for a missing comma, if the
			// upcoming runes parse successfully
			parser.TrySkip(p, comment.OrHorizontalWhitespace())
			if parser.Matches(p, elemFunc) {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "missing comma",
					Primary: quickanno.Expected(p, p.Pos(), "a comma here"),
				})
				continue
			}

			// not (just) a missing comma, capture as unexpected tokens

			err := unexpected.UntilAnyRune(p, comment.OrHorizontalWhitespace(), ',', closing)
			if err != nil {
				err.Message = "unexpected runes before `,` or `" + string(closing) + "`"
				p.CaptureError(err)
			}

			if !parser.MatchesAnyRune(p, ',', closing) {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "unclosed " + plural,
					Primary: quickanno.Expected(p, p.Pos(), "a `"+string(closing)+"`"),
					Secondary: []diagnostic.Annotation{
						anno.Position(p.File, *l.Open, "for the opening `"+string(opening)+"` here"),
					},
				})
				break
			}

			parser.TryRune(p, ',')
		}

		l.Elems = slices.Clip(l.Elems)
		return &l
	}
}

// CommaList parses a list of one or more elements separated by commas.
//
// Missing elements are reported and represented by the zero value of T.
func CommaList[T comparable](singular, plural string, elemFunc parser.Func[T]) parser.Func[[]T] {
	return func(p *parser.Parser) []T {
		var zero T

		elem0 := parser.Try(p, elemFunc)
		if elem0 == zero {
			if parser.MatchesToken(p, ",") { // missing elem
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "missing " + singular,
					Primary: quickanno.Expected(p, p.Pos(), "one or more "+plural),
				})
			} else {
				return nil
			}
		}
		elems := []T{elem0}

		for {
			parser.TrySkip(p, comment.OrHorizontalWhitespace())

			commaPos := p.Pos()
			if !parser.TryRune(p, ',') {
				if len(elems) == 0 {
					return nil
				}
				return slices.Clip(elems)
			}

			parser.TrySkip(p, comment.OrAnyWhitespace())

			elem := parser.Try(p, elemFunc)
			elems = append(elems, elem)
			if elem != zero {
				continue
			}
			if parser.MatchesToken(p, ",") { // missing elem
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "missing " + singular,
					Primary: quickanno.Expected(p, commaPos, plural),
				})
			} else {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "missing " + singular,
					Primary: quickanno.Expected(p, commaPos, "found a comma here, but no "+singular+" after it"),
				})
			}
		}
	}
}
