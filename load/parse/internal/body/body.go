package body

import (
	"github.com/mavolin/corgi/v2/fancyerr"
	"github.com/mavolin/corgi/v2/fancyerr/anno"
	"github.com/mavolin/corgi/v2/file/ast"
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
	return func(p *parser.Parser) (ast.Body, *fancyerr.Error) {
		if s := parser.Try(p, Scope()); s != nil {
			return s, nil
		} else if b := parser.Try(p, BracketText()); b != nil {
			return b, nil
		} else if s := parser.Try(p, UnderscoreBlockShorthand()); s != nil {
			if !allowUnderscoreBlockShorthand {
				p.CaptureError(&fancyerr.Error{
					Message: "underscore block shorthand not allowed here",
					Primary: []fancyerr.Annotation{
						anno.Position(p.File, s.Start(), "cannot place an underscore block shorthand here"),
					},
					Explanation: "Underscore block shorthands can only be used as the body for component calls.",
				})
			}
			return s, nil
		}

		examples := make([]fancyerr.Example, 0, 3)
		examples = append(examples,
			fancyerr.Example{Title: "scope", Example: "{ :fmt.Number(val: 21_000) }"},
			fancyerr.Example{Title: "bracket text", Example: "[ Hello, World! ]"})
		if allowUnderscoreBlockShorthand {
			examples = append(examples, fancyerr.Example{Title: "underscore block shorthand", Example: "_{ ... }"})
		}

		return nil, &fancyerr.Error{
			Message:  "missing body",
			Primary:  quickanno.Expected(p, p.Pos(), "a body"),
			Examples: examples,
		}
	}
}

func UnderscoreBlockShorthand() parser.Func[*ast.UnderscoreBlockShorthand] {
	return func(p *parser.Parser) (*ast.UnderscoreBlockShorthand, *fancyerr.Error) {
		var s ast.UnderscoreBlockShorthand
		s.Position = p.PosPtr()

		if !parser.TryRune(p, '_') {
			return nil, &fancyerr.Error{
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
			p.CaptureError(&fancyerr.Error{
				Message: "underscore block shorthand: found multiple underscores",
				Primary: []fancyerr.Annotation{
					anno.Range(p.File, excessUnderscoreStart, excessUnderscoreEnd, "remove these excess underscores"),
				},
			})
		}

		wsStart := p.Pos()
		hasWS := parser.TrySkip(p, comment.OrHorizontalWhitespace())
		wsEnd := p.Pos()

		s.Body = parser.Try(p, Body())
		if s.Body == nil {
			return nil, &fancyerr.Error{
				Message: "missing underscore block shorthand: missing body",
				Primary: quickanno.Expected(p, p.Pos(), "a body"),
			}
		} else if hasWS {
			p.CaptureError(&fancyerr.Error{
				Message: "underscore block shorthand: unexpected whitespace between `_` and body",
				Primary: []fancyerr.Annotation{
					anno.Range(p.File, wsStart, wsEnd, "this whitespace is not allowed here"),
				},
			})
		}

		return &s, nil
	}
}
