package component

import (
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/body"
)

func Body() parser.Func[ast.ComponentBody] {
	return func(p *parser.Parser) ast.ComponentBody {
		if s := parser.Try(p, body.Body()); s != nil {
			return s
		} else if e := parser.Try(p, Extend()); e != nil {
			return e
		}

		return nil
	}
}

func Extend() parser.Func[*ast.Extend] {
	return func(p *parser.Parser) *ast.Extend {
		cc := parser.Try(p, Call())
		if cc == nil {
			return nil
		}

		return &ast.Extend{ComponentCall: cc}
	}
}
