package body

import (
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

func Body() parser.Func[ast.Body] {
	return body(false)
}

func ComponentCallBody() parser.Func[ast.Body] {
	return body(true)
}

func body(allowDefaultBlockShorthand bool) parser.Func[ast.Body] {
	return func(p *parser.Parser) (ast.Body, *diagnostic.Diagnostic) {
		if s := parser.Try(p, Scope()); s != nil {
			return s, nil
		} else if b := parser.Try(p, BracketText()); b != nil {
			return b, nil
		} else if s := parser.Try(p, DefaultBlockShorthand()); s != nil {
			if !allowDefaultBlockShorthand {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "default block shorthand not allowed here",
					Primary: []diagnostic.Annotation{
						anno.Position(p.File, s.Start(), "cannot place a default block shorthand here"),
					},
					Explanation: "Default block shorthands can only be used as the body for component calls.",
				})
			}
			return s, nil
		}

		examples := make([]diagnostic.Example, 0, 3)
		examples = append(examples,
			diagnostic.Example{Title: "scope", Example: "{ :fmt.Number(val: 21_000) }"},
			diagnostic.Example{Title: "bracket text", Example: "[ Hello, World! ]"})
		if allowDefaultBlockShorthand {
			examples = append(examples, diagnostic.Example{Title: "default block shorthand", Example: "_{ ... }"})
		}

		return nil, &diagnostic.Diagnostic{
			Message:  "missing body",
			Primary:  quickanno.Expected(p, p.Pos(), "a body"),
			Examples: examples,
		}
	}
}

func DefaultBlockShorthand() parser.Func[*ast.DefaultBlockShorthand] {
	return func(p *parser.Parser) (*ast.DefaultBlockShorthand, *diagnostic.Diagnostic) {
		var s ast.DefaultBlockShorthand
		s.Position = p.PosPtr()

		if !parser.TryRune(p, '_') {
			return nil, &diagnostic.Diagnostic{
				Message: "missing default block shorthand",
				Primary: quickanno.Expected(p, p.Pos(), "a default block shorthand (`_{ ... }` or `_[ ... ]`)"),
			}
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

		s.Body = parser.Try(p, Body())
		if s.Body == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "missing default block shorthand: missing body",
				Primary: quickanno.Expected(p, p.Pos(), "a body"),
			}
		} else if hasWS {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "default block shorthand: unexpected whitespace between `_` and body",
				Primary: []diagnostic.Annotation{
					anno.Range(p.File, wsStart, wsEnd, "this whitespace is not allowed here"),
				},
			})
		}

		return &s, nil
	}
}
