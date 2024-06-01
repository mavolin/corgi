package comment

import (
	"github.com/mavolin/corgi/file/ast"
	"github.com/mavolin/corgi/file/fileerr"
	"github.com/mavolin/corgi/file/hintfmt/anno"
	parser "github.com/mavolin/corgi/load/parse/internal"
	"github.com/mavolin/corgi/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/load/parse/internal/whitespace"
)

func Comment() parser.Func[*ast.Comment] {
	return func(p *parser.Parser) (*ast.Comment, *fileerr.Error) {
		c, ok := parser.TryInOrder(p, LineComment(), BlockComment())
		if ok {
			return c, nil
		}

		return nil, &fileerr.Error{
			Message:         "missing comment",
			ErrorAnnotation: quickanno.Expected(p, p.Pos(), "expected a block or line comment"),
		}
	}
}

func LineComment() parser.Func[*ast.Comment] {
	return func(p *parser.Parser) (*ast.Comment, *fileerr.Error) {
		c, err := lineCommentWithoutEOL()(p)
		if err != nil {
			return nil, err
		}
		parser.Must(p, whitespace.EOL())
		return c, nil
	}
}

func lineCommentWithoutEOL() parser.Func[*ast.Comment] {
	return func(p *parser.Parser) (*ast.Comment, *fileerr.Error) {
		var c ast.Comment
		c.Open = p.Pos()

		if !parser.TryToken(p, "//") {
			return nil, &fileerr.Error{
				Message:         "missing line comment",
				ErrorAnnotation: quickanno.Expected(p, p.Pos(), "a line comment"),
			}
		}

		c.Comment = parser.While(p, func() bool {
			return !parser.Matches(p, whitespace.EOL())
		})
		c.Close = p.PosPtr()

		return &c, nil
	}
}

func BlockComment() parser.Func[*ast.Comment] {
	return func(p *parser.Parser) (*ast.Comment, *fileerr.Error) {
		var c ast.Comment
		c.Block = true
		c.Open = p.Pos()

		if !parser.TryToken(p, "/*") {
			return nil, &fileerr.Error{
				Message:         "missing block comment",
				ErrorAnnotation: quickanno.Expected(p, p.Pos(), "a block comment"),
			}
		}

		c.Comment = parser.While(p, func() bool {
			if p.Inline() && parser.Matches(p, whitespace.Vertical()) {
				return false
			}
			return !parser.MatchesToken(p, "*/")
		})
		c.Close = p.PosPtr()
		if !parser.TryToken(p, "*/") {
			err := &fileerr.Error{
				Message:         "unclosed block comment",
				ErrorAnnotation: anno.NChars(p.File, c.Open, len("/*"), "this comment is never closed"),
				Hints: []fileerr.Hint{
					{Hint: "close the comment with `*/`", Example: "`/* foo */`"},
				},
			}
			if p.Inline() {
				err.Hints = append(err.Hints, fileerr.Hint{
					Hint: "the current context doesn't allow newlines: make sure you close the comment on the same line",
				})
			}

			p.CaptureError(err)
		}

		return &c, nil
	}
}

func singleLineBlockComment() parser.Func[*ast.Comment] {
	return func(p *parser.Parser) (*ast.Comment, *fileerr.Error) {
		var c ast.Comment
		c.Block = true
		c.Open = p.Pos()

		if !parser.TryToken(p, "/*") {
			return nil, &fileerr.Error{
				Message:         "missing block comment",
				ErrorAnnotation: quickanno.Expected(p, p.Pos(), "a block comment"),
			}
		}

		c.Comment = parser.While(p, func() bool {
			return !parser.Matches(p, whitespace.Vertical()) && !parser.MatchesToken(p, "*/")
		})
		c.Close = p.PosPtr()
		if !parser.TryToken(p, "*/") {
			return nil, &fileerr.Error{
				Message:         "unclosed block comment",
				ErrorAnnotation: anno.NChars(p.File, c.Open, len("/*"), "this comment is never closed"),
				Hints: []fileerr.Hint{
					{Hint: "close the comment with `*/`", Example: "`/* foo */`"},
				},
			}
		}

		return &c, nil
	}
}
