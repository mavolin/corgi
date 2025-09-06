package whitespace

import (
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
)

var (
	Runes           = []rune{' ', '\t', '\r', '\n'}
	HorizontalRunes = []rune{' ', '\t'}
	VerticalRunes   = []rune{'\r', '\n'}
)

func Any() parser.WhitespaceFunc {
	return func(p *parser.Parser) bool {
		if p.Inline() {
			return Horizontal()(p)
		}

		h := parser.TrySkip(p, Horizontal())
		v := parser.TrySkip(p, Vertical())

		if !h && !v {
			return false
		}

		for h || v {
			h = parser.TrySkip(p, Horizontal())
			v = parser.TrySkip(p, Vertical())
		}
		return true
	}
}

func Horizontal() parser.WhitespaceFunc {
	return func(p *parser.Parser) bool {
		if parser.TryAnyRune(p, ' ', '\t') == 0 {
			return false
		}

		for parser.TryAnyRune(p, ' ', '\t') > 0 { //nolint:revive
		}
		return true
	}
}

func Vertical() parser.WhitespaceFunc {
	return func(p *parser.Parser) bool {
		if parser.TryAnyRune(p, '\n', '\r') == 0 {
			return false
		}

		for parser.TryAnyRune(p, '\n', '\r') != 0 { //nolint:revive
		}
		return true
	}
}

func SingleVertical() parser.WhitespaceFunc {
	return func(p *parser.Parser) bool {
		return parser.TryAnyToken(p, "\n", "\r\n", "\r") != ""
	}
}
