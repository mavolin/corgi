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
	return func(p *parser.Parser) (*ast.Conditional, *diagnostic.Diagnostic) {
		var c ast.Conditional

		var err *diagnostic.Diagnostic
		c.If, err = parser.TryErr(p, If())
		if err != nil {
			return nil, err
		}

		c.ElseIfs = parser.Collect(p, ElseIf(), 24, comment.OrAnyWhitespace())
		parser.TrySkip(p, comment.OrAnyWhitespace())
		c.Else = parser.Try(p, Else())

		return &c, nil
	}
}

func If() parser.Func[*ast.If] {
	return func(p *parser.Parser) (*ast.If, *diagnostic.Diagnostic) {
		var i ast.If

		i.If = parser.TryKeywordAt(p, "if", comment.OrAnyWhitespace())
		if i.If == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "missing if",
				Primary: quickanno.Expected(p, p.Pos(), "the `if` keyword"),
			}
		}

		i.Header = parser.Must(p, IfHeader())
		parser.TrySkip(p, comment.OrHorizontalWhitespace())
		i.Then = parser.Try(p, body.Body())
		if i.Then == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "if: missing body",
				Primary: quickanno.Expected(p, p.Pos(), "a body"),
			}
		}

		return &i, nil
	}
}

func ElseIf() parser.Func[*ast.ElseIf] {
	return func(p *parser.Parser) (*ast.ElseIf, *diagnostic.Diagnostic) {
		var ei ast.ElseIf

		ei.Else = parser.TryKeywordAt(p, "else", comment.OrAnyWhitespace())
		ei.If = parser.TryKeywordAt(p, "if", comment.OrAnyWhitespace())
		if ei.Else == nil || ei.If == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "missing else if",
				Primary: quickanno.Expected(p, p.Pos(), "the `else if` keywords"),
			}
		}

		ei.Header = parser.Must(p, IfHeader())
		parser.TrySkip(p, comment.OrHorizontalWhitespace())
		ei.Then = parser.Try(p, body.Body())
		if ei.Then == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "else if: missing body",
				Primary: quickanno.Expected(p, p.Pos(), "a body"),
			}
		}

		return &ei, nil
	}
}

func Else() parser.Func[*ast.Else] {
	return func(p *parser.Parser) (*ast.Else, *diagnostic.Diagnostic) {
		var e ast.Else

		e.Else = parser.TryKeywordAt(p, "else", comment.OrAnyWhitespace())
		if e.Else == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "missing else",
				Primary: quickanno.Expected(p, p.Pos(), "the `else` keyword"),
			}
		}

		e.Then = parser.Try(p, body.Body())
		if e.Then == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "else: missing body",
				Primary: quickanno.Expected(p, p.Pos(), "a body"),
			}
		}
		return &e, nil
	}
}

func IfHeader() parser.Func[*ast.IfHeader] {
	return func(p *parser.Parser) (*ast.IfHeader, *diagnostic.Diagnostic) {
		var h ast.IfHeader

		state := p.CloneState()

		h.Statement = parser.TryOptional(p, SimpleStatement(BodyFollows), nil)
		if parser.TrySkip(p, comment.AndEOS()) {
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
			return nil, &diagnostic.Diagnostic{
				Message: "missing if header",
				Primary: quickanno.Expected(p, p.Pos(), "an expression"),
			}
		}

		return &h, nil
	}
}

func Switch() parser.Func[*ast.Switch] {
	return func(p *parser.Parser) (*ast.Switch, *diagnostic.Diagnostic) {
		var s ast.Switch

		s.Switch = parser.TryKeywordAt(p, "switch", comment.OrAnyWhitespace())
		if s.Switch == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "missing switch",
				Primary: quickanno.Expected(p, p.Pos(), "the `switch` keyword"),
			}
		}

		s.Comparator = parser.TryOptional(p, SimpleStatement(BodyFollows), comment.OrHorizontalWhitespace())

		s.LBrace = parser.TryRuneAt(p, '{')
		if s.LBrace == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "switch: missing body",
				Primary: quickanno.Expected(p, p.Pos(), "an opening brace"),
			})
			return &s, nil
		}
		parser.TrySkip(p, comment.OrAnyWhitespace())
		s.Cases = parser.Collect(p, SwitchCase(), 24, comment.OrAnyWhitespace())

		annos := make([]diagnostic.Annotation, 0, len(s.Cases))
		for _, c := range s.Cases {
			if c.Default == nil {
				continue
			}
			switch len(annos) {
			case 0:
				annos = append(annos, anno.NRunes(p.File, *c.Default, len("default"), "first default case"))
			case 1:
				annos = append(annos, anno.NRunes(p.File, *c.Default, len("default"), "second default case"))
			default:
				annos[1] = anno.NRunes(p.File, *c.Default, len("default"), "another default case")
			}
		}
		if len(annos) > 0 {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "switch: multiple default cases",
				Primary: slices.Clip(annos),
				Explanation: "A switch statement can only have one default case. " +
					"Remove all but one default case.",
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

		return &s, nil
	}
}

func SwitchCase() parser.Func[*ast.Case] {
	return func(p *parser.Parser) (*ast.Case, *diagnostic.Diagnostic) {
		if c := parser.Try(p, Case()); c != nil {
			return c, nil
		} else if d := parser.Try(p, Default()); d != nil {
			return d, nil
		}
		return nil, &diagnostic.Diagnostic{
			Message: "missing case",
			Primary: quickanno.Expected(p, p.Pos(), "a case or default"),
		}
	}
}

func Case() parser.Func[*ast.Case] {
	return func(p *parser.Parser) (*ast.Case, *diagnostic.Diagnostic) {
		var c ast.Case

		c.Case = parser.TryKeywordAt(p, "case", comment.OrAnyWhitespace())
		if c.Case == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "missing case",
				Primary: quickanno.Expected(p, p.Pos(), "the `case` keyword"),
			}
		}

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
		return &c, nil
	}
}

func Default() parser.Func[*ast.Case] {
	return func(p *parser.Parser) (*ast.Case, *diagnostic.Diagnostic) {
		var c ast.Case

		c.Default = parser.TryKeywordAt(p, "default", comment.OrAnyWhitespace())
		if c.Default == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "missing default",
				Primary: quickanno.Expected(p, p.Pos(), "the `default` keyword"),
			}
		}

		c.Colon = parser.TryOptionalRuneAt(p, ':', comment.OrAnyWhitespace())
		if c.Colon == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "default: missing colon",
				Primary: quickanno.Expected(p, p.Pos(), "a colon"),
			})
		}

		c.Then = parser.Try(p, CaseBody())
		return &c, nil
	}
}

func CaseBody() parser.Func[[]ast.ScopeNode] {
	return func(p *parser.Parser) ([]ast.ScopeNode, *diagnostic.Diagnostic) {
		ns := make([]ast.ScopeNode, 0, 64)
		for {
			stop := parser.Matches(p, func(p *parser.Parser) (struct{}, *diagnostic.Diagnostic) {
				if parser.TryKeywordAt(p, "case", comment.OrAnyWhitespace()) != nil {
					return struct{}{}, nil
				} else if parser.TryKeywordAt(p, "default", comment.OrAnyWhitespace()) != nil {
					return struct{}{}, nil
				}
				return struct{}{}, new(diagnostic.Diagnostic)
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
			return nil, nil
		}
		return slices.Clip(ns), nil
	}
}

func For() parser.Func[*ast.For] {
	return func(p *parser.Parser) (*ast.For, *diagnostic.Diagnostic) {
		var f ast.For

		f.For = parser.TryKeywordAt(p, "for", comment.OrAnyWhitespace())
		if f.For == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "missing for loop",
				Primary: quickanno.Expected(p, p.Pos(), "the `for` keyword"),
			}
		}
		f.Header = parser.Try(p, ForHeader())

		err := unexpected.UntilAnyRune(p, comment.OrHorizontalWhitespace(), '{', '[')
		if err != nil {
			err.Message = "for loop: unexpected runes after header"
			p.CaptureError(err)
		}

		f.Body = parser.Try(p, body.Body())
		if f.Body == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "for loop: missing body",
				Primary: quickanno.Expected(p, p.Pos(), "a body"),
			}
		}
		return &f, nil
	}
}

func ForHeader() parser.Func[ast.ForHeader] {
	return func(p *parser.Parser) (ast.ForHeader, *diagnostic.Diagnostic) {
		if frh := parser.Try(p, ForRangeHeader()); frh != nil {
			return frh, nil
		} else if fch := parser.Try(p, ForClauseHeader()); fch != nil {
			return fch, nil
		} else if fch := parser.Try(p, ForConditionHeader()); fch != nil {
			return fch, nil
		}
		return nil, &diagnostic.Diagnostic{
			Message: "missing for header",
			Primary: quickanno.Expected(p, p.Pos(), "a for-clause, for-range, or for-condition header"),
		}
	}
}

func ForConditionHeader() parser.Func[*ast.ForConditionHeader] {
	return func(p *parser.Parser) (*ast.ForConditionHeader, *diagnostic.Diagnostic) {
		var h ast.ForConditionHeader

		h.Condition = parser.Try(p, Expression(BodyFollows))
		if h.Condition == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "for-condition header: missing condition",
				Primary: quickanno.Expected(p, p.Pos(), "an expression"),
			}
		}

		return &h, nil
	}
}

func ForClauseHeader() parser.Func[*ast.ForClauseHeader] {
	return func(p *parser.Parser) (*ast.ForClauseHeader, *diagnostic.Diagnostic) {
		var h ast.ForClauseHeader

		if parser.MatchesAnyRune(p, '{', '[') {
			return nil, &diagnostic.Diagnostic{
				Message: "missing for-clause header",
				Primary: quickanno.Expected(p, p.Pos(), "a for-clause header"),
			}
		}

		h.Init = parser.TryOptional(p, SimpleStatement(BodyFollows), nil)
		if !parser.TrySkip(p, comment.AndEOS()) {
			return nil, &diagnostic.Diagnostic{
				Message: "for-clause header: missing init clause",
				Primary: quickanno.Expected(p, p.Pos(), "a simple statement or a semicolon"),
			}
		}
		parser.TrySkip(p, comment.OrAnyWhitespace())
		h.Condition = parser.TryOptional(p, Expression(BodyFollows), nil)
		if !parser.TrySkip(p, comment.AndEOS()) {
			return nil, &diagnostic.Diagnostic{
				Message: "for-clause header: missing condition clause",
				Primary: quickanno.Expected(p, p.Pos(), "an expression or a semicolon"),
			}
		}
		parser.TrySkip(p, comment.OrAnyWhitespace())
		h.Post = parser.TryOptional(p, SimpleStatement(BodyFollows), nil)

		return &h, nil
	}
}

func ForRangeHeader() parser.Func[*ast.ForRangeHeader] {
	return func(p *parser.Parser) (*ast.ForRangeHeader, *diagnostic.Diagnostic) {
		var h ast.ForRangeHeader

		pos := p.Pos()
		var comma *ast.Position
		if !parser.Matches(p, func(p *parser.Parser) (struct{}, *diagnostic.Diagnostic) {
			if parser.TryKeywordAt(p, "range", comment.OrAnyWhitespace()) != nil {
				return struct{}{}, nil
			} else if parser.TryKeywordAt(p, "ordered", comment.OrHorizontalWhitespace()) != nil {
				return struct{}{}, nil
			}
			return struct{}{}, new(diagnostic.Diagnostic)
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
				return nil, &diagnostic.Diagnostic{
					Message: "missing for-range header",
					Primary: quickanno.Expected(p, pos, "a for-range header"),
				}
			}
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "for-range header: missing range keyword",
				Primary: quickanno.Expected(p, rangePos, "the `range` keyword"),
			})
		}

		h.Expression = parser.Try(p, Expression(BodyFollows))
		if h.Expression == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "for-range header: missing expression",
				Primary: quickanno.Expected(p, p.Pos(), "an expression"),
			}
		}

		return &h, nil
	}
}
