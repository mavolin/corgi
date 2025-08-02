package comment

import (
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/whitespace"
)

func Comment() parser.Func[*ast.Comment] {
	return func(p *parser.Parser) (*ast.Comment, *diagnostic.Diagnostic) {
		if p.Inline() {
			return GeneralComment()(p)
		}

		c := parser.TryInOrder(p, LineComment(), GeneralComment())
		if c != nil {
			return c, nil
		}

		return nil, &diagnostic.Diagnostic{
			Message: "missing comment",
			Primary: quickanno.Expected(p, p.Pos(), "a block or line comment"),
		}
	}
}

func LineComment() parser.Func[*ast.Comment] {
	return func(p *parser.Parser) (*ast.Comment, *diagnostic.Diagnostic) {
		c, err := parser.TryErr(p, lineCommentWithoutEOL())
		if err != nil {
			return nil, err
		}
		parser.TrySkip(p, whitespace.EOL())
		return c, nil
	}
}

func lineCommentWithoutEOL() parser.Func[*ast.Comment] {
	return func(p *parser.Parser) (*ast.Comment, *diagnostic.Diagnostic) {
		open := parser.TryTokenAt(p, "//")
		if open == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "missing line comment",
				Primary: quickanno.Expected(p, p.Pos(), "a line comment"),
			}
		}
		var c ast.Comment
		c.Open = open
		c.Comment = parser.TokenWhile(p, func() bool {
			return !parser.MatchesWS(p, whitespace.EOL())
		})
		c.Until = p.Pos()

		return &c, nil
	}
}

func GeneralComment() parser.Func[*ast.Comment] {
	return func(p *parser.Parser) (*ast.Comment, *diagnostic.Diagnostic) {
		open := parser.TryTokenAt(p, "/*")
		if open == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "missing block comment",
				Primary: quickanno.Expected(p, p.Pos(), "a general comment"),
			}
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
				Message: "unclosed block comment",
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
					Message: "illegal placement of multiline block comment",
					Primary: []diagnostic.Annotation{
						anno.Anno(p.File, anno.Annotation{
							Context:    anno.ContextRange(*c.Open, c.Until),
							Highlight:  anno.HighlightToEOL(*c.Open),
							Annotation: "at this position, only single-line comments are allowed",
						}),
					},
					Explanation: "A block comment placed here must be closed on the same line it was opened on.",
				})
			}
		}

		return &c, nil
	}
}
