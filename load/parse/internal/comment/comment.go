package comment

import (
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/whitespace"
)

func Comment() parser.Func[*ast.Comment] {
	return func(p *parser.Parser) *ast.Comment {
		if p.Inline() {
			return GeneralComment()(p)
		}

		c := parser.TryInOrder(p, LineComment(), GeneralComment())
		if c != nil {
			return c
		}

		return nil
	}
}

func LineComment() parser.Func[*ast.Comment] {
	return func(p *parser.Parser) *ast.Comment {
		c := parser.Try(p, lineCommentWithoutEOL())
		if c == nil {
			return nil
		}
		parser.TrySkip(p, whitespace.EOL())
		return c
	}
}

func lineCommentWithoutEOL() parser.Func[*ast.Comment] {
	return func(p *parser.Parser) *ast.Comment {
		open := parser.TryTokenAt(p, "//")
		if open == nil {
			return nil
		}

		var c ast.Comment
		c.Open = open
		c.Comment = parser.TokenWhile(p, func() bool {
			return !parser.MatchesWS(p, whitespace.EOL())
		})
		c.Until = p.Pos()

		return &c
	}
}

func GeneralComment() parser.Func[*ast.Comment] {
	return func(p *parser.Parser) *ast.Comment {
		open := parser.TryTokenAt(p, "/*")
		if open == nil {
			return nil
		}

		var c ast.Comment
		c.General = true
		c.Open = open

		c.Comment = parser.TokenWhile(p, func() bool {
			return !parser.MatchesToken(p, "*/")
		})
		c.Close = parser.TryTokenAt(p, "*/")
		if c.Close == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "unclosed general comment",
				Primary: []diagnostic.Annotation{
					anno.NRunes(p.File, *c.Open, len("/*"), "this comment is never closed"),
				},
				Explanation: "Unlike line comments, general comments must be closed using `*/`.\n" +
					"Either change the `/*` to a `//` if you want a single-line comment, or add " +
					"a closing `*/` at the end of the comment.",
			})
		}

		c.Until = p.Pos()
		if c.Close != nil {
			// only capture this if we didn't accidentally capture the rest of
			// the file, just because of the missing `*/`
			if p.Inline() && c.Open.Line != c.Close.Line {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "illegal placement of multiline general comment",
					Primary: []diagnostic.Annotation{
						anno.Anno(p.File, anno.Annotation{
							Context:    anno.ContextRange(*c.Open, c.Until),
							Highlight:  anno.HighlightToEOL(*c.Open),
							Annotation: "at this position, only single-line comments are allowed",
						}),
					},
					Explanation: "A general comment placed here must be closed on the same line it was opened on.",
				})
			}
		}

		return &c
	}
}
