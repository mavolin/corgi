//go:build test_stubs

package code

import (
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/body"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/golang"
)

func init() { //nolint:gochecknoinits
	SetComponentCall(componentCallStub)
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
		LParen: &ast.Position{Line: 1, Col: int(p.Col())},
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
