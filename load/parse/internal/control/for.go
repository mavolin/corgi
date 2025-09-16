package control

import (
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/body"
	"github.com/mavolin/corgi/v2/load/parse/internal/code"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/unexpected"
)

func For() parser.Func[*ast.For] {
	return func(p *parser.Parser) *ast.For {
		forKw := parser.TryKeywordAt(p, "for", comment.OrAnyWhitespace())
		if forKw == nil {
			return nil
		}

		var f ast.For
		f.For = forKw
		f.Header = parser.Try(p, ForHeader())

		err := unexpected.UntilAnyRune(p, comment.OrHorizontalWhitespace(), '{', '[')
		if err != nil {
			err.Message = "for loop: unexpected runes after header"
			p.CaptureError(err)
		}

		f.Body = parser.Try(p, body.Body())
		if f.Body == nil {
			return nil
		}
		return &f
	}
}

func ForHeader() parser.Func[ast.ForHeader] {
	return func(p *parser.Parser) ast.ForHeader {
		if frh := parser.Try(p, ForRangeHeader()); frh != nil {
			return frh
		} else if fch := parser.Try(p, ForClauseHeader()); fch != nil {
			return fch
		} else if fch := parser.Try(p, ForConditionHeader()); fch != nil {
			return fch
		}
		return nil
	}
}

func ForConditionHeader() parser.Func[*ast.ForConditionHeader] {
	return func(p *parser.Parser) *ast.ForConditionHeader {
		var h ast.ForConditionHeader

		h.Condition = parser.Try(p, code.Expression())
		if h.Condition == nil {
			return nil
		}

		return &h
	}
}

func ForClauseHeader() parser.Func[*ast.ForClauseHeader] {
	return func(p *parser.Parser) *ast.ForClauseHeader {
		if parser.MatchesAnyRune(p, '{', '[') {
			return nil
		}

		var h ast.ForClauseHeader

		h.Init = parser.TryOptional(p, code.SimpleStatement(), nil)
		if matches := parser.Try(p, comment.AndEOS()); !matches {
			return nil
		}
		parser.TrySkip(p, comment.OrAnyWhitespace())
		h.Condition = parser.TryOptional(p, code.Expression(), nil)
		if matches := parser.Try(p, comment.AndEOS()); !matches {
			return nil
		}
		parser.TrySkip(p, comment.OrAnyWhitespace())
		h.Post = parser.TryOptional(p, code.SimpleStatement(), nil)

		return &h
	}
}

func ForRangeHeader() parser.Func[*ast.ForRangeHeader] {
	return func(p *parser.Parser) *ast.ForRangeHeader {
		var h ast.ForRangeHeader

		pos := p.Pos()
		var comma *ast.Position
		if !parser.Matches(p, func(p *parser.Parser) bool {
			return parser.TryKeywordAt(p, "range", comment.OrAnyWhitespace()) != nil
		}) {
			h.Var1 = parser.TryOptional(p, code.Expression(), comment.OrHorizontalWhitespace())
			comma = parser.TryOptionalRuneAt(p, ',', comment.OrAnyWhitespace())
			if comma != nil {
				if h.Var1 == nil {
					p.CaptureError(&diagnostic.Diagnostic{
						Message: "for-range header: missing first variable",
						Primary: quickanno.Expected(p, pos, "a variable"),
					})
				}

				h.Var2 = parser.TryOptional(p, code.Expression(), comment.OrHorizontalWhitespace())
				if h.Var2 == nil {
					p.CaptureError(&diagnostic.Diagnostic{
						Message: "for-range header: missing second variable",
						Primary: quickanno.Expected(p, p.Pos(), "a variable"),
						Secondary: []diagnostic.Annotation{
							anno.Position(p.File, *comma, "because of the comma here"),
						},
					})
				}
			}
		}

		h.Colon = parser.TryOptionalRuneAt(p, ':', nil)
		h.EqualSign = parser.TryOptionalRuneAt(p, '=', comment.OrAnyWhitespace())
		if h.Colon != nil && h.EqualSign == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "for-range header: missing equal sign",
				Primary: quickanno.Expected(p, p.Pos(), "an equal sign"),
				Secondary: []diagnostic.Annotation{
					anno.Position(p.File, *h.Colon, "because of the colon here"),
				},
			})
		}
		if h.Var1 == nil && comma == nil && h.EqualSign != nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "for-range header: missing first variable",
				Primary: quickanno.Expected(p, pos, "a variable"),
				Secondary: []diagnostic.Annotation{
					anno.Position(p.File, *h.EqualSign, "because of the equal sign here"),
				},
				Hints: []diagnostic.Hint{
					{Hint: "If you don't want to declare/set any variables, remove the equal sign."},
				},
			})
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())
		h.Range = parser.TryKeywordAt(p, "range", comment.OrAnyWhitespace())
		if h.Range == nil {
			return nil
		}

		h.Expression = parser.Try(p, code.Expression())
		if h.Expression == nil {
			return nil
		}

		return &h
	}
}
