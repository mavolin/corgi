package golang

import (
	"github.com/mavolin/corgi/v2/fancyerr"
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

// https://go.dev/ref/spec#Packages

func PackageName() parser.Func[*ast.Ident] { // https://go.dev/ref/spec#PackageName
	return func(p *parser.Parser) (*ast.Ident, *fancyerr.Error) {
		ident, ok := parser.TryOk(p, Identifier())
		if !ok {
			return nil, &fancyerr.Error{
				Message:  "missing package name",
				Primary:  quickanno.Expected(p, p.Pos(), "a package name"),
				Examples: []fancyerr.Example{{Example: "`woof`"}},
			}
		}
		return ident, nil
	}
}
