package golang

import (
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
)

// https://go.dev/ref/spec#Packages

func PackageName() parser.Func[*ast.Identifier] { // https://go.dev/ref/spec#PackageName
	return func(p *parser.Parser) *ast.Identifier {
		ident := parser.Try(p, Identifier())
		if ident == nil {
			return nil
		}
		return ident
	}
}
