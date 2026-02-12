package golang

import (
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

func Identifier() parser.Func[*ast.Identifier] { // https://go.dev/ref/spec#Identifiers
	return func(p *parser.Parser) *ast.Identifier {
		start := p.RuneIndex()

		name := parser.TokenWhileRunePredicate(p, func(r rune) bool {
			return Letter(r) || (p.RuneIndex() != start && Unicode_Digit(r))
		})
		if name == "" {
			return nil
		}

		var ident ast.Identifier
		ident.Name = name
		ident.Position = p.PosPtr()
		// computing the position instead of using p.Pos() at the top saves us
		// allocations when Identifier doesn't match
		ident.Position.Col -= ast.Col(p.RuneIndex() - start)

		if IsKeyword(ident.Name) {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "keyword used as identifier",
				Primary: quickanno.Expected(p, *ident.Position, "an identifier"),
				Explanation: "Go and Corgi reserve certain words as keywords, e.g. `if` or `comp`. " +
					"Because of their special meaning, you can't use them as identifiers. " +
					ident.Name + "` is one of those keywords.",
				Hints: []diagnostic.Hint{{Hint: "Use a different identifier."}},
			})
		}

		return &ident
	}
}
