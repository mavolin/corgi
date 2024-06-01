package whitespace

import (
	"github.com/mavolin/corgi/file/fileerr"
	"github.com/mavolin/corgi/file/hintfmt/anno"
	parser "github.com/mavolin/corgi/load/parse/internal"
)

// EOL matches the EOL or EOF with optionally preceding horizontal whitespace
func EOL() parser.Func[struct{}] {
	return func(p *parser.Parser) (struct{}, *fileerr.Error) {
		if _, ok := parser.Try(p, EOF()); ok {
			return struct{}{}, nil
		}

		parser.Try(p, Horizontal())
		if _, ok := parser.Try(p, SingleVertical()); ok {
			return struct{}{}, nil
		}
		return struct{}{}, &fileerr.Error{
			Message:         "missing EOL",
			ErrorAnnotation: anno.NChars(p.File, p.Pos(), 1, "expected a LF or CRLF line ending"),
		}
	}
}

// EOF matches the end of file.
func EOF() parser.Func[struct{}] {
	return func(p *parser.Parser) (struct{}, *fileerr.Error) {
		parser.Try(p, Horizontal())
		if parser.TryRune(p, parser.EOF) {
			return struct{}{}, nil
		}

		return struct{}{}, &fileerr.Error{
			Message:         "expected EOF",
			ErrorAnnotation: anno.NChars(p.File, p.Pos(), 1, "expected EOF"),
		}
	}
}
