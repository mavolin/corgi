package list

import (
	"slices"

	"github.com/mavolin/corgi/v2/fancyerr"
	"github.com/mavolin/corgi/v2/fancyerr/anno"
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/unexpected"
)

type List[T any] struct {
	Open  *ast.Position
	Elems []T
	Close *ast.Position
}

func ParenList[T any](name string, elemFunc parser.Func[T]) parser.Func[*List[T]] {
	return list(name, '(', ')', elemFunc)
}

func BracketList[T any](name string, elemFunc parser.Func[T]) parser.Func[*List[T]] {
	return list(name, '[', ']', elemFunc)
}

func list[T any](name string, open, close rune, elemFunc parser.Func[T]) parser.Func[*List[T]] {
	return func(p *parser.Parser) (*List[T], *fancyerr.Error) {
		var l List[T]

		l.Open = parser.TryRuneAt(p, open)
		if l.Open == nil {
			return nil, &fancyerr.Error{
				Message: "missing " + name,
				Primary: quickanno.Expected(p, p.Pos(), "a `"+string(open)+"`"),
			}
		}

		l.Elems = make([]T, 0, 64)
		for {
			parser.TrySkip(p, comment.OrAnyWhitespace())

			pos := p.Pos()
			if parser.TryOptionalRune(p, close, nil) {
				l.Close = &pos
				break
			} else if parser.TryOptionalRune(p, parser.EOF, nil) {
				p.CaptureError(&fancyerr.Error{
					Message: "unclosed " + name,
					Primary: quickanno.Expected(p, *l.Open, "a `"+string(close)+"`"),
					Secondary: []fancyerr.Annotation{
						anno.Position(p.File, *l.Open, "for the opening `"+string(open)+"` here"),
					},
				})
				break
			}
			elem, err := parser.TryOptionalErr(p, elemFunc, nil)
			l.Elems = append(l.Elems, elem)
			if err != nil {
				if parser.MatchesToken(p, ",") { // missing elem
					p.CaptureError(err)
				} else {
					err = unexpected.UntilAnyRune(p, comment.OrAnyWhitespace(), ',', close)
					err.Message = "missing " + name
					err.Primary[0].Annotation = "found these unexpected runes instead"
					p.CaptureError(err)
				}
			}

			parser.TrySkip(p, comment.OrHorizontalWhitespace())

			if parser.MatchesAnyRune(p, close, parser.EOF) || parser.TryRune(p, ',') {
				continue
			}

			// let's be hyper-forgiving and allow for a missing comma, if the
			// upcoming runes parse successfully
			parser.TrySkip(p, comment.OrHorizontalWhitespace())
			if parser.Matches(p, elemFunc) {
				p.CaptureError(&fancyerr.Error{
					Message: "missing comma",
					Primary: quickanno.Expected(p, p.Pos(), "a comma here"),
				})
				continue
			}

			// not (just) a missing comma, capture as unexpected tokens

			err = unexpected.UntilAnyRune(p, comment.OrHorizontalWhitespace(), ',', close)
			if err != nil {
				err.Message = "unexpected runes before `,` or `" + string(close) + "`"
				p.CaptureError(err)
			}

			if !parser.MatchesAnyRune(p, ',', close) {
				p.CaptureError(&fancyerr.Error{
					Message: "unclosed " + name,
					Primary: quickanno.Expected(p, *l.Open, "a `"+string(close)+"`"),
					Secondary: []fancyerr.Annotation{
						anno.Position(p.File, *l.Open, "for the opening `"+string(open)+"` here"),
					},
				})
				break
			}

			parser.TryRune(p, ',')
		}

		if len(l.Elems) == 0 {
			l.Elems = nil
		} else {
			l.Elems = slices.Clip(l.Elems)
		}
		return &l, nil
	}
}

func CommaList[T any](singular, plural string, elemFunc parser.Func[T]) parser.Func[[]T] {
	return func(p *parser.Parser) ([]T, *fancyerr.Error) {
		elems := make([]T, 1, 64)

		var err *fancyerr.Error
		elems[0], err = parser.TryErr(p, elemFunc)
		if err != nil {
			if parser.MatchesToken(p, ",") { // missing elem
				p.CaptureError(err)
			} else {
				return nil, &fancyerr.Error{
					Message: "missing " + singular,
					Primary: quickanno.Expected(p, p.Pos(), "one or more "+plural),
				}
			}
		}

		for {
			parser.TrySkip(p, comment.OrHorizontalWhitespace())

			commaPos := p.Pos()
			if !parser.TryRune(p, ',') {
				if len(elems) == 0 {
					return nil, nil
				}
				return slices.Clip(elems), nil
			}

			parser.TrySkip(p, comment.OrAnyWhitespace())

			elem, err := parser.TryErr(p, elemFunc)
			if err != nil {
				if parser.MatchesToken(p, ",") { // missing elem
					p.CaptureError(err)
				} else {
					p.CaptureError(&fancyerr.Error{
						Message: "missing " + singular,
						Primary: quickanno.Expected(p, commaPos, "found a comma here, but no "+singular+" after it"),
					})
				}
			}
			elems = append(elems, elem)
		}
	}
}
