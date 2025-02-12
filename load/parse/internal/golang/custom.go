package golang

import (
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

func FullIdent() parser.Func[ast.FullIdent] {
	return func(p *parser.Parser) (ast.FullIdent, *diagnostic.Diagnostic) {
		ident1 := parser.Try(p, Identifier())
		if ident1 == nil {
			return nil, &diagnostic.Diagnostic{
				Message:  "missing identifier",
				Primary:  quickanno.Expected(p, p.Pos(), "an identifier"),
				Examples: []diagnostic.Example{{Example: "`bark` or `woof.Bark`"}},
			}
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		dot := parser.TryRuneAt(p, '.')
		if dot == nil {
			return ident1, nil
		}
		parser.TrySkip(p, comment.OrAnyWhitespace())

		ident2 := parser.Try(p, Identifier())
		if ident2 == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "qualified identifier: missing name in package",
				Primary: quickanno.Expected(p, *dot, "an identifier"),
			})
			return ident1, nil
		}

		return &ast.QualifiedIdent{
			Package: ident1,
			Dot:     dot,
			Name:    ident2,
		}, nil
	}
}
