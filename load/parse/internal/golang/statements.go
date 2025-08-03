package golang

import (
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
)

// https://go.dev/ref/spec#Statements

// ============================================================================
// Assignment statements
// ======================================================================================

func AssignOp() parser.Func[string] {
	return func(p *parser.Parser) string {
		prefix := parser.TryInOrder(p, AddOp(), MulOp())
		if !parser.TryRune(p, '=') {
			return ""
		}

		return prefix + "="
	}
}
