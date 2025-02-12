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
	return func(p *parser.Parser) (*ast.String, *diagnostic.Diagnostic) {
		var s ast.String
		s.Open = p.PosPtr()

		q := parser.TryAnyRune(p, '"', '`')
		if q < 0 {
			return nil, &diagnostic.Diagnostic{
				Message: "missing string",
				Primary: quickanno.Expected(p, p.Pos(), "a string"),
			}
		}
		s.Quote = byte(q)

		if s.Quote == '"' {
			p.DoInline(func() { stringContents(p, &s) })
		} else {
			stringContents(p, &s)
		}

		return &s, nil
	}
}

func stringContents(p *parser.Parser, s *ast.String) {
	s.Contents = make([]ast.StringNode, 0, 12)

	for {
		if parser.MatchesAnyRune(p, rune(s.Quote)) {
			s.Close = parser.TryRuneAt(p, rune(s.Quote))
			break
		}

		node, err := parser.TryErr(p, StringNode(s.Quote))
		if err != nil {
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
	return func(p *parser.Parser) (ast.StringNode, *diagnostic.Diagnostic) {
		if txt := parser.Try(p, StringText(quote)); txt != nil {
			return txt, nil
		} else if interp := parser.Try(p, interpolation.StringInterpolation()); interp != nil {
			return interp, nil
		}
		return nil, &diagnostic.Diagnostic{
			Message: "missing string node",
			Primary: quickanno.Expected(p, p.Pos(), "text or interpolation"),
		}
	}
}

func StringText(quote byte) parser.Func[*ast.StringText] {
	return func(p *parser.Parser) (*ast.StringText, *diagnostic.Diagnostic) {
		var t ast.StringText
		t.Position = p.PosPtr()

		if p.Inline() {
			t.Text = parser.TokenWhile(p, func() bool {
				return !parser.MatchesAnyRune(p, '#', '\n', rune(quote)) ||
					parser.Matches(p, interpolation.UnambiguousHash())
			})
		} else {
			t.Text = parser.TokenWhile(p, func() bool {
				return !parser.MatchesAnyRune(p, '#', rune(quote)) ||
					parser.Matches(p, interpolation.UnambiguousHash())
			})
		}
		if t.Text == "" {
			return nil, &diagnostic.Diagnostic{
				Message: "missing string text",
				Primary: quickanno.Expected(p, *t.Position, "string text"),
			}
		}

		return &t, nil
	}
}
