package whitespace

import (
	"github.com/mavolin/corgi/v2/file/diagnostic"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

// EOL matches the EOL or EOF with optionally preceding horizontal whitespace
func EOL() parser.WhitespaceFunc {
	return func(p *parser.Parser) *diagnostic.Diagnostic {
		if parser.TrySkip(p, EOF()) {
			return nil
		}

		parser.TrySkip(p, Horizontal())
		if parser.TrySkip(p, SingleVertical()) {
			return nil
		}
		return &diagnostic.Diagnostic{
			Message: "missing EOL",
			Primary: quickanno.Expected(p, p.Pos(), "a line ending"),
		}
	}
}

// EOF matches the end of file.
func EOF() parser.WhitespaceFunc {
	return func(p *parser.Parser) *diagnostic.Diagnostic {
		parser.TrySkip(p, Horizontal())
		if parser.TryRune(p, parser.EOF) {
			return nil
		}

		return &diagnostic.Diagnostic{
			Message: "expected EOF",
			Primary: quickanno.Expected(p, p.Pos(), "the end of file"),
		}
	}
}
