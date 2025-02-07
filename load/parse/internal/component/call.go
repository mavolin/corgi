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
		var c ast.ComponentCall

		c.Colon = parser.TryRuneAt(p, ':')
		if c.Colon == nil {
			return nil, &fancyerr.Error{
				Message: "missing component call",
				Primary: quickanno.Expected(p, p.Pos(), "a colon"),
			}
		}

		if !p.Inline() {
			parser.TrySkip(p, comment.OrHorizontalWhitespace())
		}

		c.Header = parser.Try(p, CallHeader())
		if c.Header == nil {
			return nil, &fancyerr.Error{
				Message: "missing component call",
				Primary: quickanno.Expected(p, p.Pos(), "a component call header"),
			}
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		c.Body = parser.Try(p, body.Body())
		return &c, nil
	}
}

func CallHeader() parser.Func[*ast.ComponentCallHeader] {
	return func(p *parser.Parser) (*ast.ComponentCallHeader, *fancyerr.Error) {
		var h ast.ComponentCallHeader

		h.Name = parser.Try(p, golang.FullIdent())
		if !p.Inline() {
			parser.TrySkip(p, comment.OrHorizontalWhitespace())
		}

		h.TypeArguments = parser.TryOptional(p, golang.TypeArgs(), nil)
		if h.TypeArguments != nil && !p.Inline() {
			parser.TrySkip(p, comment.OrHorizontalWhitespace())
		}

		h.Arguments = parser.Try(p, argument.Arguments())
		if h.Name == nil && h.Arguments == nil {
			return nil, &fancyerr.Error{
				Message: "missing component call header",
				Primary: quickanno.Expected(p, p.Pos(), "a name of a component"),
			}
		}

		return &h, nil
	}
}

func With() parser.Func[*ast.With] {
	return func(p *parser.Parser) (*ast.With, *fancyerr.Error) {
		var w ast.With

		w.With = parser.TryKeywordAt(p, "with", comment.OrAnyWhitespace())
		if w.With == nil {
			return nil, &fancyerr.Error{
				Message: "missing with",
				Primary: quickanno.Expected(p, *w.With, "a `with` here"),
			}
		}

		w.Name = parser.Try(p, golang.Identifier())
		if w.Name == nil {
			p.CaptureError(&fancyerr.Error{
				Message: "with: missing block name",
				Primary: quickanno.Expected(p, *w.With, "a name of a block"),
			})
		}
		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		w.Body = parser.Must(p, body.Body())
		return &w, nil
	}
}
