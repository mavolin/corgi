package code

import (
	"github.com/mavolin/corgi/v2/fancyerr"
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

func Expression() parser.Func[*ast.Expression] {
	return func(p *parser.Parser) (*ast.Expression, *fancyerr.Error) {
		var e ast.Expression

		e.Code = parser.Try(p, Code(false))
		if e.Code == nil {
			return nil, &fancyerr.Error{
				Message: "missing expression",
				Primary: quickanno.Expected(p, p.Pos(), "an expression"),
			}
		}

		return &e, nil
	}
}

func NonZCExpression() parser.Func[*ast.Expression] {
	return func(p *parser.Parser) (*ast.Expression, *fancyerr.Error) {
		var e ast.Expression

		e.Code = parser.Try(p, NonZCCode(false))
		if e.Code == nil {
			return nil, &fancyerr.Error{
				Message: "missing expression",
				Primary: quickanno.Expected(p, p.Pos(), "an expression"),
			}
		}

		return &e, nil
	}
}
