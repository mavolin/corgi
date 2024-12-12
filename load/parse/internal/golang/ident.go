package golang

import (
	"github.com/mavolin/corgi/v2/fancyerr"
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

func Identifier() parser.Func[*ast.Ident] { // https://go.dev/ref/spec#Identifiers
	return func(p *parser.Parser) (*ast.Ident, *fancyerr.Error) {
		ident := &ast.Ident{Position: p.Pos()}

		r := parser.TryRunePredicate(p, Letter)
		if r < 0 {
			return ident, &fancyerr.Error{
				Message:  "missing identifier",
				Primary:  quickanno.Expected(p, p.Pos(), "an identifier"),
				Examples: []fancyerr.Example{{Example: "`woof`"}},
			}
		}

		trail, _ := parser.Try(p, identTrail())
		ident.Ident = string(r) + trail
		if IsKeyword(ident.Ident) {
			p.CaptureError(&fancyerr.Error{
				Message: "keyword used as identifier",
				Primary: quickanno.Expected(p, p.Pos(), "an identifier"),
				Explanation: "Go reserves certain words as keywords with a special meaning, " +
					"for example `if` and `else`. Because of their special meaning, you can't " +
					"use them as identifiers. `" + ident.Ident + "` is one of those keywords.",
				Hints: []fancyerr.Hint{{Hint: "Use a different identifier."}},
			})
		}

		return ident, nil
	}
}

// identTrail consumes the rest of the identifier after the first rune.
// It always matches, as the single, previously captured, rune is already a
// valid identifier.
func identTrail() parser.Func[string] {
	return func(p *parser.Parser) (string, *fancyerr.Error) {
		s := parser.TokenWhile(p, func() bool {
			return parser.MatchesRunePredicate(p, func(r rune) bool {
				return Letter(r) || Unicode_Digit(r)
			})
		})
		return s, nil
	}
}
