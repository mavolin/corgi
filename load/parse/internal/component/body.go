package component

import (
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/body"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

func Body() parser.Func[ast.ComponentBody] {
	return func(p *parser.Parser) (ast.ComponentBody, *diagnostic.Diagnostic) {
		if s := parser.Try(p, body.Body()); s != nil {
			return s, nil
		} else if e := parser.Try(p, Extend()); e != nil {
			return e, nil
		}

		return nil, &diagnostic.Diagnostic{
			Message: "missing component body",
			Primary: quickanno.Expected(p, p.Pos(), "a body"),
			Examples: []diagnostic.Example{
				{Title: "scope", Example: "`{ :fmt.Number(val: 21_000) }`"},
				{Title: "bracket text", Example: "`[ Woof! ]`"},
				{Title: "extend", Example: "`:layout()`"},
			},
		}
	}
}

func Extend() parser.Func[*ast.Extend] {
	return func(p *parser.Parser) (*ast.Extend, *diagnostic.Diagnostic) {
		var e ast.Extend

		e.ComponentCall = parser.Try(p, Call())
		if e.ComponentCall == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "missing extend",
				Primary: quickanno.Expected(p, p.Pos(), "an component call"),
			}
		}

		return &e, nil
	}
}
