package attribute

import (
	"slices"
	"unicode"
	"unicode/utf8"

	"github.com/mavolin/corgi/v2/fancyerr"
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/code"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/html/codepoint"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/whitespace"
)

func IDShorthand() parser.Func[*ast.IDShorthand] {
	return func(p *parser.Parser) (*ast.IDShorthand, *fancyerr.Error) {
		s := &ast.IDShorthand{Position: p.Pos()}
		if !parser.TryRune(p, '#') {
			return nil, &fancyerr.Error{
				Message:  "missing id shorthand",
				Primary:  quickanno.Expected(p, s.Position, "an id shorthand"),
				Examples: []fancyerr.Example{{Example: "`#woof`"}},
			}
		}

		s.ID = parser.Must(p, Shorthand())
		return s, nil
	}
}

func ClassShorthand() parser.Func[*ast.ClassShorthand] {
	return func(p *parser.Parser) (*ast.ClassShorthand, *fancyerr.Error) {
		s := &ast.ClassShorthand{Position: p.Pos()}
		if !parser.TryRune(p, '.') {
			return nil, &fancyerr.Error{
				Message:  "missing class shorthand",
				Primary:  quickanno.Expected(p, s.Position, "a class shorthand"),
				Examples: []fancyerr.Example{{Example: "`.woof`"}},
			}
		}

		s.Name = parser.Must(p, Shorthand())

		suffixes := make([]ast.ClassSuffix, 0, 16)
		for {
			if !parser.TrySkipOk(p, comment.OrHorizontalWhitespace()) {
				parser.RestoreWS(p)
			}

			suffix, ok := parser.TryOk(p, ClassSuffix())
			if !ok {
				break
			}
			suffixes = append(suffixes, suffix)
		}
		s.Suffixes = slices.Clip(suffixes)

		if len(s.Suffixes) > 0 && len(s.Name) > 0 {
			lastShorthand := s.Name[len(s.Name)-1]
			if txt, ok := lastShorthand.(*ast.ShorthandText); ok && txt.Text != "" {
				r, _ := utf8.DecodeLastRuneInString(txt.Text)
				s.IsPrefix = unicode.IsPunct(r)
			}
		}

		return s, nil
	}
}

func ClassSuffix() parser.Func[ast.ClassSuffix] {
	return func(p *parser.Parser) (ast.ClassSuffix, *fancyerr.Error) {
		s, ok := parser.TryOk(p, Shorthand())
		if !ok {
			return nil, &fancyerr.Error{
				Message:  "missing class suffix",
				Primary:  quickanno.Expected(p, p.Pos(), "a class suffix"),
				Examples: []fancyerr.Example{{Example: "`.button --blue --small`"}},
			}
		}

		return ast.ClassSuffix(s), nil
	}
}

func Shorthand() parser.Func[ast.Shorthand] {
	return func(p *parser.Parser) (ast.Shorthand, *fancyerr.Error) {
		nodes := make([]ast.ShorthandNode, 1, 16)

		var ok bool
		nodes[0], ok = parser.TryOk(p, ShorthandNode())
		if !ok {
			return nil, &fancyerr.Error{
				Message: "missing shorthand name",
				Primary: quickanno.Expected(p, p.Pos(), "text or interpolation"),
				Examples: []fancyerr.Example{
					{Title: "just text", Example: "`.woof`"},
					{Title: "with interpolation", Example: "`.button--#{size}`"},
				},
			}
		}

		for {
			n, ok := parser.TryOk(p, ShorthandNode())
			if !ok {
				return slices.Clip(nodes), nil
			}
			nodes = append(nodes, n)
		}
	}
}

func ShorthandNode() parser.Func[ast.ShorthandNode] {
	return func(p *parser.Parser) (ast.ShorthandNode, *fancyerr.Error) {
		if txt, ok := parser.TryOk(p, ShorthandText()); ok {
			return txt, nil
		} else if interp, ok := parser.TryOk(p, ShorthandInterpolation()); ok {
			return interp, nil
		}
		return nil, &fancyerr.Error{
			Message: "missing shorthand node",
			Primary: quickanno.Expected(p, p.Pos(), "shorthand text or interpolation"),
			Examples: []fancyerr.Example{
				{Title: "text", Example: "`.woof`"},
				{Title: "interpolation", Example: "`#{bark}`"},
			},
		}
	}
}

func ShorthandText() parser.Func[*ast.ShorthandText] {
	return func(p *parser.Parser) (*ast.ShorthandText, *fancyerr.Error) {
		txt := &ast.ShorthandText{Position: p.Pos()}
		txt.Text = parser.TokenWhile(p, func() bool {
			return !parser.MatchesWS(p, whitespace.Any()) && !parser.MatchesToken(p, "#") &&
				// https://html.spec.whatwg.org/multipage/common-microsyntaxes.html#set-of-space-separated-tokens
				!codepoint.MatchesAny(p, codepoint.ASCIIWhitespace) // includes FF
		})
		if txt.Text == "" {
			return nil, &fancyerr.Error{
				Message: "missing shorthand text",
				Primary: quickanno.Expected(p, p.Pos(), "text, but not interpolation"),
			}
		}

		return txt, nil
	}
}

func ShorthandInterpolation() parser.Func[*ast.ShorthandInterpolation] {
	return func(p *parser.Parser) (*ast.ShorthandInterpolation, *fancyerr.Error) {
		interp, ok := parser.TryOk(p, code.ExpressionInterpolation())
		if !ok {
			if parser.MatchesToken(p, "#") {
				p.CaptureError(&fancyerr.Error{
					Message: "interpolation: missing opening brace",
					Primary: quickanno.Expected(p, quickanno.DeltaPos(p.Pos(), 0, 1), "a `{`"),
					Hints: []fancyerr.Hint{
						{
							Hint:    "If your class name or id, for some odd reason, contains a `#`, use a named attribute.",
							Example: "`div(id=\"woof#bark\", class=\"woof#bark\")`",
						},
					},
				})
				return (*ast.ShorthandInterpolation)(interp), nil
			}

			return nil, &fancyerr.Error{
				Message: "missing interpolation",
				Primary: quickanno.Expected(p, p.Pos(), "an interpolation"),
			}
		}

		return (*ast.ShorthandInterpolation)(interp), nil
	}
}
