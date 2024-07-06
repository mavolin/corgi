package list

import (
	"slices"

	"github.com/mavolin/corgi/fancyerr"
	"github.com/mavolin/corgi/fancyerr/anno"
	"github.com/mavolin/corgi/file/ast"
	parser "github.com/mavolin/corgi/load/parse/internal"
	"github.com/mavolin/corgi/load/parse/internal/comment"
	"github.com/mavolin/corgi/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/load/parse/internal/unexpected"
)

type List[T any] struct {
	Open  ast.Position
	Close *ast.Position
	Elems []T
}

func ParenList[T any](name string, elemFunc parser.Func[T]) parser.Func[*List[T]] {
	return list(name, '(', ')', elemFunc)
}

func BracketList[T any](name string, elemFunc parser.Func[T]) parser.Func[*List[T]] {
	return list(name, '[', ']', elemFunc)
}

func list[T any](name string, open, close rune, elemFunc parser.Func[T]) parser.Func[*List[T]] {
	return func(p *parser.Parser) (*List[T], *fancyerr.Error) {
		l := &List[T]{Open: p.Pos()}

		if !parser.TryRune(p, open) {
			return nil, &fancyerr.Error{
				Message: "missing " + name,
				Primary: quickanno.Expected(p, p.Pos(), "a `"+string(open)+"`"),
			}
		}

		l.Elems = make([]T, 0, 64)
		for {
			parser.Try(p, comment.OrAnyWhitespace())

			pos := p.Pos()
			if parser.TryRune(p, close) {
				l.Close = &pos
				break
			} else if parser.TryRune(p, parser.EOF) {
				p.CaptureError(&fancyerr.Error{
					Message: "unclosed " + name,
					Primary: quickanno.Expected(p, l.Open, "a `"+string(close)+"`"),
					Secondary: []fancyerr.Annotation{
						anno.Position(p.File, l.Open, "for the opening `"+string(open)+"` here"),
					},
				})
				break
			}

			elem, err := elemFunc(p)
			l.Elems = append(l.Elems, elem)
			if err != nil {
				if parser.MatchesToken(p, ",") { // missing elem
					p.CaptureError(err)
				} else {
					err = unexpected.UntilAnyRune(p, comment.OrAnyWhitespace(), ',', close)
					if err != nil {
						err.Message = "missing " + name + " element"
						err.Primary[0].Annotation = "found these unexpected runes instead"
						p.CaptureError(err)
					}
				}
			}

			parser.Try(p, comment.OrHorizontalWhitespace())

			if parser.TryRune(p, ',') {
				continue
			} else if parser.MatchesAnyRune(p, close, parser.EOF) {
				continue
			}

			// let's be hyper-forgiving and allow for a missing comma, if the
			// upcoming runes parse successfully
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
					Primary: quickanno.Expected(p, l.Open, "a `"+string(close)+"`"),
					Secondary: []fancyerr.Annotation{
						anno.Position(p.File, l.Open, "for the opening `"+string(open)+"` here"),
					},
				})
				break
			}

			parser.TryRune(p, ',')
		}

		l.Elems = slices.Clip(l.Elems)
		return l, nil
	}
}
