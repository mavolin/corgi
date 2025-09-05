package code

import (
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
)

func Expression(o Options) parser.Func[*ast.Expression] {
	o &= ^Statements
	return func(p *parser.Parser) *ast.Expression {
		var e ast.Expression

		e.Nodes = parser.Try(p, Code(o))
		if e.Nodes == nil {
			return nil
		}

		return &e
	}
}

func NonZCExpression(o Options) parser.Func[*ast.Expression] {
	o &= ^Statements
	return func(p *parser.Parser) *ast.Expression {
		var e ast.Expression
		e.Nodes = parser.Try(p, GoCode(o))
		if e.Nodes == nil {
			return nil
		}

		return &e
	}
}
