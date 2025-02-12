package code

import (
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/whitespace"
)

func ImplicitCodeLine() parser.Func[*ast.ImplicitCodeLine] {
	return func(p *parser.Parser) (*ast.ImplicitCodeLine, *diagnostic.Diagnostic) {
		s := parser.Try(p, ParsedStatement())
		if s == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "missing implicit code line",
				Primary: quickanno.Expected(p, p.Pos(), "an implicit code line"),
			}
		}

		return &ast.ImplicitCodeLine{Statement: s}, nil
	}
}

func ExplicitCodeLine() parser.Func[*ast.ExplicitCodeLine] {
	return func(p *parser.Parser) (*ast.ExplicitCodeLine, *diagnostic.Diagnostic) {
		var e ast.ExplicitCodeLine

		e.Minus = parser.TryKeywordAt(p, "-", whitespace.Horizontal())
		if e.Minus == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "missing explicit code line",
				Primary: quickanno.Expected(p, p.Pos(), "an explicit code line"),
			}
		}

		e.Statement = parser.Must(p, Statement(Regular))
		return &e, nil
	}
}
