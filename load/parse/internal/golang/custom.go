package golang

import (
	"github.com/mavolin/corgi/v2/fancyerr"
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

func FullIdent() parser.Func[ast.FullIdent] {
	return func(p *parser.Parser) (ast.FullIdent, *fancyerr.Error) {
		ident1, ok := parser.TryOk(p, Identifier())
		if !ok {
			return nil, &fancyerr.Error{
				Message:  "missing identifier",
				Primary:  quickanno.Expected(p, p.Pos(), "an identifier"),
				Examples: []fancyerr.Example{{Example: "`bark` or `woof.Bark`"}},
			}
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		dot := p.Pos()
		if !parser.TryRune(p, '.') {
			return ident1, nil
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())

		ident2, ok := parser.TryOk(p, Identifier())
		if !ok {
			p.CaptureError(&fancyerr.Error{
				Message: "qualified identifier: missing name in package",
				Primary: quickanno.Expected(p, dot, "an identifier"),
			})
			return ident1, nil
		}

		return &ast.QualifiedIdent{
			Package: *ident1,
			Dot:     &dot,
			Name:    ident2,
		}, nil
	}
}
