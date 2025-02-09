package code

import (
	"github.com/mavolin/corgi/v2/fancyerr"
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

func Expression(o Options) parser.Func[*ast.Expression] {
	o &= ^Statements
	return func(p *parser.Parser) (*ast.Expression, *fancyerr.Error) {
		var e ast.Expression

		e.Code = parser.Try(p, Code(o))
		if e.Code == nil {
			return nil, &fancyerr.Error{
				Message: "missing expression",
				Primary: quickanno.Expected(p, p.Pos(), "an expression"),
			}
		}

		return &e, nil
	}
}

func NonZCExpression(o Options) parser.Func[*ast.Expression] {
	o &= ^Statements
	return func(p *parser.Parser) (*ast.Expression, *fancyerr.Error) {
		var e ast.Expression

		e.Code = parser.Try(p, NonZCCode(o))
		if e.Code == nil {
			return nil, &fancyerr.Error{
				Message: "missing expression",
				Primary: quickanno.Expected(p, p.Pos(), "an expression"),
			}
		}

		return &e, nil
	}
}
