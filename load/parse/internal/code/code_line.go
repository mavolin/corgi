package code

import (
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
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
		minus := parser.TryKeywordAt(p, "-")
		if minus == nil {
			return nil
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		var e ast.ExplicitCodeLine
		e.Minus = minus

		e.Statement = parser.Try(p, Statement())
		if e.Statement == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "explicit code line: missing statement",
				Primary: quickanno.Expected(p, p.Pos(), "a statement"),
			})
		}
		return &e
	}
}
