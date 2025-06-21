//go:build test_stubs

package body

import (
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/whitespace"
)

func init() {
	SetTextLine(textLineStub)
	// doesn't make the VerbatimBracketText very effective, but whatever
	SetVerbatimTextLine(textLineStub)
	SetScopeNode(scopeNodeStub)
}

func textLineStub(term rune) parser.Func[ast.TextLine] {
	return func(p *parser.Parser) (ast.TextLine, *diagnostic.Diagnostic) {
		pos := p.Pos()
		line := parser.TokenWhile(p, func() bool {
			return !parser.Matches(p, textLineEnd(term))
		})
		if line == "" {
			return nil, &diagnostic.Diagnostic{
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
	return func(p *parser.Parser) (struct{}, *diagnostic.Diagnostic) {
		parser.TrySkip(p, whitespace.Horizontal())
		if parser.MatchesAnyRune(p, term, '\n') {
			return struct{}{}, nil
		}

		return struct{}{}, &diagnostic.Diagnostic{
			Message: "missing text line end",
		}
	}
}

func scopeNodeStub(p *parser.Parser) (ast.ScopeNode, *diagnostic.Diagnostic) {
	pos := p.Pos()
	name := parser.TokenWhile(p, func() bool {
		return parser.MatchesRunePredicate(p, func(r rune) bool {
			return r >= 'a' && r <= 'z'
		})
	})
	if name == "" {
		return nil, &diagnostic.Diagnostic{
			Message: "missing scope node",
			Primary: quickanno.Expected(p, p.Pos(), "a scope node"),
		}
	}
	return &ast.Element{
		Header: &ast.ElementHeader{
			Name: &ast.ElementReference{
				Name: &ast.ElementName{
					Name:     name,
					Position: &pos,
				},
			},
		},
	}, nil
}
