package golang

import (
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

func Identifier() parser.Func[*ast.Identifier] { // https://go.dev/ref/spec#Identifiers
	return func(p *parser.Parser) *ast.Identifier {
		start := p.Index()

		r := parser.TryRunePredicate(p, Letter)
		if r == 0 {
			return nil
		}

		var ident ast.Identifier
		ident.Position = p.PosPtr()
		// computing the position instead of using p.Pos() at the top saves us
		// allocations when Identifier doesn't match
		ident.Position.Col--

		parser.Try(p, identTrail())
		ident.Name = p.AST.Raw[start:p.Index()]
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

// identTrail consumes the rest of the identifier after the first rune.
// It always matches, as the single, previously captured, rune is already a
// valid identifier.
func identTrail() parser.Func[bool] {
	return func(p *parser.Parser) bool {
		s := parser.TokenWhile(p, func() bool {
			return parser.MatchesRunePredicate(p, func(r rune) bool {
				return Letter(r) || Unicode_Digit(r)
			})
		})
		return s != ""
	}
}
