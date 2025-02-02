package code

import (
	"slices"

	"github.com/mavolin/corgi/v2/fancyerr"
	"github.com/mavolin/corgi/v2/fancyerr/anno"
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/interpolation"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

func String() parser.Func[*ast.String] {
	return func(p *parser.Parser) (*ast.String, *fancyerr.Error) {
		s := &ast.String{Open: p.Pos()}

		q := parser.TryAnyRune(p, '"', '`')
		if q < 0 {
			return nil, &fancyerr.Error{
				Message: "missing string",
				Primary: quickanno.Expected(p, p.Pos(), "a string"),
			}
		}
		s.Quote = byte(q)

		s.Contents = make([]ast.StringNode, 0, 16)
		if s.Quote == '"' {
			p.DoInline(func() {
				stringContents(p, s)
			})
		} else {
			stringContents(p, s)
		}

		return s, nil
	}
}

func stringContents(p *parser.Parser, s *ast.String) {
	s.Contents = make([]ast.StringNode, 0, 12)

	for {
		if parser.MatchesAnyRune(p, rune(s.Quote)) {
			s.Contents = slices.Clip(s.Contents)
			s.Close = p.PosPtr()
			if !parser.TryRune(p, rune(s.Quote)) {
				s.Close = nil
			}
			return
		} else if parser.MatchesAnyRune(p, parser.EOF) {
			err := &fancyerr.Error{Message: "missing closing quote"}
			if s.Open.Line <= p.Pos().Line-3 {
				err.Primary = []fancyerr.Annotation{
					anno.Range(p.File, quickanno.DeltaPos(s.Open, 0, 1), p.Pos(), "expected a closing quote"),
				}
			} else {
				err.Primary = quickanno.Expected(p, p.Pos(), "a closing quote")
			}
			err.Secondary = []fancyerr.Annotation{anno.Position(p.File, s.Open, "for the opening quote here")}
			p.CaptureError(err)
			return
		} else if p.Inline() && parser.MatchesAnyRune(p, '\n') {
			p.CaptureError(&fancyerr.Error{
				Message: "missing closing quote",
				Primary: []fancyerr.Annotation{
					anno.Range(p.File, quickanno.DeltaPos(s.Open, 0, 1), p.Pos(), "expected a closing quote"),
				},
				Secondary: []fancyerr.Annotation{anno.Position(p.File, s.Open, "for the opening quote here")},
			})
			s.Contents = slices.Clip(s.Contents)
			return
		}

		node, err := parser.Try(p, StringNode(s.Quote))
		if err != nil {
			p.CaptureError(err)
			s.Contents = slices.Clip(s.Contents)
			return
		}
		s.Contents = append(s.Contents, node)
	}
}

func StringNode(quote byte) parser.Func[ast.StringNode] {
	return func(p *parser.Parser) (ast.StringNode, *fancyerr.Error) {
		txt, ok := parser.TryOk(p, StringText(quote))
		if ok {
			return txt, nil
		}
		interp, ok := parser.TryOk(p, interpolation.StringInterpolation())
		if ok {
			return interp, nil
		}

		return nil, &fancyerr.Error{
			Message: "missing string node",
			Primary: quickanno.Expected(p, p.Pos(), "text or interpolation"),
		}
	}
}

func StringText(quote byte) parser.Func[*ast.StringText] {
	return func(p *parser.Parser) (*ast.StringText, *fancyerr.Error) {
		t := &ast.StringText{Position: p.Pos()}
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
			return t, &fancyerr.Error{
				Message: "missing string text",
				Primary: quickanno.Expected(p, t.Position, "string text"),
			}
		}
		return t, nil
	}
}
