package whitespace

import (
	"github.com/mavolin/corgi/file/fileerr"
	"github.com/mavolin/corgi/file/hintfmt/anno"
	parser "github.com/mavolin/corgi/load/parse/internal"
)

func Any() parser.Func[struct{}] {
	return func(p *parser.Parser) (struct{}, *fileerr.Error) {
		if p.Inline() {
			return Horizontal()(p)
		}

		pos := p.Pos()

		if _, ok := parser.TryInOrder(p, Horizontal(), Vertical()); !ok {
			return struct{}{}, &fileerr.Error{
				Message:         "missing whitespace",
				ErrorAnnotation: anno.NChars(p.File, pos, 1, "expected a space, tab, LF, or CRLF line ending"),
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
	return func(p *parser.Parser) (struct{}, *fileerr.Error) {
		pos := p.Pos()

		if !parser.TryAnyRune(p, ' ', '\t') {
			return struct{}{}, &fileerr.Error{
				Message:         "missing horizontal whitespace",
				ErrorAnnotation: anno.NChars(p.File, pos, 1, "expected a space or tab"),
			}
		}

		for parser.TryAnyRune(p, ' ', '\t') {
		}
		return struct{}{}, nil
	}
}

func Vertical() parser.Func[struct{}] {
	return func(p *parser.Parser) (struct{}, *fileerr.Error) {
		pos := p.Pos()

		if !parser.TryAnyTokens(p, "\r\n", "\n") {
			return struct{}{}, &fileerr.Error{
				Message:         "missing vertical whitespace",
				ErrorAnnotation: anno.NChars(p.File, pos, 1, "expected a LF or CRLF line ending"),
			}
		}

		for parser.TryAnyTokens(p, "\r\n", "\n") {
		}
		return struct{}{}, nil
	}
}

func SingleVertical() parser.Func[struct{}] {
	return func(p *parser.Parser) (struct{}, *fileerr.Error) {
		pos := p.Pos()

		if !parser.TryAnyTokens(p, "\r\n", "\n") {
			return struct{}{}, &fileerr.Error{
				Message:         "missing vertical whitespace",
				ErrorAnnotation: anno.NChars(p.File, pos, 1, "expected a LF or CRLF line ending"),
			}
		}

		return struct{}{}, nil
	}
}
