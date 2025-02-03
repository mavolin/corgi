package component

import (
	"github.com/mavolin/corgi/v2/fancyerr"
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/argument"
	"github.com/mavolin/corgi/v2/load/parse/internal/body"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/golang"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

func Call() parser.Func[*ast.ComponentCall] {
	return func(p *parser.Parser) (*ast.ComponentCall, *fancyerr.Error) {
		h, ok := parser.TryOk(p, CallHeader())
		if !ok {
			return nil, &fancyerr.Error{
				Message: "missing component call",
				Primary: quickanno.Expected(p, p.Pos(), "a component call"),
			}
		}

		c := &ast.ComponentCall{Header: *h}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		c.Body, _ = parser.Try(p, body.Body())
		return c, nil
	}
}

func CallHeader() parser.Func[*ast.ComponentCallHeader] {
	return func(p *parser.Parser) (*ast.ComponentCallHeader, *fancyerr.Error) {
		h := &ast.ComponentCallHeader{Colon: p.Pos()}
		if !parser.TryRune(p, ':') {
			return nil, &fancyerr.Error{
				Message: "missing component call header",
				Primary: quickanno.Expected(p, h.Colon, "a colon"),
			}
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())

		var ok bool
		h.Name, ok = parser.TryOk(p, golang.FullIdent())
		if !ok {
			return nil, &fancyerr.Error{
				Message: "component call header: missing name",
				Primary: quickanno.Expected(p, p.Pos(), "a name of a component"),
			}
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		h.TypeArguments, ok = parser.TryOptionalOk(p, golang.TypeArgs())
		if ok {
			parser.TrySkip(p, comment.OrHorizontalWhitespace())
		}

		h.Arguments, _ = parser.Try(p, argument.Arguments())
		return h, nil
	}
}

func With() parser.Func[*ast.With] {
	return func(p *parser.Parser) (*ast.With, *fancyerr.Error) {
		w := &ast.With{With: p.Pos()}
		if !parser.TryToken(p, "with") || !parser.TrySkipOk(p, comment.OrAnyWhitespace()) {
			return nil, &fancyerr.Error{
				Message: "missing with",
				Primary: quickanno.Expected(p, w.With, "a `with` here"),
			}
		}

		var ok bool
		w.Name, ok = parser.TryOk(p, golang.Identifier())
		if !ok {
			return nil, &fancyerr.Error{
				Message: "with: missing block name",
				Primary: quickanno.Expected(p, w.With, "a name of a block"),
			}
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		w.Body = parser.Must(p, body.Body())
		return w, nil
	}
}
