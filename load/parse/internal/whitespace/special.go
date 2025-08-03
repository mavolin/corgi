package whitespace

import (
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
)

// EOL matches the EOL or EOF with optionally preceding horizontal whitespace.
func EOL() parser.WhitespaceFunc {
	return func(p *parser.Parser) bool {
		if parser.TrySkip(p, EOF()) {
			return true
		}

		parser.TrySkip(p, Horizontal())
		return parser.TrySkip(p, SingleVertical())
	}
}

// EOF matches the end of file.
func EOF() parser.WhitespaceFunc {
	return func(p *parser.Parser) bool {
		parser.TrySkip(p, Horizontal())
		return parser.TryRune(p, parser.EOF)
	}
}
