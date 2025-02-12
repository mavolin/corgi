package golang

import (
	"github.com/mavolin/corgi/v2/file/diagnostic"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

// https://go.dev/ref/spec#Statements

// ============================================================================
// Assignment statements
// ======================================================================================

func AssignOp() parser.Func[string] {
	return func(p *parser.Parser) (string, *diagnostic.Diagnostic) {
		prefix := parser.TryInOrder(p, AddOp(), MulOp())
		if !parser.TryRune(p, '=') {
			return "", &diagnostic.Diagnostic{
				Message: "missing assign op",
				Primary: quickanno.Expected(p, p.Pos(), "an assignment operator, e.g. `=` or `+=`"),
			}
		}

		return prefix + "=", nil
	}
}
