package body

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/whitespace"
)

//nolint:gochecknoinits
func init() {
	if testing.Testing() {
		SetTextLine(textLineStub)
		// doesn't make the VerbatimBracketText very effective, but whatever
		SetVerbatimTextLine(textLineStub)
		SetScopeNode(scopeNodeStub)
	}
}

func textLineStub(term rune) parser.Func[ast.TextLine] {
	return func(p *parser.Parser) ast.TextLine {
		pos := p.Pos()
		line := parser.TokenWhile(p, func() bool {
			return !parser.Matches(p, textLineEnd(term))
		})
		if line == "" {
			return nil
		}
		return ast.TextLine{
			&ast.Text{
				Text:     line,
				Position: &pos,
			},
		}
	}
}

func textLineEnd(term rune) parser.Func[bool] {
	return func(p *parser.Parser) bool {
		parser.TrySkip(p, whitespace.Horizontal())
		return parser.MatchesAnyRune(p, term, '\n')
	}
}

func scopeNodeStub(p *parser.Parser) ast.ScopeNode {
	pos := p.Pos()
	name := parser.TokenWhile(p, func() bool {
		return parser.MatchesRunePredicate(p, func(r rune) bool {
			return r >= 'a' && r <= 'z'
		})
	})
	if name == "" {
		return nil
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
	}
}
