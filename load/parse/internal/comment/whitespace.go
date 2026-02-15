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

// AndEOS matches the end of statement, optionally preceded by general comments.
// It consumes trailing horizontal whitespace; it doesn't consume the EOL.
// If the line ends with a line comment, AndEOS accepts, but does not consume
// it.
func AndEOS() parser.Func[bool] {
	return func(p *parser.Parser) bool {
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
			return true
		case parser.MatchesWS(p, whitespace.EOL()):
			return true
		case parser.MatchesToken(p, "}"):
			return true
		case parser.MatchesToken(p, "//"):
			return true
		}

		return false
	}
}

// AndMustEOS matches the end of statement, optionally preceded by general
// comments.
// If not at the end of statement, it captures an error and returns at the
// current position.
func AndMustEOS() parser.Func[bool] {
	return func(p *parser.Parser) bool {
		matches := parser.Try(p, AndEOS())
		if !matches {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "expected end of statement",
				Primary: quickanno.Expected(p, p.Pos(), "a semicolon, EOL, or a line comment"),
			})
		}
		return true
	}
}

// AndForceEOS guarantees that the end of statement is reached.
// If it encounters unexpected tokens before the end of statement, it consumes
// them, captures an error, and then continues to parse the end of statement.
func AndForceEOS() parser.Func[bool] {
	return func(p *parser.Parser) bool {
		parser.TrySkip(p, OrHorizontalWhitespace())
		if matches := parser.TryOptional(p, AndEOS(), nil); matches { // fast path
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
		return hasEOL
	}
}

// OrAnyWhitespace parses and captures comments and any whitespace.
func OrAnyWhitespace() parser.WhitespaceFunc {
	return func(p *parser.Parser) bool {
		start := p.ByteIndex()
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

		parser.TrySkip(p, orLoneWhitespace())
		return start != p.ByteIndex()
	}
}

// orLoneWhitespace must be called before any non-comment node on a line.
// It consumes any whitespace and captures all comments until the next node or
// the EOF.
func orLoneWhitespace() parser.WhitespaceFunc {
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
			return hasWS
		}

		var cs []*ast.Comment
		for c != nil {
			cs = append(cs, c)

			parser.TrySkip(p, whitespace.Horizontal())
			c = parser.Try(p, LineComment())
		}
		p.CaptureComment(&ast.CommentGroup{Comments: slices.Clip(cs)})
		parser.TrySkip(p, orLoneWhitespace())
		return true
	}
}
