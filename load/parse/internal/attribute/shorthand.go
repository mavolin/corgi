package attribute

import (
	"slices"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/html/codepoint"
	"github.com/mavolin/corgi/v2/load/parse/internal/interpolation"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/whitespace"
)

func IDShorthand() parser.Func[*ast.IDShorthand] {
	return func(p *parser.Parser) *ast.IDShorthand {
		var s ast.IDShorthand

		s.Hash = parser.TryRuneAt(p, '#')
		if s.Hash == nil {
			return nil
		}

		s.ID = parser.Try(p, Shorthand())
		if s.ID == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message:  "id shorthand: missing id",
				Primary:  quickanno.Expected(p, p.Pos(), "an id"),
				Examples: []diagnostic.Example{{Example: "`#woof`"}},
			})
			return nil
		}
		return &s
	}
}

func ClassShorthand() parser.Func[*ast.ClassShorthand] {
	return func(p *parser.Parser) *ast.ClassShorthand {
		var s ast.ClassShorthand

		s.Dot = parser.TryRuneAt(p, '.')
		if s.Dot == nil {
			return nil
		}

		name0 := parser.Try(p, Shorthand())
		if name0 == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message:  "class shorthand: missing class name",
				Primary:  quickanno.Expected(p, p.Pos(), "a class name"),
				Examples: []diagnostic.Example{{Example: "`.woof`"}},
			})
			return nil
		}
		s.Names = make([]ast.Shorthand, 1, 16)
		s.Names[0] = name0

		for parser.TrySkip(p, whitespace.Horizontal()) {
			name := parser.Try(p, Shorthand())
			if name == nil {
				break
			}
			s.Names = append(s.Names, name)
		}
		s.Names = slices.Clip(s.Names)

		return &s
	}
}

func Shorthand() parser.Func[ast.Shorthand] {
	return func(p *parser.Parser) ast.Shorthand {
		s := parser.Collect(p, ShorthandNode(), 16, nil)
		if len(s) == 0 {
			return nil
		}
		return s
	}
}

func ShorthandNode() parser.Func[ast.ShorthandNode] {
	return func(p *parser.Parser) ast.ShorthandNode {
		if txt := parser.Try(p, ShorthandText()); txt != nil {
			return txt
		} else if interp := parser.Try(p, ShorthandInterpolation()); interp != nil {
			return interp
		}
		return nil
	}
}

func ShorthandText() parser.Func[*ast.ShorthandText] {
	return func(p *parser.Parser) *ast.ShorthandText {
		var txt ast.ShorthandText
		txt.Position = p.PosPtr()

		txt.Text = parser.TokenWhile(p, func() bool {
			return !parser.MatchesWS(p, whitespace.Any()) && !parser.MatchesAnyRune(p, '#', ',', ')') &&
				// https://html.spec.whatwg.org/multipage/common-microsyntaxes.html#set-of-space-separated-tokens
				!codepoint.MatchesAny(p, codepoint.ASCIIWhitespace) // includes FF
		})
		if txt.Text == "" {
			return nil
		}

		return &txt
	}
}

func ShorthandInterpolation() parser.Func[*ast.ShorthandInterpolation] {
	return func(p *parser.Parser) *ast.ShorthandInterpolation {
		interp := parser.Try(p, interpolation.ExpressionInterpolation())
		if interp == nil {
			if parser.MatchesToken(p, "#") {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "interpolation: missing opening brace",
					Primary: quickanno.Expected(p, quickanno.DeltaPos(p.Pos(), 0, 1), "a `{`"),
					Hints: []diagnostic.Hint{
						{
							Hint:    "If your class name or id, for some odd reason, contains a `#`, use a named attribute.",
							Example: "`div(id=\"woof#bark\", class=\"woof#bark\")`",
						},
					},
				})
				return (*ast.ShorthandInterpolation)(interp)
			}

			return nil
		}

		return (*ast.ShorthandInterpolation)(interp)
	}
}
