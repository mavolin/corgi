package argument

import (
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/attribute"
	"github.com/mavolin/corgi/v2/load/parse/internal/list"
)

func Argument() parser.Func[ast.Argument] {
	return func(p *parser.Parser) ast.Argument {
		if a := parser.Try(p, ComponentArgument()); a != nil {
			return a
		} else if a := parser.Try(p, attribute.Attribute()); a != nil {
			return a
		}
		return nil
	}
}

func Arguments() parser.Func[*ast.Arguments] {
	return func(p *parser.Parser) *ast.Arguments {
		l := parser.Try(p, list.ParenList("argument", "arguments", Argument()))
		if l == nil {
			return nil
		}
		return &ast.Arguments{
			LParen: l.Open,
			List:   l.Elems,
			RParen: l.Close,
		}
	}
}
