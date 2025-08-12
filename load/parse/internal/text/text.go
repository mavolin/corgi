package text

import (
	"slices"
	"strings"

	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/interpolation"
	"github.com/mavolin/corgi/v2/load/parse/internal/whitespace"
)

func ArrowBlock() parser.Func[*ast.ArrowBlock] {
	return func(p *parser.Parser) *ast.ArrowBlock {
		refCol := p.Col()

		arrow := parser.TryRuneAt(p, '>')
		if arrow == nil {
			return nil
		}
		parser.TrySkip(p, whitespace.Horizontal())

		var b ast.ArrowBlock
		b.Arrow = arrow

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

		return &b
	}
}

// Line parses a text line until the terminator rune or the EOL.
func Line(term rune) parser.Func[ast.TextLine] {
	return func(p *parser.Parser) ast.TextLine {
		l := parser.Collect(p, Node(term), nil)
		if len(l) == 0 {
			return nil
		}
		return l
	}
}

func Node(term rune) parser.Func[ast.TextNode] {
	return func(p *parser.Parser) ast.TextNode {
		if t := parser.Try(p, Text(term)); t != nil {
			return t
		} else if interp := parser.Try(p, interpolation.TextInterpolation()); interp != nil {
			return interp
		} else if bi := parser.Try(p, interpolation.BadInterpolation()); bi != nil {
			return bi
		}

		return nil
	}
}

func Text(term rune) parser.Func[*ast.Text] {
	return func(p *parser.Parser) *ast.Text {
		text := parser.TokenWhile(p, func() bool {
			return !parser.MatchesAnyRune(p, term, '\r', '\n') &&
				(parser.Matches(p, interpolation.UnambiguousHash()) || !parser.MatchesAnyRune(p, '#'))
		})
		text = strings.TrimRight(text, " \t")
		if text == "" {
			return nil
		}

		var t ast.Text
		t.Text = text
		t.Position = p.PosPtr()
		t.Position.Col -= len(text) // save allocations, only works bc t is horizontal

		return &t
	}
}

// VerbatimLine parses a text line consisting only of text nodes.
// Sequences that when calling the regular [Line] would normally yield
// other nodes, such as interpolation, will be included in text nodes.
// The only exception are HashBrackets.
func VerbatimLine(term rune) parser.Func[ast.TextLine] {
	return func(p *parser.Parser) ast.TextLine {
		l := parser.Collect(p, VerbatimNode(term), nil)
		if len(l) == 0 {
			return nil
		}
		return l
	}
}

// VerbatimNode parses a text node consisting only of text nodes.
// Sequences that when calling the regular [Node] would normally yield
// other nodes, such as interpolation, will be included in text nodes.
// The only exception are escaped right brackets.
func VerbatimNode(term rune) parser.Func[ast.TextNode] {
	return func(p *parser.Parser) ast.TextNode {
		if t := parser.Try(p, VerbatimText(term)); t != nil {
			return t
		} else if escapedRBracket := parser.Try(p, interpolation.EscapedRBracket()); escapedRBracket != nil {
			return escapedRBracket
		}

		return nil
	}
}

// VerbatimText parses a text node consisting only of text nodes.
// Sequences that when calling the regular [Text] would normally return to
// allow parsing interpolation, will be included in text nodes.
// The only exception are escaped right brackets.
func VerbatimText(term rune) parser.Func[*ast.Text] {
	return func(p *parser.Parser) *ast.Text {
		pos := p.Pos()

		text := parser.TokenWhile(p, func() bool {
			return !parser.MatchesAnyRune(p, term, '\r', '\n') && !parser.MatchesToken(p, "#]")
		})
		text = strings.TrimRight(text, " \t")
		if text == "" {
			return nil
		}

		return &ast.Text{Text: text, Position: &pos}
	}
}
