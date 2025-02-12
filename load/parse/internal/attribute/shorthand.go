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
	return func(p *parser.Parser) (*ast.IDShorthand, *diagnostic.Diagnostic) {
		var s ast.IDShorthand

		s.Hash = parser.TryRuneAt(p, '#')
		if s.Hash == nil {
			return nil, &diagnostic.Diagnostic{
				Message:  "missing id shorthand",
				Primary:  quickanno.Expected(p, p.Pos(), "an id shorthand"),
				Examples: []diagnostic.Example{{Example: "`#woof`"}},
			}
		}

		s.ID = parser.Must(p, Shorthand())
		return &s, nil
	}
}

func ClassShorthand() parser.Func[*ast.ClassShorthand] {
	return func(p *parser.Parser) (*ast.ClassShorthand, *diagnostic.Diagnostic) {
		var s ast.ClassShorthand

		s.Dot = parser.TryRuneAt(p, '.')
		if s.Dot == nil {
			return nil, &diagnostic.Diagnostic{
				Message:  "missing class shorthand",
				Primary:  quickanno.Expected(p, p.Pos(), "a class shorthand"),
				Examples: []diagnostic.Example{{Example: "`.woof`"}},
			}
		}

		s.Names = make([]ast.Shorthand, 1, 16)
		var err *diagnostic.Diagnostic
		s.Names[0], err = parser.TryErr(p, Shorthand())
		if err != nil {
			p.CaptureError(err)
		}

		for parser.TrySkip(p, whitespace.Horizontal()) {
			name := parser.Try(p, Shorthand())
			if name == nil {
				break
			}
			s.Names = append(s.Names, name)
		}
		s.Names = slices.Clip(s.Names)

		return &s, nil
	}
}

func Shorthand() parser.Func[ast.Shorthand] {
	return func(p *parser.Parser) (ast.Shorthand, *diagnostic.Diagnostic) {
		s := parser.Collect(p, ShorthandNode(), 16, nil)
		if len(s) == 0 {
			return nil, &diagnostic.Diagnostic{
				Message: "missing shorthand name",
				Primary: quickanno.Expected(p, p.Pos(), "text or interpolation"),
				Examples: []diagnostic.Example{
					{Title: "just text", Example: "`.woof` or `#bark`"},
					{Title: "with interpolation", Example: "`.button--#{size}` or `#button-#{i}`"},
				},
			}
		}
		return s, nil
	}
}

func ShorthandNode() parser.Func[ast.ShorthandNode] {
	return func(p *parser.Parser) (ast.ShorthandNode, *diagnostic.Diagnostic) {
		if txt := parser.Try(p, ShorthandText()); txt != nil {
			return txt, nil
		} else if interp := parser.Try(p, ShorthandInterpolation()); interp != nil {
			return interp, nil
		}
		return nil, &diagnostic.Diagnostic{
			Message: "missing shorthand node",
			Primary: quickanno.Expected(p, p.Pos(), "shorthand text or interpolation"),
			Examples: []diagnostic.Example{
				{Title: "text", Example: "`.woof`"},
				{Title: "interpolation", Example: "`#{bark}`"},
			},
		}
	}
}

func ShorthandText() parser.Func[*ast.ShorthandText] {
	return func(p *parser.Parser) (*ast.ShorthandText, *diagnostic.Diagnostic) {
		var txt ast.ShorthandText
		txt.Position = p.PosPtr()

		txt.Text = parser.TokenWhile(p, func() bool {
			return !parser.MatchesWS(p, whitespace.Any()) && !parser.MatchesAnyRune(p, '#', ',', ')') &&
				// https://html.spec.whatwg.org/multipage/common-microsyntaxes.html#set-of-space-separated-tokens
				!codepoint.MatchesAny(p, codepoint.ASCIIWhitespace) // includes FF
		})
		if txt.Text == "" {
			return nil, &diagnostic.Diagnostic{
				Message: "missing shorthand text",
				Primary: quickanno.Expected(p, p.Pos(), "text, but not interpolation"),
			}
		}

		return &txt, nil
	}
}

func ShorthandInterpolation() parser.Func[*ast.ShorthandInterpolation] {
	return func(p *parser.Parser) (*ast.ShorthandInterpolation, *diagnostic.Diagnostic) {
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
				return (*ast.ShorthandInterpolation)(interp), nil
			}

			return nil, &diagnostic.Diagnostic{
				Message: "missing interpolation",
				Primary: quickanno.Expected(p, p.Pos(), "an interpolation"),
			}
		}

		return (*ast.ShorthandInterpolation)(interp), nil
	}
}
