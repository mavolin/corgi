package golang

import (
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

// https://go.dev/ref/spec#Packages

func PackageName() parser.Func[*ast.Ident] { // https://go.dev/ref/spec#PackageName
	return func(p *parser.Parser) (*ast.Ident, *diagnostic.Diagnostic) {
		ident := parser.Try(p, Identifier())
		if ident == nil {
			return nil, &diagnostic.Diagnostic{
				Message:  "missing package name",
				Primary:  quickanno.Expected(p, p.Pos(), "a package name"),
				Examples: []diagnostic.Example{{Example: "`woof`"}},
			}
		}
		return ident, nil
	}
}
