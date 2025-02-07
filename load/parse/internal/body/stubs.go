//go:build test_stubs

package body

import (
	"github.com/mavolin/corgi/v2/fancyerr"
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/whitespace"
)

func init() {
	SetTextLine(textLineStub)
	SetScopeNode(scopeNodeStub)
}

func textLineStub(term rune) parser.Func[ast.TextLine] {
	return func(p *parser.Parser) (ast.TextLine, *fancyerr.Error) {
		pos := p.Pos()
		line := parser.TokenWhile(p, func() bool {
			return !parser.Matches(p, textLineEnd(term))
		})
		if line == "" {
			return nil, &fancyerr.Error{
				Message: "missing text line",
				Primary: quickanno.Expected(p, p.Pos(), "a text line"),
			}
		}
		return ast.TextLine{
			&ast.Text{
				Text:     line,
				Position: &pos,
			},
		}, nil
	}
}

func textLineEnd(term rune) parser.Func[struct{}] {
	return func(p *parser.Parser) (struct{}, *fancyerr.Error) {
		parser.TrySkip(p, whitespace.Horizontal())
		if parser.MatchesAnyRune(p, term, '\n') {
			return struct{}{}, nil
		}

		return struct{}{}, &fancyerr.Error{
			Message: "missing text line end",
		}
	}
}

func scopeNodeStub(p *parser.Parser) (ast.ScopeNode, *fancyerr.Error) {
	pos := p.Pos()
	name := parser.TokenWhile(p, func() bool {
		return parser.MatchesRunePredicate(p, func(r rune) bool {
			return r >= 'a' && r <= 'z'
		})
	})
	if name == "" {
		return nil, &fancyerr.Error{
			Message: "missing scope node",
			Primary: quickanno.Expected(p, p.Pos(), "a scope node"),
		}
	}
	return &ast.Element{
		Header: &ast.ElementHeader{
			Name: &ast.ElementName{
				Name:     name,
				Position: &pos,
			},
		},
	}, nil
}
