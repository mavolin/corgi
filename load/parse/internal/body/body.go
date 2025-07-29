package body

import (
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

func Body() parser.Func[ast.Body] {
	return func(p *parser.Parser) (ast.Body, *diagnostic.Diagnostic) {
		if s := parser.Try(p, Scope()); s != nil {
			return s, nil
		} else if b := parser.Try(p, BracketText()); b != nil {
			return b, nil
		}

		return nil, &diagnostic.Diagnostic{
			Message: "missing body",
			Primary: quickanno.Expected(p, p.Pos(), "a body"),
			Examples: []diagnostic.Example{
				{Title: "scope", Example: "{ :fmt.Number(val: 21_000) }"},
				{Title: "bracket text", Example: "[ Hello, World! ]"},
			},
		}
	}
}
