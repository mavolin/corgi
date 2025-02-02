package text

import (
	"slices"
	"strings"

	"github.com/mavolin/corgi/v2/fancyerr"
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/interpolation"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/whitespace"
)

func ArrowBlock() parser.Func[*ast.ArrowBlock] {
	return func(p *parser.Parser) (*ast.ArrowBlock, *fancyerr.Error) {
		refCol := p.Col()
		b := &ast.ArrowBlock{Arrow: p.Pos()}
		if !parser.TryRune(p, '>') {
			return nil, &fancyerr.Error{
				Message: "missing arrow block",
				Primary: quickanno.Expected(p, p.Pos(), "an arrow block"),
			}
		}

		parser.TrySkip(p, whitespace.Horizontal())

		b.Lines = make(ast.TextBlock, 0, 36)
		for {
			line, ok := parser.TryOk(p, Line('\n'))
			if !ok {
				break
			}
			b.Lines = append(b.Lines, line)

			parser.TrySkip(p, whitespace.Any())
			if p.Col() <= refCol {
				parser.RestoreWS(p)
				break
			}
		}
		if len(b.Lines) == 0 {
			b.Lines = nil
		} else {
			b.Lines = slices.Clip(b.Lines)
		}

		return b, nil
	}
}

// Line parses a text line until the terminator rune or the EOL.
func Line(term rune) parser.Func[ast.TextLine] {
	return func(p *parser.Parser) (ast.TextLine, *fancyerr.Error) {
		l := make(ast.TextLine, 0, 8)

		for {
			n, ok := parser.TryOk(p, Node(term))
			if !ok {
				break
			}
			l = append(l, n)
		}
		if len(l) == 0 {
			return nil, &fancyerr.Error{
				Message: "missing text line",
				Primary: quickanno.Expected(p, p.Pos(), "text"),
			}
		}

		return slices.Clip(l), nil
	}
}

func Node(term rune) parser.Func[ast.TextNode] {
	return func(p *parser.Parser) (ast.TextNode, *fancyerr.Error) {
		if t, ok := parser.TryOk(p, Text(term)); ok {
			return t, nil
		} else if interp, ok := parser.TryOk(p, interpolation.TextInterpolation()); ok {
			return interp, nil
		} else if bi, ok := parser.TryOk(p, interpolation.BadInterpolation()); ok {
			return bi, nil
		}

		return nil, &fancyerr.Error{
			Message: "missing text node",
			Primary: quickanno.Expected(p, p.Pos(), "text or interpolation"),
		}
	}
}

func Text(term rune) parser.Func[*ast.Text] {
	return func(p *parser.Parser) (*ast.Text, *fancyerr.Error) {
		t := &ast.Text{Position: p.Pos()}

		t.Text = parser.TokenWhile(p, func() bool {
			return !parser.MatchesAnyRune(p, term, term, '\r', '\n') &&
				(parser.Matches(p, interpolation.UnambiguousHash()) || !parser.MatchesAnyRune(p, '#'))
		})
		t.Text = strings.TrimRight(t.Text, " \t")
		if t.Text == "" {
			return nil, &fancyerr.Error{
				Message: "missing text",
				Primary: quickanno.Expected(p, p.Pos(), "text"),
			}
		}

		return t, nil
	}
}
