package comment

import (
	"slices"

	"github.com/mavolin/corgi/file/ast"
	"github.com/mavolin/corgi/file/fileerr"
	"github.com/mavolin/corgi/file/hintfmt/anno"
	parser "github.com/mavolin/corgi/load/parse/internal"
	"github.com/mavolin/corgi/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/load/parse/internal/whitespace"
)

// InlineGroup parses a comment group that mustn't contain vertical
// whitespace.
func InlineGroup() parser.Func[*ast.CommentGroup] {
	return func(p *parser.Parser) (*ast.CommentGroup, *fileerr.Error) {
		var c *ast.Comment
		var ok bool
		p.DoInline(func() { c, ok = parser.Try(p, BlockComment()) })
		if ok {
			return &ast.CommentGroup{Comments: []*ast.Comment{c}}, nil
		}

		c, ok = parser.Try(p, lineCommentWithoutEOL())
		if !ok {
			return nil, &fileerr.Error{
				Message:         "missing inline comment",
				ErrorAnnotation: quickanno.Expected(p, p.Pos(), "a single-line block comment"),
			}
		}

		p.CaptureError(&fileerr.Error{
			Message:         "illegal placement of line comment",
			ErrorAnnotation: anno.Range(p.File, c.Open, p.Pos(), "cannot place a line comment here"),
		})
		return &ast.CommentGroup{Comments: []*ast.Comment{c}}, nil
	}
}

// LoneGroup parses a comment group on a separate line.
func LoneGroup() parser.Func[*ast.CommentGroup] {
	return func(p *parser.Parser) (*ast.CommentGroup, *fileerr.Error) {
		c, ok := parser.Try(p, BlockComment())
		if ok {
			return &ast.CommentGroup{Comments: []*ast.Comment{c}}, nil
		}

		c, ok = parser.Try(p, LineComment())
		if !ok {
			return nil, &fileerr.Error{
				Message:         "missing comment",
				ErrorAnnotation: quickanno.Expected(p, p.Pos(), "a block or line comment"),
			}
		}

		cs := make([]*ast.Comment, 0, 32)
		for ok {
			cs = append(cs, c)
			c, ok = parser.Try(p, func(p *parser.Parser) (*ast.Comment, *fileerr.Error) {
				parser.While(p, func() bool {
					return parser.Matches(p, whitespace.Horizontal())
				})

				return LineComment()(p)
			})
		}
		return &ast.CommentGroup{Comments: slices.Clip(cs)}, nil
	}
}
