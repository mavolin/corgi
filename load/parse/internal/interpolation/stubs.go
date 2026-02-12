package interpolation

import (
	"testing"

	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/body"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/golang"
)

func init() { //nolint:gochecknoinits
	if testing.Testing() {
		SetComponentCall(componentCallStub)
		SetExpression(expressionStub)
	}
}

func componentCallStub(p *parser.Parser) *ast.ComponentCall {
	colon := parser.TryRuneAt(p, ':')
	if colon == nil {
		return nil
	}

	var cc ast.ComponentCall
	cc.Colon = colon

	cc.Header = new(ast.ComponentCallHeader)
	cc.Header.Name = parser.Try(p, golang.Identifier())
	if cc.Header.Name == nil {
		return nil
	}

	cc.Header.Arguments = &ast.Arguments{
		LParen: &ast.Position{Line: 1, Col: p.Col()},
	}
	if !parser.TryRune(p, '(') {
		return nil
	}
	cc.Header.Arguments.RParen = parser.TryRuneAt(p, ')')
	if cc.Header.Arguments.RParen == nil {
		return nil
	}

	parser.TrySkip(p, comment.OrHorizontalWhitespace())
	cc.Body = parser.Try(p, body.Scope())
	if cc.Body == (*ast.Scope)(nil) { // typed nil
		cc.Body = nil
	}

	return &cc
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
