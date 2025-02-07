package argument

import (
	"github.com/mavolin/corgi/v2/fancyerr"
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/attribute"
	"github.com/mavolin/corgi/v2/load/parse/internal/list"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

func Argument() parser.Func[ast.Argument] {
	return func(p *parser.Parser) (ast.Argument, *fancyerr.Error) {
		if a := parser.Try(p, ComponentArgument()); a != nil {
			return a, nil
		} else if a := parser.Try(p, attribute.Attribute()); a != nil {
			return a, nil
		}

		return nil, &fancyerr.Error{
			Message: "missing argument",
			Primary: quickanno.Expected(p, p.Pos(), "an argument"),
			Examples: []fancyerr.Example{
				{Title: "component argument", Example: "`foo: 123`"},
				{Title: "attribute", Example: "`class=\"woof\"`"},
			},
		}
	}
}

func Arguments() parser.Func[*ast.Arguments] {
	return func(p *parser.Parser) (*ast.Arguments, *fancyerr.Error) {
		l := parser.Try(p, list.ParenList("arguments", Argument()))
		if l == nil {
			return nil, &fancyerr.Error{
				Message: "missing arguments",
				Primary: quickanno.Expected(p, p.Pos(), "a list of arguments"),
			}
		}

		return &ast.Arguments{
			LParen: l.Open,
			Args:   l.Elems,
			RParen: l.Close,
		}, nil
	}
}
