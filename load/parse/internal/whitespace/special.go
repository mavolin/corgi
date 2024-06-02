package whitespace

import (
	"github.com/mavolin/corgi/fancyerr"
	parser "github.com/mavolin/corgi/load/parse/internal"
	"github.com/mavolin/corgi/load/parse/internal/quickanno"
)

// EOL matches the EOL or EOF with optionally preceding horizontal whitespace
func EOL() parser.Func[struct{}] {
	return func(p *parser.Parser) (struct{}, *fancyerr.Error) {
		if _, ok := parser.Try(p, EOF()); ok {
			return struct{}{}, nil
		}

		parser.Try(p, Horizontal())
		if _, ok := parser.Try(p, SingleVertical()); ok {
			return struct{}{}, nil
		}
		return struct{}{}, &fancyerr.Error{
			Message: "missing EOL",
			Primary: quickanno.Expected(p, p.Pos(), "a line ending"),
		}
	}
}

// EOF matches the end of file.
func EOF() parser.Func[struct{}] {
	return func(p *parser.Parser) (struct{}, *fancyerr.Error) {
		parser.Try(p, Horizontal())
		if parser.TryRune(p, parser.EOF) {
			return struct{}{}, nil
		}

		return struct{}{}, &fancyerr.Error{
			Message: "expected EOF",
			Primary: quickanno.Expected(p, p.Pos(), "the end of file"),
		}
	}
}
