package whitespace

import (
	"github.com/mavolin/corgi/v2/fancyerr"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

// EOL matches the EOL or EOF with optionally preceding horizontal whitespace
func EOL() parser.WhitespaceFunc {
	return func(p *parser.Parser) *fancyerr.Error {
		if err := parser.TrySkip(p, EOF()); err == nil {
			return nil
		}

		parser.TrySkip(p, Horizontal())
		if err := parser.TrySkip(p, SingleVertical()); err == nil {
			return nil
		}
		return &fancyerr.Error{
			Message: "missing EOL",
			Primary: quickanno.Expected(p, p.Pos(), "a line ending"),
		}
	}
}

// EOF matches the end of file.
func EOF() parser.WhitespaceFunc {
	return func(p *parser.Parser) *fancyerr.Error {
		parser.TrySkip(p, Horizontal())
		if parser.TryRune(p, parser.EOF) {
			return nil
		}

		return &fancyerr.Error{
			Message: "expected EOF",
			Primary: quickanno.Expected(p, p.Pos(), "the end of file"),
		}
	}
}
