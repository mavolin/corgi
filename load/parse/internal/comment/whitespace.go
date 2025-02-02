package comment

import (
	"slices"

	"github.com/mavolin/corgi/v2/fancyerr"
	"github.com/mavolin/corgi/v2/fancyerr/anno"
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/whitespace"
)

// OrHorizontalWhitespace parses and captures inline comments and horizontal
// whitespace.
func OrHorizontalWhitespace() parser.WhitespaceFunc {
	return func(p *parser.Parser) *fancyerr.Error {
		hasWS := parser.TrySkipOk(p, whitespace.Horizontal())
		c, hasComment := parser.TryOk(p, GeneralComment())
		if !hasWS && !hasComment {
			return &fancyerr.Error{
				Message: "missing horizontal whitespace",
				Primary: quickanno.Expected(p, p.Pos(), "a space, tab, or a block comment"),
			}
		}
		if hasComment {
			p.CaptureComment(&ast.CommentGroup{Comments: []*ast.Comment{c}})
		}

		for hasWS || hasComment {
			hasWS = parser.TrySkipOk(p, whitespace.Horizontal())
			c, hasComment = parser.TryOk(p, GeneralComment())
			if hasComment {
				p.CaptureComment(&ast.CommentGroup{Comments: []*ast.Comment{c}})
			}
		}
		return nil
	}
}

// AndEOS matches the end of statement, optionally preceded by block comments.
// It consumes trailing horizontal whitespace; it doesn't consume the EOL.
// If the line ends with a line comment, AndEOS accepts, but does not consume
// it.
func AndEOS() parser.WhitespaceFunc {
	return func(p *parser.Parser) *fancyerr.Error {
		pos := p.Pos()
		for {
			parser.TrySkip(p, whitespace.Horizontal())
			c, hasComment := parser.TryOk(p, GeneralComment())
			if !hasComment {
				break
			}
			p.CaptureComment(&ast.CommentGroup{Comments: []*ast.Comment{c}})
		}

		switch {
		case parser.TryRune(p, ';'):
			return nil
		case parser.MatchesWS(p, whitespace.EOL()):
			return nil
		case parser.MatchesToken(p, "}"):
			return nil
		case parser.MatchesToken(p, "//"):
			return nil
		}

		return &fancyerr.Error{
			Message: "expected end of statement",
			Primary: quickanno.Expected(p, pos, "a semicolon, EOL, or a line comment"),
		}
	}
}

func AndMustEOS() parser.WhitespaceFunc {
	return func(p *parser.Parser) *fancyerr.Error {
		parser.TrySkip(p, OrHorizontalWhitespace())
		if parser.TrySkipOk(p, AndEOS()) { // fast path
			return nil
		}

		start := p.Pos()
		parser.TokenWhile(p, func() bool {
			return !parser.MatchesWS(p, AndEOS())
		})
		p.CaptureError(&fancyerr.Error{
			Message: "end of statement: unexpected tokens",
			Primary: []fancyerr.Annotation{
				anno.Range(p.File, start, p.Pos(), "unexpected tokens, expected end of statement"),
			},
		})

		parser.TrySkip(p, AndEOS())
		return nil
	}
}

// AndEOL captures the comments until and including the first EOL.
func AndEOL() parser.WhitespaceFunc {
	return func(p *parser.Parser) *fancyerr.Error {
		for {
			parser.TrySkip(p, whitespace.Horizontal())
			c, hasComment := parser.TryOk(p, GeneralComment())
			if !hasComment {
				break
			}
			p.CaptureComment(&ast.CommentGroup{Comments: []*ast.Comment{c}})
		}

		c, hasComment := parser.TryOk(p, lineCommentWithoutEOL())
		if hasComment {
			p.CaptureComment(&ast.CommentGroup{Comments: []*ast.Comment{c}})
		}

		hasEOL := parser.TrySkipOk(p, whitespace.EOL())
		if hasEOL {
			return nil
		}

		return &fancyerr.Error{
			Message: "expected EOL",
			Primary: quickanno.Expected(p, p.Pos(), "the end of line, end of file, or a line comment"),
		}
	}
}

// OrAnyWhitespace parses and captures comments and any whitespace.
func OrAnyWhitespace() parser.WhitespaceFunc {
	return func(p *parser.Parser) *fancyerr.Error {
		pos := p.Pos()
		for {
			parser.TrySkip(p, whitespace.Horizontal())
			c, hasComment := parser.TryOk(p, GeneralComment())
			if !hasComment {
				break
			}
			p.CaptureComment(&ast.CommentGroup{Comments: []*ast.Comment{c}})
		}

		parser.TrySkip(p, whitespace.Horizontal())
		c, hasComment := parser.TryOk(p, LineComment())
		if hasComment {
			p.CaptureComment(&ast.CommentGroup{Comments: []*ast.Comment{c}})
		}

		parser.TrySkip(p, OrLoneWS())
		if pos == p.Pos() {
			return &fancyerr.Error{
				Message: "missing whitespace",
				Primary: quickanno.Expected(p, p.Pos(), "a space, tab, newline, or a comment"),
			}
		}
		return nil
	}
}

func OrLoneWS() parser.WhitespaceFunc {
	return func(p *parser.Parser) *fancyerr.Error {
		hasWS := parser.TrySkipOk(p, whitespace.Any())

		c, hasComment := parser.TryOk(p, GeneralComment())
		if hasComment {
			p.CaptureComment(&ast.CommentGroup{Comments: []*ast.Comment{c}})
			parser.TrySkip(p, OrAnyWhitespace())
			return nil
		}

		c, hasComment = parser.TryOk(p, LineComment())
		if !hasComment {
			if hasWS {
				return nil
			}
			return &fancyerr.Error{
				Message: "missing whitespace",
				Primary: quickanno.Expected(p, p.Pos(), "a space, tab, newline, or a comment"),
			}
		}

		cs := make([]*ast.Comment, 0, 48)
		for hasComment {
			cs = append(cs, c)

			parser.TrySkip(p, whitespace.Horizontal())
			c, hasComment = parser.TryOk(p, LineComment())
		}
		p.CaptureComment(&ast.CommentGroup{Comments: slices.Clip(cs)})
		parser.TrySkip(p, OrLoneWS())
		return nil
	}
}
