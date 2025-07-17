package code

import (
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

func Expression(o Options) parser.Func[*ast.Expression] {
	o &= ^Statements
	return func(p *parser.Parser) (*ast.Expression, *diagnostic.Diagnostic) {
		var e ast.Expression

		e.Nodes = parser.Try(p, Code(o))
		if e.Nodes == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "missing expression",
				Primary: quickanno.Expected(p, p.Pos(), "an expression"),
			}
		}

		return &e, nil
	}
}

func NonZCExpression(o Options) parser.Func[*ast.Expression] {
	o &= ^Statements
	return func(p *parser.Parser) (*ast.Expression, *diagnostic.Diagnostic) {
		var e ast.Expression

		e.Nodes = parser.Try(p, NonZCCode(o))
		if e.Nodes == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "missing expression",
				Primary: quickanno.Expected(p, p.Pos(), "an expression"),
			}
		}

		return &e, nil
	}
}
