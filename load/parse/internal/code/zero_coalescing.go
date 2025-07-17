package code

import (
	"slices"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/golang"
	"github.com/mavolin/corgi/v2/load/parse/internal/list"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

func ZeroCoalescing() parser.Func[*ast.ZeroCoalescing] {
	return func(p *parser.Parser) (*ast.ZeroCoalescing, *diagnostic.Diagnostic) {
		var zc ast.ZeroCoalescing

		zc.DerefPosition = p.PosPtr()
		zc.DerefCount = len(parser.TokenWhile(p, func() bool {
			return parser.MatchesAnyRune(p, '*')
		}))
		if zc.DerefCount > 0 {
			parser.TrySkip(p, comment.OrAnyWhitespace())
		} else {
			zc.DerefPosition = nil
		}

		zc.Root = parser.Try(p, zeroCoalescingRoot())
		if zc.Root == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "missing zero coalescing",
				Primary: quickanno.Expected(p, p.Pos(), "a root"),
				Examples: []diagnostic.Example{
					{Example: "`foo?`"},
				},
			}
		}
		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		zc.CheckRoot = parser.TryOptionalRuneAt(p, '?', comment.OrHorizontalWhitespace())

		var chainHasCheck bool
		zc.Chain = make([]ast.ZeroCoalescingNode, 0, 12)
		for {
			parser.TrySkip(p, comment.OrHorizontalWhitespace())
			n := parser.TryOptional(p, zeroCoalescingNode(), nil)
			if n == nil {
				break
			}
			zc.Chain = append(zc.Chain, n.node)
			chainHasCheck = chainHasCheck || n.hasCheck
		}
		if len(zc.Chain) == 0 {
			zc.Chain = nil
		} else {
			zc.Chain = slices.Clip(zc.Chain)
		}

		zc.Tilde = parser.TryRuneAt(p, '~')
		if zc.Tilde == nil {
			if zc.CheckRoot == nil && !chainHasCheck {
				return nil, &diagnostic.Diagnostic{
					Message: "zero coalescing: just regular GoCode",
					Primary: quickanno.Expected(p, p.Pos(), "a zero coalescing"),
					Hints: []diagnostic.Hint{
						{Hint: "This is a sentinel error that you shouldn't see, please open an issue."},
					},
				}
			}
			return &zc, nil
		}
		parser.TrySkip(p, comment.OrAnyWhitespace())

		zc.Default = parser.Try(p, NonZCExpression(Regular))
		if zc.Default == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "zero coalescing: missing default value",
				Primary: quickanno.Expected(p, p.Pos(), "a default value"),
				Secondary: []diagnostic.Annotation{
					anno.Position(p.File, *zc.Tilde, "because of the tilde here"),
				},
				Hints: []diagnostic.Hint{
					{Hint: "If you don't want a default value, remove the tilde."},
				},
			})
		}

		if zc.CheckRoot == nil && !chainHasCheck {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "zero coalescing: redundant default value",
				Primary: []diagnostic.Annotation{
					anno.Range(p.File, *zc.Tilde, zc.End(),
						"this default value is never used, because no element of the zero coalescing has a check"),
				},
				Hints: []diagnostic.Hint{
					{
						Hint:    "You can add a check to a node by appending a '?' to it.",
						Example: "`foo.bar?.baz` or `foo[1?]?`",
					},
				},
				Docs: "zero-coalescing",
			})
		}
		return &zc, nil
	}
}

func zeroCoalescingRoot() parser.Func[*ast.Expression] {
	return func(p *parser.Parser) (*ast.Expression, *diagnostic.Diagnostic) {
		if parser.MatchesAnyRune(p, '(') {
			c, err := parser.TryErr(p, goCode(FirstParen))
			if err != nil {
				return nil, err
			}
			return &ast.Expression{Nodes: c.Nodes}, err
		}

		ident := parser.Try(p, golang.Identifier())
		if ident != nil {
			return &ast.Expression{
				Nodes: ast.Code{
					&ast.GoCode{Code: ident.Ident, Position: ident.Position},
				},
			}, nil
		}

		return nil, &diagnostic.Diagnostic{
			Message: "missing root",
			Primary: quickanno.Expected(p, p.Pos(), "a root"),
		}
	}
}

type zeroCoalescingNodeData[T ast.ZeroCoalescingNode] struct {
	node     T
	hasCheck bool
}

func zeroCoalescingNode() parser.Func[*zeroCoalescingNodeData[ast.ZeroCoalescingNode]] {
	return func(p *parser.Parser) (*zeroCoalescingNodeData[ast.ZeroCoalescingNode], *diagnostic.Diagnostic) {
		if ie := parser.Try(p, zcIndexExpression()); ie != nil {
			return &zeroCoalescingNodeData[ast.ZeroCoalescingNode]{ie.node, ie.hasCheck}, nil
		} else if se := parser.Try(p, zcSelectorExpression()); se != nil {
			return &zeroCoalescingNodeData[ast.ZeroCoalescingNode]{se.node, se.hasCheck}, nil
		} else if pe := parser.Try(p, zcParenExpression()); pe != nil {
			return &zeroCoalescingNodeData[ast.ZeroCoalescingNode]{pe.node, pe.hasCheck}, nil
		} else if tae := parser.Try(p, zcTypeAssertionExpression()); tae != nil {
			return &zeroCoalescingNodeData[ast.ZeroCoalescingNode]{tae.node, tae.hasCheck}, nil
		}

		return nil, &diagnostic.Diagnostic{
			Message: "missing zero coalescing node",
			Primary: quickanno.Expected(p, p.Pos(), "a zero coalescing node"),
		}
	}
}

func ZCIndexExpression() parser.Func[*ast.ZCIndexExpression] {
	return func(p *parser.Parser) (*ast.ZCIndexExpression, *diagnostic.Diagnostic) {
		d, err := parser.TryErr(p, zcIndexExpression())
		return d.node, err
	}
}

func zcIndexExpression() parser.Func[*zeroCoalescingNodeData[*ast.ZCIndexExpression]] {
	return func(p *parser.Parser) (*zeroCoalescingNodeData[*ast.ZCIndexExpression], *diagnostic.Diagnostic) {
		var ie ast.ZCIndexExpression

		ie.LBracket = parser.TryRuneAt(p, '[')
		if ie.LBracket == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "missing '['",
				Primary: quickanno.Expected(p, p.Pos(), "["),
			}
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())
		ie.Index = parser.Must(p, NonZCExpression(Regular))

		parser.TrySkip(p, comment.OrHorizontalWhitespace())
		ie.CheckIndex = parser.TryOptionalRuneAt(p, '?', comment.OrHorizontalWhitespace())

		ie.Comma = parser.TryOptionalRuneAt(p, ',', comment.OrAnyWhitespace())

		ie.RBracket = parser.TryOptionalRuneAt(p, ']', comment.OrHorizontalWhitespace())
		if ie.RBracket == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "index expression: missing ']'",
				Primary: quickanno.Expected(p, *ie.LBracket, "a closing ']' for the opening '[' here"),
			})
		}

		ie.CheckValue = parser.TryRuneAt(p, '?')

		return &zeroCoalescingNodeData[*ast.ZCIndexExpression]{
			node:     &ie,
			hasCheck: ie.CheckIndex != nil || ie.CheckValue != nil,
		}, nil
	}
}

func ZCSelectorExpression() parser.Func[*ast.ZCSelectorExpression] {
	return func(p *parser.Parser) (*ast.ZCSelectorExpression, *diagnostic.Diagnostic) {
		d, err := parser.TryErr(p, zcSelectorExpression())
		if err != nil {
			return nil, err
		}
		return d.node, err
	}
}

func zcSelectorExpression() parser.Func[*zeroCoalescingNodeData[*ast.ZCSelectorExpression]] {
	return func(p *parser.Parser) (*zeroCoalescingNodeData[*ast.ZCSelectorExpression], *diagnostic.Diagnostic) {
		var se ast.ZCSelectorExpression

		se.Dot = parser.TryRuneAt(p, '.')
		if se.Dot == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "missing selector expression",
				Primary: quickanno.Expected(p, p.Pos(), "a selector expression"),
			}
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())
		se.Ident = parser.Try(p, golang.Identifier())
		if se.Ident == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "selector expression: missing identifier",
				Primary: quickanno.Expected(p, p.Pos(), "an identifier"),
			}
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())
		se.Check = parser.TryRuneAt(p, '?')

		return &zeroCoalescingNodeData[*ast.ZCSelectorExpression]{
			node:     &se,
			hasCheck: se.Check != nil,
		}, nil
	}
}

func ZCParenExpression() parser.Func[*ast.ZCParenExpression] {
	return func(p *parser.Parser) (*ast.ZCParenExpression, *diagnostic.Diagnostic) {
		d, err := parser.TryErr(p, zcParenExpression())
		if err != nil {
			return nil, err
		}
		return d.node, err
	}
}

func zcParenExpression() parser.Func[*zeroCoalescingNodeData[*ast.ZCParenExpression]] {
	return func(p *parser.Parser) (*zeroCoalescingNodeData[*ast.ZCParenExpression], *diagnostic.Diagnostic) {
		var pe ast.ZCParenExpression

		l := parser.Try(p, list.ParenList("arguments", NonZCExpression(Regular)))
		if l == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "missing paren expression",
				Primary: quickanno.Expected(p, p.Pos(), "a paren expression"),
			}
		}
		pe.LParen, pe.Args, pe.RParen = l.Open, l.Elems, l.Close

		parser.TrySkip(p, comment.OrHorizontalWhitespace())
		pe.Check = parser.TryRuneAt(p, '?')

		return &zeroCoalescingNodeData[*ast.ZCParenExpression]{
			node:     &pe,
			hasCheck: pe.Check != nil,
		}, nil
	}
}

func ZCTypeAssertionExpression() parser.Func[*ast.ZCTypeAssertionExpression] {
	return func(p *parser.Parser) (*ast.ZCTypeAssertionExpression, *diagnostic.Diagnostic) {
		d, err := parser.TryErr(p, zcTypeAssertionExpression())
		if err != nil {
			return nil, err
		}
		return d.node, err
	}
}

func zcTypeAssertionExpression() parser.Func[*zeroCoalescingNodeData[*ast.ZCTypeAssertionExpression]] {
	return func(p *parser.Parser) (*zeroCoalescingNodeData[*ast.ZCTypeAssertionExpression], *diagnostic.Diagnostic) {
		var tae ast.ZCTypeAssertionExpression

		tae.Dot = parser.TryRuneAt(p, '.')
		if tae.Dot == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "missing type assertion expression",
				Primary: quickanno.Expected(p, p.Pos(), "a type assertion expression"),
			}
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())

		tae.LParen = parser.TryRuneAt(p, '(')
		if tae.LParen == nil {
			return nil, &diagnostic.Diagnostic{
				Message: "type assertion expression: missing '('",
				Primary: quickanno.Expected(p, *tae.Dot, "a '(' after here"),
			}
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())
		tae.PointerCount = len(parser.TokenWhile(p, func() bool {
			return parser.MatchesAnyRune(p, '*')
		}))
		parser.TrySkip(p, comment.OrAnyWhitespace())

		tae.Type = parser.Try(p, golang.FullIdent())
		if tae.Type == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "type assertion expression: missing type",
				Primary: quickanno.Expected(p, p.Pos(), "a type"),
			})
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())
		tae.CheckType = parser.TryOptionalRuneAt(p, '?', comment.OrHorizontalWhitespace())

		tae.RParen = parser.TryOptionalRuneAt(p, ')', comment.OrHorizontalWhitespace())
		if tae.RParen == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "type assertion expression: missing ')'",
				Primary: quickanno.Expected(p, *tae.LParen, "a closing ')' for the opening '(' here"),
			})
		}

		tae.CheckValue = parser.TryRuneAt(p, '?')

		return &zeroCoalescingNodeData[*ast.ZCTypeAssertionExpression]{
			node:     &tae,
			hasCheck: tae.CheckType != nil || tae.CheckValue != nil,
		}, nil
	}
}
