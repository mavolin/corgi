package body

import (
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/whitespace"
)

func BracketText() parser.Func[*ast.BracketText] {
	return func(p *parser.Parser) (*ast.BracketText, *diagnostic.Diagnostic) {
		var bt ast.BracketText

		bt.LBracket = parser.TryRuneAt(p, '[')
		if bt.LBracket == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "missing bracket text",
				Primary: quickanno.Expected(p, p.Pos(), "expected a `[` here"),
			}
		}

		bt.Lines = parser.Collect(p, textLine(']'), 24, whitespace.Any())

		parser.TrySkip(p, whitespace.Any())
		bt.RBracket = parser.TryRuneAt(p, ']')
		if bt.RBracket == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "bracket text: unclosed '['",
				Primary: quickanno.Expected(p, *bt.LBracket, "expected a `]` for the opening `[` here"),
			})
		}

		return &bt, nil
	}
}

var textLine func(term rune) parser.Func[ast.TextLine]

func SetTextLine(f func(term rune) parser.Func[ast.TextLine]) {
	textLine = f
}
