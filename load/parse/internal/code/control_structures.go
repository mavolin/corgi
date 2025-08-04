package code

import (
	"slices"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/body"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/unexpected"
)

func Conditional() parser.Func[*ast.Conditional] {
	return func(p *parser.Parser) *ast.Conditional {
		ifNode := parser.Try(p, If())
		if ifNode == nil {
			return nil
		}

		var c ast.Conditional
		c.If = ifNode
		c.ElseIfs = parser.Collect(p, ElseIf(), comment.OrAnyWhitespace())
		parser.TrySkip(p, comment.OrAnyWhitespace())
		c.Else = parser.Try(p, Else())

		return &c
	}
}

func If() parser.Func[*ast.If] {
	return func(p *parser.Parser) *ast.If {
		ifKw := parser.TryKeywordAt(p, "if", comment.OrAnyWhitespace())
		if ifKw == nil {
			return nil
		}

		var i ast.If
		i.If = ifKw
		i.Header = parser.Try(p, IfHeader())
		if i.Header == nil {
			return nil
		}
		parser.TrySkip(p, comment.OrHorizontalWhitespace())
		i.Then = parser.Try(p, body.Body())
		if i.Then == nil {
			return nil
		}

		return &i
	}
}

func ElseIf() parser.Func[*ast.ElseIf] {
	return func(p *parser.Parser) *ast.ElseIf {
		elseKw := parser.TryKeywordAt(p, "else", comment.OrAnyWhitespace())
		ifKw := parser.TryKeywordAt(p, "if", comment.OrAnyWhitespace())
		if elseKw == nil || ifKw == nil {
			return nil
		}

		var ei ast.ElseIf
		ei.Else = elseKw
		ei.If = ifKw
		ei.Header = parser.Try(p, IfHeader())
		if ei.Header == nil {
			return nil
		}
		parser.TrySkip(p, comment.OrHorizontalWhitespace())
		ei.Then = parser.Try(p, body.Body())
		if ei.Then == nil {
			return nil
		}

		return &ei
	}
}

func Else() parser.Func[*ast.Else] {
	return func(p *parser.Parser) *ast.Else {
		elseKw := parser.TryKeywordAt(p, "else", comment.OrAnyWhitespace())
		if elseKw == nil {
			return nil
		}

		var e ast.Else
		e.Else = elseKw
		e.Then = parser.Try(p, body.Body())
		if e.Then == nil {
			return nil
		}
		return &e
	}
}

func IfHeader() parser.Func[*ast.IfHeader] {
	return func(p *parser.Parser) *ast.IfHeader {
		var h ast.IfHeader

		state := p.CloneState()

		h.Statement = parser.TryOptional(p, SimpleStatement(BodyFollows), nil)
		if matches := parser.Try(p, comment.AndEOS()); matches {
			parser.TrySkip(p, comment.OrAnyWhitespace())
		} else {
			p.RestoreState(state)
			h.Statement = nil
		}
		h.Condition = parser.Try(p, Expression(BodyFollows))
		if h.Statement != nil && h.Condition == nil {
			p.RestoreState(state)
			h.Statement = nil
			h.Condition = parser.Try(p, Expression(BodyFollows))
		}

		if h.Condition == nil {
			return nil
		}

		return &h
	}
}

func Switch() parser.Func[*ast.Switch] {
	return func(p *parser.Parser) *ast.Switch {
		switchKw := parser.TryKeywordAt(p, "switch", comment.OrAnyWhitespace())
		if switchKw == nil {
			return nil
		}

		var s ast.Switch
		s.Switch = switchKw
		s.Comparator = parser.TryOptional(p, SimpleStatement(BodyFollows), comment.OrHorizontalWhitespace())

		s.LBrace = parser.TryRuneAt(p, '{')
		if s.LBrace == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "switch: missing body",
				Primary: quickanno.Expected(p, p.Pos(), "an opening brace"),
			})
			return &s
		}
		parser.TrySkip(p, comment.OrAnyWhitespace())
		s.Cases = parser.Collect(p, SwitchCase(), comment.OrAnyWhitespace())

		var annos []diagnostic.Annotation
		for _, c := range s.Cases {
			if c.Default == nil {
				continue
			}
			annos = append(annos, anno.Node(p.File, c, "here"))
		}
		if len(annos) > 0 {
			p.CaptureError(&diagnostic.Diagnostic{
				Message:     "switch: multiple default cases",
				Primary:     slices.Clip(annos),
				Explanation: "A switch statement can only have one default case. Remove all but one default case.",
			})
		}

		err := unexpected.UntilAnyRune(p, comment.OrAnyWhitespace(), '}')
		if err != nil {
			err.Message = "switch: unexpected runes after cases"
			p.CaptureError(err)
		}

		s.RBrace = parser.TryRuneAt(p, '}')
		if s.RBrace == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "switch: missing closing brace",
				Primary: quickanno.Expected(p, *s.LBrace, "expected a `}` for the opening `{` here"),
			})
		}

		return &s
	}
}

func SwitchCase() parser.Func[*ast.Case] {
	return func(p *parser.Parser) *ast.Case {
		if c := parser.Try(p, Case()); c != nil {
			return c
		} else if d := parser.Try(p, Default()); d != nil {
			return d
		}
		return nil
	}
}

func Case() parser.Func[*ast.Case] {
	return func(p *parser.Parser) *ast.Case {
		caseKw := parser.TryKeywordAt(p, "case", comment.OrAnyWhitespace())
		if caseKw == nil {
			return nil
		}

		var c ast.Case
		c.Case = caseKw

		c.Expression = parser.Try(p, Expression(Regular))
		if c.Expression == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "case: missing expression",
				Primary: quickanno.Expected(p, p.Pos(), "an expression"),
			})
		}

		c.Colon = parser.TryOptionalRuneAt(p, ':', comment.OrAnyWhitespace())
		if c.Colon == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "case: missing colon",
				Primary: quickanno.Expected(p, p.Pos(), "a colon"),
			})
		}

		c.Then = parser.Try(p, CaseBody())
		return &c
	}
}

func Default() parser.Func[*ast.Case] {
	return func(p *parser.Parser) *ast.Case {
		defaultKw := parser.TryKeywordAt(p, "default", comment.OrAnyWhitespace())
		if defaultKw == nil {
			return nil
		}

		var c ast.Case
		c.Default = defaultKw

		c.Colon = parser.TryOptionalRuneAt(p, ':', comment.OrAnyWhitespace())
		if c.Colon == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "default: missing colon",
				Primary: quickanno.Expected(p, p.Pos(), "a colon"),
			})
		}

		c.Then = parser.Try(p, CaseBody())
		return &c
	}
}

func CaseBody() parser.Func[[]ast.ScopeNode] {
	return func(p *parser.Parser) []ast.ScopeNode {
		var ns []ast.ScopeNode
		for {
			stop := parser.Matches(p, func(p *parser.Parser) bool {
				if parser.TryKeywordAt(p, "case", comment.OrAnyWhitespace()) != nil {
					return true
				} else if parser.TryKeywordAt(p, "default", comment.OrAnyWhitespace()) != nil {
					return true
				}
				return false
			})
			if stop {
				parser.RestoreWS(p)
				break
			}

			n := parser.Try(p, body.ScopeNode())
			if n == nil {
				break
			}
			ns = append(ns, n)
			parser.TrySkip(p, comment.OrAnyWhitespace())
		}
		if len(ns) == 0 {
			return []ast.ScopeNode{}
		}
		return slices.Clip(ns)
	}
}

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

		h.Condition = parser.Try(p, Expression(BodyFollows))
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

		h.Init = parser.TryOptional(p, SimpleStatement(BodyFollows), nil)
		if matches := parser.Try(p, comment.AndEOS()); !matches {
			return nil
		}
		parser.TrySkip(p, comment.OrAnyWhitespace())
		h.Condition = parser.TryOptional(p, Expression(BodyFollows), nil)
		if matches := parser.Try(p, comment.AndEOS()); !matches {
			return nil
		}
		parser.TrySkip(p, comment.OrAnyWhitespace())
		h.Post = parser.TryOptional(p, SimpleStatement(BodyFollows), nil)

		return &h
	}
}

func ForRangeHeader() parser.Func[*ast.ForRangeHeader] {
	return func(p *parser.Parser) *ast.ForRangeHeader {
		var h ast.ForRangeHeader

		pos := p.Pos()
		var comma *ast.Position
		if !parser.Matches(p, func(p *parser.Parser) bool {
			if parser.TryKeywordAt(p, "range", comment.OrAnyWhitespace()) != nil {
				return true
			} else if parser.TryKeywordAt(p, "ordered", comment.OrHorizontalWhitespace()) != nil {
				return true
			}
			return false
		}) {
			h.Var1 = parser.TryOptional(p, Expression(BodyFollows), comment.OrHorizontalWhitespace())
			comma = parser.TryOptionalRuneAt(p, ',', comment.OrAnyWhitespace())
			if comma != nil {
				if h.Var1 == nil {
					p.CaptureError(&diagnostic.Diagnostic{
						Message: "for-range header: missing first variable",
						Primary: quickanno.Expected(p, pos, "a variable"),
					})
				}

				h.Var2 = parser.TryOptional(p, Expression(BodyFollows), comment.OrHorizontalWhitespace())
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

		h.Ordered = parser.TryOptionalKeywordAt(p, "ordered", comment.OrHorizontalWhitespace())
		rangePos := p.Pos()
		parser.TrySkip(p, comment.OrAnyWhitespace())
		h.Range = parser.TryKeywordAt(p, "range", comment.OrAnyWhitespace())
		if h.Range == nil {
			if h.Ordered == nil {
				return nil
			}
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "for-range header: missing range keyword",
				Primary: quickanno.Expected(p, rangePos, "the `range` keyword"),
			})
		}

		h.Expression = parser.Try(p, Expression(BodyFollows))
		if h.Expression == nil {
			return nil
		}

		return &h
	}
}
