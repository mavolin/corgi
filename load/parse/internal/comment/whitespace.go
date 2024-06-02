package comment

import (
	"github.com/mavolin/corgi/fancyerr"
	"github.com/mavolin/corgi/file/ast"
	parser "github.com/mavolin/corgi/load/parse/internal"
	"github.com/mavolin/corgi/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/load/parse/internal/whitespace"
)

// OrHorizontalWhitespace parses and captures inline comments and horizontal
// whitespace.
func OrHorizontalWhitespace() parser.Func[struct{}] {
	return func(p *parser.Parser) (struct{}, *fancyerr.Error) {
		_, hasWS := parser.Try(p, whitespace.Horizontal())
		c, hasComment := parser.Try(p, InlineGroup())
		if !hasWS && !hasComment {
			return struct{}{}, &fancyerr.Error{
				Message: "missing horizontal whitespace",
				Primary: quickanno.Expected(p, p.Pos(), "a space, tab, or a single-line block comment"),
			}
		}
		if hasComment {
			p.CaptureComment(c)
		}

		for hasWS || hasComment {
			_, hasWS = parser.Try(p, whitespace.Horizontal())
			c, hasComment = parser.Try(p, InlineGroup())
			if hasComment {
				p.CaptureComment(c)
			}
		}
		return struct{}{}, nil
	}
}

// AndEOS matches the end of statement, optionally preceded by block comments.
// It consumes trailing horizontal whitespace; it doesn't consume the EOL.
// If the line ends with a line comment, AndEOS accepts, but does not consume
// it.
func AndEOS() parser.Func[struct{}] {
	return func(p *parser.Parser) (struct{}, *fancyerr.Error) {
		pos := p.Pos()
		for {
			_, hasWS := parser.Try(p, whitespace.Horizontal())
			c, hasComment := parser.Try(p, singleLineBlockComment())
			if !hasComment && !hasWS {
				break
			}
			if hasComment {
				p.CaptureComment(&ast.CommentGroup{Comments: []*ast.Comment{c}})
			}
		}

		switch {
		case parser.TryRune(p, ';'):
			return struct{}{}, nil
		case parser.Matches(p, whitespace.EOL()):
			return struct{}{}, nil
		case parser.MatchesToken(p, "}"):
			return struct{}{}, nil
		case parser.MatchesToken(p, "//"):
			return struct{}{}, nil
		}

		state := p.CloneState()
		if !parser.TryToken(p, "/*") {
			return struct{}{}, &fancyerr.Error{
				Message: "expected end of statement",
				Primary: quickanno.Expected(p, pos, "a semicolon, EOL, or a line comment"),
			}
		}

		parser.TokenWhile(p, func() bool {
			return !parser.MatchesToken(p, "*/") && parser.Matches(p, whitespace.EOL())
		})
		if !parser.MatchesToken(p, "*/") { // IS multi-line
			p.RestoreState(state)
			return struct{}{}, nil
		}

		return struct{}{}, &fancyerr.Error{
			Message: "expected end of statement",
			Primary: quickanno.Expected(p, pos, "a semicolon, EOL, or a line comment"),
		}
	}
}

// OrEOL captures the comments until and including the first EOL.
// Only single-line block comments are captured.
func OrEOL() parser.Func[struct{}] {
	return func(p *parser.Parser) (struct{}, *fancyerr.Error) {
		for {
			parser.Try(p, whitespace.Horizontal())
			c, ok := parser.Try(p, singleLineBlockComment())
			if !ok {
				break
			}
			p.CaptureComment(&ast.CommentGroup{Comments: []*ast.Comment{c}})
		}

		c, ok := parser.Try(p, lineCommentWithoutEOL())
		if ok {
			p.CaptureComment(&ast.CommentGroup{Comments: []*ast.Comment{c}})
		}

		_, ok = parser.Try(p, whitespace.EOL())
		if ok {
			return struct{}{}, nil
		}

		return struct{}{}, &fancyerr.Error{
			Message: "expected EOL",
			Primary: quickanno.Expected(p, p.Pos(), "the end of line, end of file, or a line comment"),
		}
	}
}

// OrAnyWhitespace parses and captures comments and any whitespace.
func OrAnyWhitespace() parser.Func[struct{}] {
	return func(p *parser.Parser) (struct{}, *fancyerr.Error) {
		pos := p.Pos()
		for {
			parser.Try(p, whitespace.Horizontal())
			c, ok := parser.Try(p, singleLineBlockComment())
			if !ok {
				break
			}
			p.CaptureComment(&ast.CommentGroup{Comments: []*ast.Comment{c}})
		}

		c, ok := parser.Try(p, lineCommentWithoutEOL())
		if ok {
			p.CaptureComment(&ast.CommentGroup{Comments: []*ast.Comment{c}})
		}

		_, ok = parser.Try(p, whitespace.EOL())
		if !ok {
			if pos == p.Pos() {
				return struct{}{}, &fancyerr.Error{
					Message: "missing whitespace",
					Primary: quickanno.Expected(p, p.Pos(), "a space, tab, newline, or a comment"),
				}
			}

			return struct{}{}, nil
		}

		parser.Try(p, OrLoneWS())
		return struct{}{}, nil
	}
}

func OrLoneWS() parser.Func[struct{}] {
	return func(p *parser.Parser) (struct{}, *fancyerr.Error) {
		_, hasWS := parser.Try(p, whitespace.Any())
		g, hasComment := parser.Try(p, LoneGroup())
		if !hasWS && !hasComment {
			return struct{}{}, &fancyerr.Error{
				Message: "missing whitespace",
				Primary: quickanno.Expected(p, p.Pos(), "a space, tab, newline, or a comment"),
			}
		}

		for hasWS || hasComment {
			if hasComment {
				p.CaptureComment(g)

				if g.Comments[0].Block {
					if _, ok := parser.Try(p, OrEOL()); !ok {
						break // we've reached EOF
					}
				}
			}

			_, hasWS = parser.Try(p, whitespace.Any())
			g, hasComment = parser.Try(p, LoneGroup())
		}

		return struct{}{}, nil
	}
}
