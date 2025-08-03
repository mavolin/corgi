package body

import (
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
)

func Body() parser.Func[ast.Body] {
	return func(p *parser.Parser) ast.Body {
		if s := parser.Try(p, Scope()); s != nil {
			return s
		} else if b := parser.Try(p, BracketText()); b != nil {
			return b
		}

		return nil
	}
}
