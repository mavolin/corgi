package whitespace

import (
	"github.com/mavolin/corgi/fancyerr"
	parser "github.com/mavolin/corgi/load/parse/internal"
	"github.com/mavolin/corgi/load/parse/internal/quickanno"
)

var Runes = []rune{' ', '\t', '\r', '\n'}

func Any() parser.Func[struct{}] {
	return func(p *parser.Parser) (struct{}, *fancyerr.Error) {
		if p.Inline() {
			return Horizontal()(p)
		}

		pos := p.Pos()

		if _, ok := parser.TryInOrder(p, Horizontal(), Vertical()); !ok {
			return struct{}{}, &fancyerr.Error{
				Message: "missing whitespace",
				Primary: quickanno.Expected(p, pos, "a space, tab, or line ending"),
			}
		}

		for {
			if _, ok := parser.TryInOrder(p, Horizontal(), Vertical()); !ok {
				return struct{}{}, nil
			}
		}
	}
}

func Horizontal() parser.Func[struct{}] {
	return func(p *parser.Parser) (struct{}, *fancyerr.Error) {
		pos := p.Pos()

		if !parser.TryAnyRune(p, ' ', '\t') {
			return struct{}{}, &fancyerr.Error{
				Message: "missing horizontal whitespace",
				Primary: quickanno.Expected(p, pos, "a space or tab"),
			}
		}

		for parser.TryAnyRune(p, ' ', '\t') {
		}
		return struct{}{}, nil
	}
}

func Vertical() parser.Func[struct{}] {
	return func(p *parser.Parser) (struct{}, *fancyerr.Error) {
		pos := p.Pos()

		if !parser.TryAnyTokens(p, "\r\n", "\n") {
			return struct{}{}, &fancyerr.Error{
				Message: "missing vertical whitespace",
				Primary: quickanno.Expected(p, pos, "a line ending"),
			}
		}

		for parser.TryAnyTokens(p, "\r\n", "\n") {
		}
		return struct{}{}, nil
	}
}

func SingleVertical() parser.Func[struct{}] {
	return func(p *parser.Parser) (struct{}, *fancyerr.Error) {
		pos := p.Pos()

		if !parser.TryAnyTokens(p, "\r\n", "\n") {
			return struct{}{}, &fancyerr.Error{
				Message: "missing vertical whitespace",
				Primary: quickanno.Expected(p, pos, "a line ending"),
			}
		}

		return struct{}{}, nil
	}
}
