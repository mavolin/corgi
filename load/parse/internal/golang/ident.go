package golang

import (
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

func Identifier() parser.Func[*ast.Identifier] { // https://go.dev/ref/spec#Identifiers
	return func(p *parser.Parser) (*ast.Identifier, *diagnostic.Diagnostic) {
		var ident ast.Identifier
		ident.Position = p.PosPtr()

		r := parser.TryRunePredicate(p, Letter)
		if r < 0 {
			return nil, &diagnostic.Diagnostic{
				Message:  "missing identifier",
				Primary:  quickanno.Expected(p, p.Pos(), "an identifier"),
				Examples: []diagnostic.Example{{Example: "`woof`"}},
			}
		}

		trail := parser.Try(p, identTrail())
		ident.Name = string(r) + trail
		if IsKeyword(ident.Name) {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "keyword used as identifier",
				Primary: quickanno.Expected(p, p.Pos(), "an identifier"),
				Explanation: "Go reserves certain words as keywords with a special meaning, " +
					"for example `if` and `else`. Because of their special meaning, you can't " +
					"use them as identifiers. `" + ident.Name + "` is one of those keywords.",
				Hints: []diagnostic.Hint{{Hint: "Use a different identifier."}},
			})
		}

		return &ident, nil
	}
}

// identTrail consumes the rest of the identifier after the first rune.
// It always matches, as the single, previously captured, rune is already a
// valid identifier.
func identTrail() parser.Func[string] {
	return func(p *parser.Parser) (string, *diagnostic.Diagnostic) {
		s := parser.TokenWhile(p, func() bool {
			return parser.MatchesRunePredicate(p, func(r rune) bool {
				return Letter(r) || Unicode_Digit(r)
			})
		})
		return s, nil
	}
}
