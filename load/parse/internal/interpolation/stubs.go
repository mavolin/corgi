//go:build test_stubs

package interpolation

import (
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/golang"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

func init() {
	SetComponentCallHeader(componentCallHeaderStub)
	SetElementHeader(elementHeaderStub)
	SetExpression(expressionStub)
}

func elementHeaderStub(p *parser.Parser) (*ast.ElementHeader, *diagnostic.Diagnostic) {
	pos := p.Pos()
	name := parser.TokenWhile(p, func() bool {
		return parser.MatchesRunePredicate(p, isInRange('a', 'z'))
	})
	if name == "" {
		return nil, &diagnostic.Diagnostic{
			Message: "missing element name",
			Primary: quickanno.Expected(p, pos, "an element name"),
		}
	}
	return &ast.ElementHeader{
		Name: &ast.ElementReference{
			Name: &ast.ElementName{
				Name:     name,
				Position: &pos,
			},
		},
	}, nil
}

func componentCallHeaderStub(p *parser.Parser) (*ast.ComponentCallHeader, *diagnostic.Diagnostic) {
	name, err := parser.TryErr(p, golang.Identifier())
	if err != nil {
		return nil, &diagnostic.Diagnostic{
			Message: "missing component name",
			Primary: quickanno.Expected(p, p.Pos(), "a component name"),
		}
	}
	h := &ast.ComponentCallHeader{Name: name}

	h.Arguments = &ast.Arguments{
		LParen: &ast.Position{Line: 1, Col: p.Col()},
	}
	if !parser.TryRune(p, '(') {
		return nil, &diagnostic.Diagnostic{
			Message: "missing opening parenthesis",
			Primary: quickanno.Expected(p, p.Pos(), "opening parenthesis"),
		}
	}
	h.Arguments.RParen = p.PosPtr()
	if !parser.TryRune(p, ')') {
		return nil, &diagnostic.Diagnostic{
			Message: "missing closing parenthesis",
			Primary: quickanno.Expected(p, p.Pos(), "closing parenthesis"),
		}
	}

	return h, nil
}

func expressionStub(p *parser.Parser) (*ast.Expression, *diagnostic.Diagnostic) {
	pos := p.Pos()
	code := parser.TokenWhile(p, func() bool {
		return !parser.MatchesAnyRune(p, '}')
	})
	if code == "" {
		return nil, &diagnostic.Diagnostic{
			Message: "missing expression",
			Primary: quickanno.Expected(p, pos, "an expression"),
		}
	}
	return &ast.Expression{
		Code: ast.Code{
			&ast.GoCode{
				Code:     code,
				Position: &pos,
			},
		},
	}, nil
}
