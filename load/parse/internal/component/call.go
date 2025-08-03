package component

import (
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/argument"
	"github.com/mavolin/corgi/v2/load/parse/internal/body"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/golang"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

func Call() parser.Func[*ast.ComponentCall] {
	return call(false)
}

func call(must bool) parser.Func[*ast.ComponentCall] {
	return func(p *parser.Parser) *ast.ComponentCall {
		var c ast.ComponentCall

		c.Colon = parser.TryRuneAt(p, ':')
		if c.Colon == nil {
			if must {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "component call: missing colon",
					Primary: quickanno.Expected(p, p.Pos(), "a colon"),
				})
			} else {
				return nil
			}
		}
		if !p.Inline() {
			parser.TrySkip(p, comment.OrHorizontalWhitespace())
		}

		c.Header = parser.Try(p, CallHeader())
		if c.Header == nil {
			return nil
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		c.Body = parser.Try(p, CallBody())
		return &c
	}
}

func CallHeader() parser.Func[*ast.ComponentCallHeader] {
	return func(p *parser.Parser) *ast.ComponentCallHeader {
		var h ast.ComponentCallHeader

		h.Name = parser.Try(p, golang.FullIdent())
		if h.Name == nil {
			return nil
		}
		if !p.Inline() {
			parser.TrySkip(p, comment.OrHorizontalWhitespace())
		}
		h.TypeArguments = parser.TryOptional(p, golang.TypeArgs(), nil)
		if h.TypeArguments != nil && !p.Inline() {
			parser.TrySkip(p, comment.OrHorizontalWhitespace())
		}

		h.Arguments = parser.Try(p, argument.Arguments())
		if h.Name == nil && h.Arguments == nil {
			return nil
		}

		return &h
	}
}

func With() parser.Func[*ast.With] {
	return func(p *parser.Parser) *ast.With {
		var w ast.With

		w.With = parser.TryKeywordAt(p, "with", comment.OrAnyWhitespace())
		if w.With == nil {
			return nil
		}

		w.Identifier = parser.TryOptional(p, golang.Identifier(), comment.OrHorizontalWhitespace())
		pos := p.Pos()
		w.Body = parser.Try(p, body.Body())
		if w.Body == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "with: missing body",
				Primary: quickanno.Expected(p, pos, "a body for the `with`"),
				Examples: []diagnostic.Example{
					{Example: "with woof { ... }"},
				},
				Hints: []diagnostic.Hint{
					{Hint: "If you want to inhibit the block default, use an empty scope.", Example: "`with woof {}`"},
				},
			})
		}

		return &w
	}
}
