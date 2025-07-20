package text

import (
	"slices"
	"strings"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/interpolation"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/whitespace"
)

func ArrowBlock() parser.Func[*ast.ArrowBlock] {
	return func(p *parser.Parser) (*ast.ArrowBlock, *diagnostic.Diagnostic) {
		var b ast.ArrowBlock
		refCol := p.Col()

		b.Arrow = parser.TryRuneAt(p, '>')
		if b.Arrow == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "missing arrow block",
				Primary: quickanno.Expected(p, p.Pos(), "an arrow block"),
			}
		}
		parser.TrySkip(p, whitespace.Horizontal())

		b.Lines = make(ast.TextBlock, 0, 36)
		for {
			line := parser.Try(p, Line('\n'))
			if line == nil {
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

		return &b, nil
	}
}

// Line parses a text line until the terminator rune or the EOL.
func Line(term rune) parser.Func[ast.TextLine] {
	return func(p *parser.Parser) (ast.TextLine, *diagnostic.Diagnostic) {
		l := parser.Collect(p, Node(term), 8, nil)
		if len(l) == 0 {
			return nil, &diagnostic.Diagnostic{
				Message: "missing text line",
				Primary: quickanno.Expected(p, p.Pos(), "text"),
			}
		}
		return l, nil
	}
}

// VerbatimLine parses a text line consisting only of text nodes.
// Sequences that when calling the regular [Line] would normally yield
// other nodes, such as interpolation, will be included in text nodes.
// The only exception are HashBrackets.
func VerbatimLine(term rune) parser.Func[ast.TextLine] {
	return func(p *parser.Parser) (ast.TextLine, *diagnostic.Diagnostic) {
		var t ast.Text
		t.Position = p.PosPtr()

		t.Text = parser.TokenWhile(p, func() bool {
			return !parser.MatchesAnyRune(p, term, '\r', '\n')
		})
		t.Text = strings.TrimRight(t.Text, " \t")
		if t.Text == "" {
			return nil, &diagnostic.Diagnostic{
				Message: "missing text line",
				Primary: quickanno.Expected(p, p.Pos(), "text"),
			}
		}

		return ast.TextLine{&t}, nil
	}
}

func Node(term rune) parser.Func[ast.TextNode] {
	return func(p *parser.Parser) (ast.TextNode, *diagnostic.Diagnostic) {
		if t := parser.Try(p, Text(term)); t != nil {
			return t, nil
		} else if interp := parser.Try(p, interpolation.TextInterpolation()); interp != nil {
			return interp, nil
		} else if bi := parser.Try(p, interpolation.BadInterpolation()); bi != nil {
			return bi, nil
		}

		return nil, &diagnostic.Diagnostic{
			Message: "missing text node",
			Primary: quickanno.Expected(p, p.Pos(), "text or interpolation"),
		}
	}
}

func Text(term rune) parser.Func[*ast.Text] {
	return func(p *parser.Parser) (*ast.Text, *diagnostic.Diagnostic) {
		var t ast.Text
		t.Position = p.PosPtr()

		t.Text = parser.TokenWhile(p, func() bool {
			return !parser.MatchesAnyRune(p, term, '\r', '\n') &&
				(parser.Matches(p, interpolation.UnambiguousHash()) || !parser.MatchesAnyRune(p, '#'))
		})
		t.Text = strings.TrimRight(t.Text, " \t")
		if t.Text == "" {
			return nil, &diagnostic.Diagnostic{
				Message: "missing text",
				Primary: quickanno.Expected(p, p.Pos(), "text"),
			}
		}

		return &t, nil
	}
}
