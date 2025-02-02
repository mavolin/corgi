package code

import (
	"github.com/mavolin/corgi/v2/fancyerr"
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/whitespace"
)

func ImplicitCodeLine() parser.Func[*ast.ImplicitCodeLine] {
	return func(p *parser.Parser) (*ast.ImplicitCodeLine, *fancyerr.Error) {
		s, ok := parser.TryOk(p, ParsedStatement())
		if !ok {
			return nil, &fancyerr.Error{
				Message: "missing implicit code line",
				Primary: quickanno.Expected(p, p.Pos(), "an implicit code line"),
			}
		}

		return &ast.ImplicitCodeLine{Statement: *s}, nil
	}
}

func ExplicitCodeLine() parser.Func[*ast.ExplicitCodeLine] {
	return func(p *parser.Parser) (*ast.ExplicitCodeLine, *fancyerr.Error) {
		e := &ast.ExplicitCodeLine{Minus: p.Pos()}
		if !parser.TryRune(p, '-') || !parser.TrySkipOk(p, whitespace.Horizontal()) {
			return nil, &fancyerr.Error{
				Message: "missing explicit code line",
				Primary: quickanno.Expected(p, p.Pos(), "an explicit code line"),
			}
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		e.Statement = parser.Must(p, Statement())
		return e, nil
	}
}
