package attribute

import (
	"github.com/mavolin/corgi/file/ast"
	"github.com/mavolin/corgi/file/fileerr"
	parser "github.com/mavolin/corgi/load/parse/internal"
	"github.com/mavolin/corgi/load/parse/internal/quickanno"
)

func AndPlaceholder() parser.Func[*ast.AndPlaceholder] {
	return func(p *parser.Parser) (*ast.AndPlaceholder, *fileerr.Error) {
		pos := p.Pos()
		if !parser.TryRune(p, '&') {
			return nil, &fileerr.Error{
				Message:         "missing `&`",
				ErrorAnnotation: quickanno.Expected(p, pos, "an and placeholder (`&`)"),
			}
		}

		return &ast.AndPlaceholder{Position: pos}, nil
	}
}
