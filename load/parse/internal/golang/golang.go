package golang

import (
	"github.com/mavolin/corgi/fancyerr"
	parser "github.com/mavolin/corgi/load/parse/internal"
	"github.com/mavolin/corgi/load/parse/internal/quickanno"
)

func AnyIdentifierRune() parser.Func[rune] {
	return func(p *parser.Parser) (rune, *fancyerr.Error) {
		r, ok := parser.TryRunePredicate(p, func(r rune) bool {
			return r == '_' || r >= '0' && r <= '9' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z'
		})
		if !ok {
			return 0, &fancyerr.Error{
				Message: "missing identifier",
				Primary: quickanno.Expected(p, p.Pos(), "an identifier"),
			}
		}
		return r, nil
	}
}
