package whitespace

import (
	"github.com/mavolin/corgi/v2/file/diagnostic"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

var (
	Runes           = []rune{' ', '\t', '\r', '\n'}
	HorizontalRunes = []rune{' ', '\t'}
	VerticalRunes   = []rune{'\r', '\n'}
)

func Any() parser.WhitespaceFunc {
	return func(p *parser.Parser) *diagnostic.Diagnostic {
		if p.Inline() {
			return Horizontal()(p)
		}

		pos := p.Pos()

		h := parser.TrySkip(p, Horizontal())
		v := parser.TrySkip(p, Vertical())

		if !h && !v {
			return &diagnostic.Diagnostic{
				Message: "missing whitespace",
				Primary: quickanno.Expected(p, pos, "a space, tab, or line ending"),
			}
		}

		for h || v {
			h = parser.TrySkip(p, Horizontal())
			v = parser.TrySkip(p, Vertical())
		}
		return nil
	}
}

func Horizontal() parser.WhitespaceFunc {
	return func(p *parser.Parser) *diagnostic.Diagnostic {
		pos := p.Pos()

		if parser.TryAnyRune(p, ' ', '\t') <= 0 {
			return &diagnostic.Diagnostic{
				Message: "missing horizontal whitespace",
				Primary: quickanno.Expected(p, pos, "a space or tab"),
			}
		}

		for parser.TryAnyRune(p, ' ', '\t') > 0 {
		}
		return nil
	}
}

func Vertical() parser.WhitespaceFunc {
	return func(p *parser.Parser) *diagnostic.Diagnostic {
		pos := p.Pos()

		if parser.TryAnyToken(p, "\n", "\r\n") == "" {
			return &diagnostic.Diagnostic{
				Message: "missing vertical whitespace",
				Primary: quickanno.Expected(p, pos, "a line ending"),
			}
		}

		for parser.TryAnyToken(p, "\n", "\r\n") != "" {
		}
		return nil
	}
}

func SingleVertical() parser.WhitespaceFunc {
	return func(p *parser.Parser) *diagnostic.Diagnostic {
		pos := p.Pos()

		if parser.TryAnyToken(p, "\n", "\r\n") == "" {
			return &diagnostic.Diagnostic{
				Message: "missing vertical whitespace",
				Primary: quickanno.Expected(p, pos, "a line ending"),
			}
		}

		return nil
	}
}
