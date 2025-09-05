package code

import (
	"slices"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/interpolation"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

func String() parser.Func[*ast.String] {
	return func(p *parser.Parser) *ast.String {
		q := parser.TryAnyRune(p, '"', '`')
		if q == 0 {
			return nil
		}

		var s ast.String
		s.Open = p.PosPtr()
		// computing the position instead of using p.Pos() at the top saves us
		// allocations when String doesn't match
		s.Open.Col--
		s.Quote = byte(q)

		if s.Quote == '"' {
			p.DoInline(func() { stringContents(p, &s) })
		} else {
			stringContents(p, &s)
		}

		return &s
	}
}

func stringContents(p *parser.Parser, s *ast.String) {
	for {
		if parser.MatchesAnyRune(p, rune(s.Quote)) {
			s.Close = parser.TryRuneAt(p, rune(s.Quote))
			break
		}

		node := parser.Try(p, StringNode(s.Quote))
		if node == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "string: missing closing quote",
				Primary: quickanno.Expected(p, p.Pos(), "a closing quote for the opening quote here"),
			})
			break
		}
		s.Contents = append(s.Contents, node)
	}

	if len(s.Contents) == 0 {
		s.Contents = nil
	} else {
		s.Contents = slices.Clip(s.Contents)
	}
}

func StringNode(quote byte) parser.Func[ast.StringNode] {
	return func(p *parser.Parser) ast.StringNode {
		if txt := parser.Try(p, StringText(quote)); txt != nil {
			return txt
		} else if inter := parser.Try(p, interpolation.StringInterpolation()); inter != nil {
			return inter
		}
		return nil
	}
}

func StringText(quote byte) parser.Func[*ast.StringText] {
	return func(p *parser.Parser) *ast.StringText {
		pos := p.Pos()

		var text string
		if p.Inline() {
			text = parser.TokenWhile(p, func() bool {
				return !parser.MatchesAnyRune(p, '#', '\n', rune(quote)) ||
					parser.Matches(p, interpolation.UnambiguousHash())
			})
		} else {
			text = parser.TokenWhile(p, func() bool {
				return !parser.MatchesAnyRune(p, '#', rune(quote)) ||
					parser.Matches(p, interpolation.UnambiguousHash())
			})
		}
		if text == "" {
			return nil
		}

		return &ast.StringText{Text: text, Position: &pos}
	}
}
