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

func body(allowUnderscoreBlockShorthand bool) parser.Func[ast.Body] {
	return func(p *parser.Parser) (ast.Body, *diagnostic.Diagnostic) {
		if s := parser.Try(p, Scope()); s != nil {
			return s, nil
		} else if b := parser.Try(p, BracketText()); b != nil {
			return b, nil
		} else if s := parser.Try(p, UnderscoreBlockShorthand()); s != nil {
			if !allowUnderscoreBlockShorthand {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "underscore block shorthand not allowed here",
					Primary: []diagnostic.Annotation{
						anno.Position(p.File, s.Start(), "cannot place an underscore block shorthand here"),
					},
					Explanation: "Underscore block shorthands can only be used as the body for component calls.",
				})
			}
			return s, nil
		}

		examples := make([]diagnostic.Example, 0, 3)
		examples = append(examples,
			diagnostic.Example{Title: "scope", Example: "{ :fmt.Number(val: 21_000) }"},
			diagnostic.Example{Title: "bracket text", Example: "[ Hello, World! ]"})
		if allowUnderscoreBlockShorthand {
			examples = append(examples, diagnostic.Example{Title: "underscore block shorthand", Example: "_{ ... }"})
		}

		return nil, &diagnostic.Diagnostic{
			Message:  "missing body",
			Primary:  quickanno.Expected(p, p.Pos(), "a body"),
			Examples: examples,
		}
	}
}

func UnderscoreBlockShorthand() parser.Func[*ast.UnderscoreBlockShorthand] {
	return func(p *parser.Parser) (*ast.UnderscoreBlockShorthand, *diagnostic.Diagnostic) {
		var s ast.UnderscoreBlockShorthand
		s.Position = p.PosPtr()

		if !parser.TryRune(p, '_') {
			return nil, &diagnostic.Diagnostic{
				Message: "missing underscore block shorthand",
				Primary: quickanno.Expected(p, p.Pos(), "an underscore block shorthand (`_{ ... }` or `_[ ... ]`)"),
			}
		}

		// Handle the special case of a nested underscore block shorthand, i.e.
		// __{ ... }, because the error message of body(false) might be misleading.
		excessUnderscoreStart := p.Pos()
		for parser.TryRune(p, '_') {
		}
		excessUnderscoreEnd := p.Pos()
		if excessUnderscoreStart != excessUnderscoreEnd {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "underscore block shorthand: found multiple underscores",
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
				Message: "missing underscore block shorthand: missing body",
				Primary: quickanno.Expected(p, p.Pos(), "a body"),
			}
		} else if hasWS {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "underscore block shorthand: unexpected whitespace between `_` and body",
				Primary: []diagnostic.Annotation{
					anno.Range(p.File, wsStart, wsEnd, "this whitespace is not allowed here"),
				},
			})
		}

		return &s, nil
	}
}
