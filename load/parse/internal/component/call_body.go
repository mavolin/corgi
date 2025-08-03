package component

import (
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/body"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
)

func CallBody() parser.Func[ast.ComponentCallBody] {
	return func(p *parser.Parser) ast.ComponentCallBody {
		if s := parser.Try(p, DefaultBlockShorthand()); s != nil {
			return s
		} else if s := parser.Try(p, body.Scope()); s != nil {
			return s
		} else if b := parser.Try(p, body.Body()); b != nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "component call: illegal body",
				Primary: []diagnostic.Annotation{
					anno.Node(p.File, b, "expected a scope or a default block shorthand"),
				},
			})
		}

		return nil
	}
}

func DefaultBlockShorthand() parser.Func[*ast.DefaultBlockShorthand] {
	return func(p *parser.Parser) *ast.DefaultBlockShorthand {
		var s ast.DefaultBlockShorthand
		s.Position = p.PosPtr()

		if !parser.TryRune(p, '_') {
			return nil
		}

		// Handle the special case of a nested default block shorthand, i.e.
		// __{ ... }, because the error message of body(false) might be misleading.
		excessUnderscoreStart := p.Pos()
		for parser.TryRune(p, '_') {
		}
		excessUnderscoreEnd := p.Pos()
		if excessUnderscoreStart != excessUnderscoreEnd {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "default block shorthand: found multiple underscores",
				Primary: []diagnostic.Annotation{
					anno.Range(p.File, excessUnderscoreStart, excessUnderscoreEnd, "remove these excess underscores"),
				},
			})
		}

		wsStart := p.Pos()
		hasWS := parser.TrySkip(p, comment.OrHorizontalWhitespace())
		wsEnd := p.Pos()

		s.Body = parser.Try(p, body.Body())
		if s.Body == nil {
			return nil
		} else if hasWS {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "default block shorthand: unexpected whitespace between `_` and body",
				Primary: []diagnostic.Annotation{
					anno.Range(p.File, wsStart, wsEnd, "this whitespace is not allowed here"),
				},
			})
		}

		return &s
	}
}
