package comment

import (
	"slices"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/whitespace"
)

// OrHorizontalWhitespace parses and captures inline comments and horizontal
// whitespace.
func OrHorizontalWhitespace() parser.WhitespaceFunc {
	return func(p *parser.Parser) bool {
		hasWS := parser.TrySkip(p, whitespace.Horizontal())
		c := parser.Try(p, GeneralComment())
		if !hasWS && c == nil {
			return false
		}
		if c != nil {
			p.CaptureComment(&ast.CommentGroup{Comments: []*ast.Comment{c}})
		}

		for hasWS || c != nil {
			hasWS = parser.TrySkip(p, whitespace.Horizontal())
			c = parser.Try(p, GeneralComment())
			if c != nil {
				p.CaptureComment(&ast.CommentGroup{Comments: []*ast.Comment{c}})
			}
		}
		return true
	}
}

// AndEOS matches the end of statement, optionally preceded by block comments.
// It consumes trailing horizontal whitespace; it doesn't consume the EOL.
// If the line ends with a line comment, AndEOS accepts, but does not consume
// it.
func AndEOS() parser.Func[struct{}] {
	return func(p *parser.Parser) (struct{}, *diagnostic.Diagnostic) {
		pos := p.Pos()
		for {
			parser.TrySkip(p, whitespace.Horizontal())
			c := parser.TryOptional(p, GeneralComment(), nil)
			if c == nil {
				break
			}
			p.CaptureComment(&ast.CommentGroup{Comments: []*ast.Comment{c}})
		}

		switch {
		case parser.TryOptionalRune(p, ';', nil):
			return struct{}{}, nil
		case parser.MatchesWS(p, whitespace.EOL()):
			return struct{}{}, nil
		case parser.MatchesToken(p, "}"):
			return struct{}{}, nil
		case parser.MatchesToken(p, "//"):
			return struct{}{}, nil
		}

		return struct{}{}, &diagnostic.Diagnostic{
			Message: "expected end of statement",
			Primary: quickanno.Expected(p, pos, "a semicolon, EOL, or a line comment"),
		}
	}
}

func AndForceEOS() parser.WhitespaceFunc {
	return func(p *parser.Parser) bool {
		parser.TrySkip(p, OrHorizontalWhitespace())
		if _, err := parser.TryOptionalErr(p, AndEOS(), nil); err == nil { // fast path
			return true
		}

		start := p.Pos()
		parser.TokenWhile(p, func() bool {
			return !parser.Matches(p, AndEOS())
		})
		p.CaptureError(&diagnostic.Diagnostic{
			Message: "end of statement: unexpected tokens",
			Primary: []diagnostic.Annotation{
				anno.Range(p.File, start, p.Pos(), "unexpected tokens, expected end of statement"),
			},
		})

		parser.Try(p, AndEOS())
		return true
	}
}

// AndEOL captures the comments until and including the first EOL.
func AndEOL() parser.WhitespaceFunc {
	return func(p *parser.Parser) bool {
		for {
			parser.TrySkip(p, whitespace.Horizontal())
			c := parser.Try(p, GeneralComment())
			if c == nil {
				break
			}
			p.CaptureComment(&ast.CommentGroup{Comments: []*ast.Comment{c}})
		}

		c := parser.Try(p, lineCommentWithoutEOL())
		if c != nil {
			p.CaptureComment(&ast.CommentGroup{Comments: []*ast.Comment{c}})
		}

		hasEOL := parser.TrySkip(p, whitespace.EOL())
		if hasEOL {
			return true
		}

		return false
	}
}

// OrAnyWhitespace parses and captures comments and any whitespace.
func OrAnyWhitespace() parser.WhitespaceFunc {
	return func(p *parser.Parser) bool {
		pos := p.Pos()
		for {
			parser.TrySkip(p, whitespace.Horizontal())
			c := parser.Try(p, GeneralComment())
			if c == nil {
				break
			}
			p.CaptureComment(&ast.CommentGroup{Comments: []*ast.Comment{c}})
		}

		parser.TrySkip(p, whitespace.Horizontal())
		c := parser.Try(p, LineComment())
		if c != nil {
			p.CaptureComment(&ast.CommentGroup{Comments: []*ast.Comment{c}})
		}

		parser.TrySkip(p, OrLoneWS())
		if pos == p.Pos() {
			return false
		}
		return true
	}
}

func OrLoneWS() parser.WhitespaceFunc {
	return func(p *parser.Parser) bool {
		hasWS := parser.TrySkip(p, whitespace.Any())

		c := parser.Try(p, GeneralComment())
		if c != nil {
			p.CaptureComment(&ast.CommentGroup{Comments: []*ast.Comment{c}})
			parser.TrySkip(p, OrAnyWhitespace())
			return true
		}

		c = parser.Try(p, LineComment())
		if c == nil {
			if hasWS {
				return true
			}
			return false
		}

		cs := make([]*ast.Comment, 0, 48)
		for c != nil {
			cs = append(cs, c)

			parser.TrySkip(p, whitespace.Horizontal())
			c = parser.Try(p, LineComment())
		}
		p.CaptureComment(&ast.CommentGroup{Comments: slices.Clip(cs)})
		parser.TrySkip(p, OrLoneWS())
		return true
	}
}
