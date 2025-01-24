package body

import (
	"slices"

	"github.com/mavolin/corgi/v2/fancyerr"
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/whitespace"
)

func BracketText() parser.Func[*ast.BracketText] {
	return func(p *parser.Parser) (*ast.BracketText, *fancyerr.Error) {
		bt := &ast.BracketText{LBracket: p.Pos()}
		if !parser.TryRune(p, '[') {
			return nil, &fancyerr.Error{
				Message: "missing bracket text",
				Primary: quickanno.Expected(p, bt.LBracket, "expected a `[` here"),
			}
		}

		bt.Lines = make([]ast.TextLine, 0, 24)
		for {
			parser.TrySkip(p, whitespace.Any())
			l, ok := parser.TryOk(p, textLine(']'))
			if !ok {
				break
			}
			bt.Lines = append(bt.Lines, l)
		}
		if len(bt.Lines) == 0 {
			bt.Lines = nil
		} else {
			bt.Lines = slices.Clip(bt.Lines)
		}

		parser.TrySkip(p, whitespace.Any())
		bt.RBracket = p.PosPtr()
		if !parser.TryRune(p, ']') {
			bt.RBracket = nil
			p.CaptureError(&fancyerr.Error{
				Message: "bracket text: unclosed '['",
				Primary: quickanno.Expected(p, bt.LBracket, "expected a `]` for the opening `[` here"),
			})
		}

		return bt, nil
	}
}

var textLine func(term rune) parser.Func[ast.TextLine]

func SetTextLine(f func(term rune) parser.Func[ast.TextLine]) {
	textLine = f
}
