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
	return func(p *parser.Parser) *ast.ComponentCall {
		colon := parser.TryRuneAt(p, ':')
		if colon == nil {
			return nil
		}
		if !p.Inline() {
			parser.TrySkip(p, comment.OrHorizontalWhitespace())
		}

		header := parser.Try(p, CallHeader())
		if header == nil {
			return nil
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		var c ast.ComponentCall
		c.Colon = colon
		c.Header = header

		c.Body = parser.Try(p, CallBody())
		return &c
	}
}

func CallHeader() parser.Func[*ast.ComponentCallHeader] {
	return func(p *parser.Parser) *ast.ComponentCallHeader {
		name := parser.Try(p, golang.FullIdent())
		if name == nil {
			return nil
		}
		if !p.Inline() {
			parser.TrySkip(p, comment.OrHorizontalWhitespace())
		}
		typeArguments := parser.TryOptional(p, golang.TypeArgs("component call"), nil)
		if typeArguments != nil && !p.Inline() {
			parser.TrySkip(p, comment.OrHorizontalWhitespace())
		}

		arguments := parser.Try(p, argument.Arguments("component call"))
		if name == nil && arguments == nil {
			return nil
		}

		return &ast.ComponentCallHeader{
			Name:          name,
			TypeArguments: typeArguments,
			Arguments:     arguments,
		}
	}
}

func With() parser.Func[*ast.With] {
	return func(p *parser.Parser) *ast.With {
		with := parser.TryKeywordAt(p, "with")
		if with == nil {
			return nil
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())

		var w ast.With
		w.With = with

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
