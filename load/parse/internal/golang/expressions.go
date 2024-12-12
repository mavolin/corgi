package golang

import (
	"github.com/mavolin/corgi/v2/fancyerr"
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

// https://go.dev/ref/spec#Expressions

// ============================================================================
// Qualified identifiers
// ======================================================================================

func QualifiedIdent() parser.Func[*ast.QualifiedIdent] { // https://go.dev/ref/spec#QualifiedIdent
	return func(p *parser.Parser) (*ast.QualifiedIdent, *fancyerr.Error) {
		ident := &ast.QualifiedIdent{}

		pkg, ok := parser.TryOk(p, PackageName())
		if !ok {
			return nil, &fancyerr.Error{
				Message:  "missing qualified identifier",
				Primary:  quickanno.Expected(p, p.Pos(), "an identifier"),
				Examples: []fancyerr.Example{{Example: "`woof.Bark`"}},
			}
		}
		ident.Package = *pkg

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		dot := p.Pos()
		if !parser.TryRune(p, '.') {
			p.CaptureError(&fancyerr.Error{
				Message: "qualified identifier: missing dot and name in package",
				Primary: quickanno.Expected(p, p.Pos(), "a dot"),
				Explanation: "A qualified identifier consists of a package name, " +
					"and the name of a symbol in that package separated by a dot. " +
					"You are missing the dot and the name of the symbol.",
				Examples: []fancyerr.Example{{Example: "`" + pkg.Ident + ".Woof`"}},
			})
			return ident, nil
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())

		ident.Name, ok = parser.TryOk(p, Identifier())
		if !ok {
			p.CaptureError(&fancyerr.Error{
				Message: "qualified identifier: missing name in package",
				Primary: quickanno.Expected(p, dot, "an identifier"),
			})
		}

		return ident, nil
	}
}
