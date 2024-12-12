package comment

import (
	"github.com/mavolin/corgi/v2/fancyerr"
	"github.com/mavolin/corgi/v2/fancyerr/anno"
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/whitespace"
)

func Comment() parser.Func[*ast.Comment] {
	return func(p *parser.Parser) (*ast.Comment, *fancyerr.Error) {
		if p.Inline() {
			return BlockComment()(p)
		}

		c, ok := parser.TryInOrder(p, LineComment(), BlockComment())
		if ok {
			return c, nil
		}

		return nil, &fancyerr.Error{
			Message: "missing comment",
			Primary: quickanno.Expected(p, p.Pos(), "a block or line comment"),
		}
	}
}

func LineComment() parser.Func[*ast.Comment] {
	return func(p *parser.Parser) (*ast.Comment, *fancyerr.Error) {
		c, err := lineCommentWithoutEOL()(p)
		if err != nil {
			return nil, err
		}
		parser.TrySkip(p, whitespace.EOL())
		return c, nil
	}
}

func lineCommentWithoutEOL() parser.Func[*ast.Comment] {
	return func(p *parser.Parser) (*ast.Comment, *fancyerr.Error) {
		var c ast.Comment
		c.Open = p.Pos()

		if !parser.TryToken(p, "//") {
			return nil, &fancyerr.Error{
				Message: "missing line comment",
				Primary: quickanno.Expected(p, p.Pos(), "a line comment"),
			}
		}

		c.Comment = parser.TokenWhile(p, func() bool {
			return !parser.MatchesWS(p, whitespace.EOL())
		})
		c.Close = p.PosPtr()

		return &c, nil
	}
}

func BlockComment() parser.Func[*ast.Comment] {
	return func(p *parser.Parser) (*ast.Comment, *fancyerr.Error) {
		var c ast.Comment
		c.Block = true
		c.Open = p.Pos()

		if !parser.TryToken(p, "/*") {
			return nil, &fancyerr.Error{
				Message: "missing block comment",
				Primary: quickanno.Expected(p, p.Pos(), "a block comment"),
			}
		}

		c.Comment = parser.TokenWhile(p, func() bool {
			return !parser.MatchesToken(p, "*/")
		})
		c.Close = p.PosPtr()
		if !parser.TryToken(p, "*/") {
			err := &fancyerr.Error{
				Message: "unclosed block comment",
				Primary: []fancyerr.Annotation{
					anno.NChars(p.File, c.Open, len("/*"), "this comment is never closed"),
				},
				Explanation: "Unlike line comments, block comments must be closed using `*/`.\n" +
					"Either change the `/*` to a `//` if you want a single-line comment, or add " +
					"a closing `*/` at the end of the comment.",
			}

			p.CaptureError(err)
			c.Close = nil
		} else {
			// only capture this if we didn't accidentally capture the rest of
			// the file, just because of the missing `*/`
			if p.Inline() && c.Open.Line != c.Close.Line {
				p.CaptureError(&fancyerr.Error{
					Message: "illegal placement of multiline block comment",
					Primary: []fancyerr.Annotation{
						anno.Anno(p.File, anno.Annotation{
							Context:    anno.ContextLines(c.Open, *c.Close),
							Highlight:  anno.HighlightToEOL(c.Open),
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
