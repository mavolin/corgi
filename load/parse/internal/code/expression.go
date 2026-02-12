package code

import (
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
)

// Expression parses a full expression, either a zero-coalescing expression or
// a simple expression.
func Expression() parser.Func[*ast.Expression] {
	return func(p *parser.Parser) *ast.Expression {
		if zc := parser.Try(p, ZeroCoalescing()); zc != nil {
			return &ast.Expression{Nodes: ast.Code{zc}}
		}
		return parser.Try(p, SimpleExpression())
	}
}

// SimpleExpression parses a simple expression, either a component call or an
// enhanced expression.
func SimpleExpression() parser.Func[*ast.Expression] {
	return func(p *parser.Parser) *ast.Expression {
		if cc := parser.Try(p, componentCall); cc != nil {
			return &ast.Expression{Nodes: ast.Code{cc}}
		} else if ee := parser.Try(p, EnhancedExpression()); ee != nil {
			return ee
		}

		return nil
	}
}

// ParenExpression parses a parenthesised enhanced expression, e.g.
// `(foo.bar())`.
func ParenExpression() parser.Func[*ast.Expression] {
	return func(p *parser.Parser) *ast.Expression {
		if !parser.MatchesRune(p, '(') {
			return nil
		}

		nodes := parseEnhanced(p, enhancedParenExpressionParser)
		if len(nodes) == 0 {
			return nil
		}
		return &ast.Expression{Nodes: nodes}
	}
}

func EnhancedExpression() parser.Func[*ast.Expression] {
	return func(p *parser.Parser) *ast.Expression {
		// expression can't start with a brace or bracket
		if parser.MatchesAnyRune(p, '{', '[') {
			return nil
		}

		nodes := parseEnhanced(p, enhancedExpressionParser)
		if len(nodes) == 0 {
			return nil
		}
		return &ast.Expression{Nodes: nodes}
	}
}

// Enhancement parses one of corgi's extensions to Go's expression syntax, that
// can appear inside an expression.
func Enhancement() parser.Func[ast.CodeNode] {
	return func(p *parser.Parser) ast.CodeNode {
		if bf := parser.Try(p, BlockFunction()); bf != nil {
			return bf
		} else if s := parser.Try(p, String()); s != nil {
			return s
		} else if t := parser.Try(p, Ternary()); t != nil {
			return t
		}

		return nil
	}
}
