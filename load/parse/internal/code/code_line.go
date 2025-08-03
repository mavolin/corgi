package code

import (
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/whitespace"
)

func ImplicitCodeLine() parser.Func[*ast.ImplicitCodeLine] {
	return func(p *parser.Parser) *ast.ImplicitCodeLine {
		s := parser.Try(p, ParsedStatement())
		if s == nil {
			return nil
		}

		return &ast.ImplicitCodeLine{Statement: s}
	}
}

func ExplicitCodeLine() parser.Func[*ast.ExplicitCodeLine] {
	return func(p *parser.Parser) *ast.ExplicitCodeLine {
		var e ast.ExplicitCodeLine

		e.Minus = parser.TryKeywordAt(p, "-", whitespace.Horizontal())
		if e.Minus == nil {
			return nil
		}

		e.Statement = parser.Try(p, Statement(Regular))
		if e.Statement == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "explicit code line: missing statement",
				Primary: quickanno.Expected(p, p.Pos(), "a statement"),
			})
		}
		return &e
	}
}
