package golang

import (
	"github.com/mavolin/corgi/fancyerr"
	"github.com/mavolin/corgi/file/ast"
	parser "github.com/mavolin/corgi/load/parse/internal"
	"github.com/mavolin/corgi/load/parse/internal/quickanno"
)

func Identifier() parser.Func[*ast.Ident] { // https://go.dev/ref/spec#Identifiers
	return func(p *parser.Parser) (*ast.Ident, *fancyerr.Error) {
		ident := &ast.Ident{Position: p.Pos()}
		r, ok := parser.TryRunePredicate(p, Letter)
		if !ok {
			return nil, &fancyerr.Error{
				Message: "missing identifier",
				Primary: quickanno.Expected(p, ident.Position, "an identifier"),
			}
		}

		ident.Ident = string(r)
		if s, ok := parser.Try(p, IdentTrail()); ok {
			ident.Ident += s
		}

		return ident, nil
	}
}

// IdentTrail consumes the rest of the identifier after the first rune.
// It always matches, as the single, previously captured, rune is already a
// valid identifier.
func IdentTrail() parser.Func[string] {
	return func(p *parser.Parser) (string, *fancyerr.Error) {
		s := parser.TokenWhile(p, func() bool {
			return parser.MatchesAnyRunePredicate(p, func(r rune) bool {
				return Letter(r) || Unicode_Digit(r)
			})
		})
		return s, nil
	}
}
