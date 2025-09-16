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
	return func(p *parser.Parser) *ast.ZeroCoalescing {
		var derefPosition *ast.Position
		derefCount := len(parser.TokenWhile(p, func() bool {
			return parser.MatchesAnyRune(p, '*')
		}))
		if derefCount > 0 {
			derefPosition = p.PosPtr()
			derefPosition.Col -= derefCount
			parser.TrySkip(p, comment.OrAnyWhitespace())
		} else {
			derefPosition = nil
		}

		root := parser.Try(p, zeroCoalescingRoot())
		if root == nil {
			return nil
		}

		var zc ast.ZeroCoalescing
		zc.DerefPosition = derefPosition
		zc.DerefCount = derefCount
		zc.Root = root
		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		zc.CheckRoot = parser.TryOptionalRuneAt(p, '?', comment.OrHorizontalWhitespace())

		var chainHasCheck bool
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
				return nil
			}
			return &zc
		}
		parser.TrySkip(p, comment.OrAnyWhitespace())

		zc.Default = parser.Try(p, SimpleExpression())
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
		return &zc
	}
}

func zeroCoalescingRoot() parser.Func[*ast.Expression] {
	return func(p *parser.Parser) *ast.Expression {
		if parenExpr := parser.Try(p, ParenExpression()); parenExpr != nil {
			return parenExpr
		}

		ident := parser.Try(p, golang.Identifier())
		if ident != nil {
			return &ast.Expression{
				Nodes: ast.Code{
					&ast.GoCode{Code: ident.Name, Position: ident.Position},
				},
			}
		}

		return nil
	}
}

type zeroCoalescingNodeData[T ast.ZeroCoalescingNode] struct {
	node     T
	hasCheck bool
}

func zeroCoalescingNode() parser.Func[*zeroCoalescingNodeData[ast.ZeroCoalescingNode]] {
	return func(p *parser.Parser) *zeroCoalescingNodeData[ast.ZeroCoalescingNode] {
		if ie := parser.Try(p, zcIndexExpression()); ie != nil {
			return &zeroCoalescingNodeData[ast.ZeroCoalescingNode]{ie.node, ie.hasCheck}
		} else if se := parser.Try(p, zcSelectorExpression()); se != nil {
			return &zeroCoalescingNodeData[ast.ZeroCoalescingNode]{se.node, se.hasCheck}
		} else if pe := parser.Try(p, zcParenExpression()); pe != nil {
			return &zeroCoalescingNodeData[ast.ZeroCoalescingNode]{pe.node, pe.hasCheck}
		} else if tae := parser.Try(p, zcTypeAssertionExpression()); tae != nil {
			return &zeroCoalescingNodeData[ast.ZeroCoalescingNode]{tae.node, tae.hasCheck}
		}

		return nil
	}
}

func ZCIndexExpression() parser.Func[*ast.ZCIndexExpression] {
	return func(p *parser.Parser) *ast.ZCIndexExpression {
		d := parser.Try(p, zcIndexExpression())
		return d.node
	}
}

func zcIndexExpression() parser.Func[*zeroCoalescingNodeData[*ast.ZCIndexExpression]] {
	return func(p *parser.Parser) *zeroCoalescingNodeData[*ast.ZCIndexExpression] {
		lBracket := parser.TryRuneAt(p, '[')
		if lBracket == nil {
			return nil
		}

		var ie ast.ZCIndexExpression
		ie.LBracket = lBracket

		parser.TrySkip(p, comment.OrAnyWhitespace())
		pos := p.Pos()
		ie.Index = parser.Try(p, SimpleExpression())
		if ie.Index == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "index expression: missing index",
				Primary: quickanno.Expected(p, pos, "an index expression"),
				Secondary: []diagnostic.Annotation{
					anno.Position(p.File, *ie.LBracket, "because of the opening '[' here"),
				},
			})
			return nil
		}

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
		}
	}
}

func ZCSelectorExpression() parser.Func[*ast.ZCSelectorExpression] {
	return func(p *parser.Parser) *ast.ZCSelectorExpression {
		d := parser.Try(p, zcSelectorExpression())
		if d == nil {
			return nil
		}
		return d.node
	}
}

func zcSelectorExpression() parser.Func[*zeroCoalescingNodeData[*ast.ZCSelectorExpression]] {
	return func(p *parser.Parser) *zeroCoalescingNodeData[*ast.ZCSelectorExpression] {
		dot := parser.TryRuneAt(p, '.')
		if dot == nil {
			return nil
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())
		ident := parser.Try(p, golang.Identifier())
		if ident == nil {
			return nil
		}

		var se ast.ZCSelectorExpression
		se.Dot = dot
		se.Ident = ident

		parser.TrySkip(p, comment.OrHorizontalWhitespace())
		se.Check = parser.TryRuneAt(p, '?')

		return &zeroCoalescingNodeData[*ast.ZCSelectorExpression]{
			node:     &se,
			hasCheck: se.Check != nil,
		}
	}
}

func ZCParenExpression() parser.Func[*ast.ZCParenExpression] {
	return func(p *parser.Parser) *ast.ZCParenExpression {
		d := parser.Try(p, zcParenExpression())
		if d == nil {
			return nil
		}
		return d.node
	}
}

func zcParenExpression() parser.Func[*zeroCoalescingNodeData[*ast.ZCParenExpression]] {
	return func(p *parser.Parser) *zeroCoalescingNodeData[*ast.ZCParenExpression] {
		l := parser.Try(p, list.ParenList("function", "argument", "arguments", SimpleExpression()))
		if l == nil {
			return nil
		}

		var pe ast.ZCParenExpression
		pe.LParen, pe.Args, pe.RParen = l.Open, l.Elems, l.Close

		parser.TrySkip(p, comment.OrHorizontalWhitespace())
		pe.Check = parser.TryRuneAt(p, '?')

		return &zeroCoalescingNodeData[*ast.ZCParenExpression]{
			node:     &pe,
			hasCheck: pe.Check != nil,
		}
	}
}

func ZCTypeAssertionExpression() parser.Func[*ast.ZCTypeAssertionExpression] {
	return func(p *parser.Parser) *ast.ZCTypeAssertionExpression {
		d := parser.Try(p, zcTypeAssertionExpression())
		if d == nil {
			return nil
		}
		return d.node
	}
}

func zcTypeAssertionExpression() parser.Func[*zeroCoalescingNodeData[*ast.ZCTypeAssertionExpression]] {
	return func(p *parser.Parser) *zeroCoalescingNodeData[*ast.ZCTypeAssertionExpression] {
		dot := parser.TryRuneAt(p, '.')
		if dot == nil {
			return nil
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())

		lParen := parser.TryRuneAt(p, '(')
		if lParen == nil {
			return nil
		}

		var tae ast.ZCTypeAssertionExpression
		tae.Dot = dot
		tae.LParen = lParen

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
		}
	}
}
