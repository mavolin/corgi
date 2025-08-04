//go:build test_stubs

package interpolation

import (
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/golang"
)

func init() {
	SetComponentCallHeader(componentCallHeaderStub)
	SetElementHeader(elementHeaderStub)
	SetExpression(expressionStub)
}

func elementHeaderStub(p *parser.Parser) *ast.ElementHeader {
	pos := p.Pos()
	name := parser.TokenWhile(p, func() bool {
		return parser.MatchesRunePredicate(p, isInRange('a', 'z'))
	})
	if name == "" {
		return nil
	}
	return &ast.ElementHeader{
		Name: &ast.ElementReference{
			Name: &ast.ElementName{
				Name:     name,
				Position: &pos,
			},
		},
	}
}

func componentCallHeaderStub(p *parser.Parser) *ast.ComponentCallHeader {
	name := parser.Try(p, golang.Identifier())
	if name == nil {
		return nil
	}
	h := &ast.ComponentCallHeader{Name: name}

	h.Arguments = &ast.Arguments{
		LParen: &ast.Position{Line: 1, Col: int(p.Col())},
	}
	if !parser.TryRune(p, '(') {
		return nil
	}
	h.Arguments.RParen = parser.TryRuneAt(p, ')')
	if h.Arguments.RParen == nil {
		return nil
	}

	return h
}

func expressionStub(p *parser.Parser) *ast.Expression {
	pos := p.Pos()
	code := parser.TokenWhile(p, func() bool {
		return !parser.MatchesAnyRune(p, '}')
	})
	if code == "" {
		return nil
	}
	return &ast.Expression{
		Nodes: ast.Code{
			&ast.GoCode{
				Code:     code,
				Position: &pos,
			},
		},
	}
}
