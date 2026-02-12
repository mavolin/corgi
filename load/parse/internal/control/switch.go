package control

import (
	"slices"

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

func Switch() parser.Func[*ast.Switch] {
	return func(p *parser.Parser) *ast.Switch {
		switchKw := parser.TryKeywordAt(p, "switch")
		if switchKw == nil {
			return nil
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())

		var s ast.Switch
		s.Switch = switchKw
		s.Comparator = parser.TryOptional(p, code.SimpleStatement(), comment.OrHorizontalWhitespace())

		s.LBrace = parser.TryRuneAt(p, '{')
		if s.LBrace == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "switch: missing body",
				Primary: quickanno.Expected(p, p.Pos(), "an opening brace"),
			})
			return &s
		}
		parser.TrySkip(p, comment.OrAnyWhitespace())

		err := unexpected.UntilAnyToken(p, comment.OrAnyWhitespace(), "case", "default", "}")
		if err != nil {
			err.Message = "switch: unexpected runes before cases"
			p.CaptureError(err)
			parser.TrySkip(p, comment.OrAnyWhitespace())
		}

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

		err = unexpected.UntilAnyRune(p, comment.OrAnyWhitespace(), '}')
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
		caseKw := parser.TryKeywordAt(p, "case")
		if caseKw == nil {
			return nil
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())

		var c ast.Case
		c.Case = caseKw

		c.Expression = parser.Try(p, code.Expression())
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
		defaultKw := parser.TryKeywordAt(p, "default")
		if defaultKw == nil {
			return nil
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())

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
				if parser.TryKeywordAt(p, "case") != nil {
					return true
				} else if parser.TryKeywordAt(p, "default") != nil {
					return true
				}
				return parser.MatchesRune(p, '}')
			})
			if stop {
				parser.RestoreWS(p)
				break
			}

			n := parser.TryOptional(p, body.ScopeNode(), nil)
			if n != nil {
				ns = append(ns, n)
				parser.Try(p, comment.AndMustEOS())
				continue
			}

			bn := parser.TryOptional(p, body.BadNode(), nil)
			if bn == nil {
				break
			}
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "bad scope node",
				Primary: []diagnostic.Annotation{
					anno.Range(p.File, bn.From, bn.Until, "unexpected tokens"),
				},
			})
			ns = append(ns, bn)
			parser.Try(p, comment.AndMustEOS())
			parser.TrySkip(p, comment.OrAnyWhitespace())
		}
		if len(ns) == 0 {
			return []ast.ScopeNode{}
		}
		return slices.Clip(ns)
	}
}
