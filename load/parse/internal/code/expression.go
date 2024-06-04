package code

import (
	"github.com/mavolin/corgi/fancyerr"
	"github.com/mavolin/corgi/file/ast"
	parser "github.com/mavolin/corgi/load/parse/internal"
)

func Expression() parser.Func[ast.Expression] {
	return func(p *parser.Parser) (ast.Expression, *fancyerr.Error) {
		panic("implement me")
		// todo: implement
	}
}

func ExpressionInterpolation() parser.Func[*ast.ExpressionInterpolation] {
	return func(p *parser.Parser) (*ast.ExpressionInterpolation, *fancyerr.Error) {
		panic("implement me")
		// todo: implement
	}
}
