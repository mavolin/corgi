package code

import (
	"github.com/mavolin/corgi/v2/fancyerr"
	"github.com/mavolin/corgi/v2/fancyerr/anno"
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/golang"
	"github.com/mavolin/corgi/v2/load/parse/internal/list"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

func ZeroCoalescing() parser.Func[*ast.ZeroCoalescing] {
	return func(p *parser.Parser) (*ast.ZeroCoalescing, *fancyerr.Error) {
		zc := &ast.ZeroCoalescing{Position: p.Pos()}

		zc.DerefCount = len(parser.TokenWhile(p, func() bool {
			return parser.MatchesAnyRune(p, '*')
		}))
		if zc.DerefCount > 0 {
			parser.TrySkip(p, comment.OrAnyWhitespace())
		}

		var ok bool
		zc.Root, ok = parser.TryOk(p, zeroCoalescingRoot())
		if !ok {
			return nil, &fancyerr.Error{
				Message: "missing zero coalescing",
				Primary: quickanno.Expected(p, p.Pos(), "a root"),
				Examples: []fancyerr.Example{
					{Example: "`foo?`"},
				},
			}
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())
		zc.CheckRoot = p.PosPtr()
		if !parser.TryOptionalRune(p, '?') {
			zc.CheckRoot = nil
		}

		var chainHasCheck bool
		zc.Chain = make([]ast.ZeroCoalescingNode, 0, 12)
		for {
			parser.TrySkip(p, comment.OrHorizontalWhitespace())
			n, ok := parser.TryOk(p, zeroCoalescingNode())
			if !ok {
				break
			}
			zc.Chain = append(zc.Chain, n.node)
			chainHasCheck = chainHasCheck || n.hasCheck
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())
		zc.Tilde = p.PosPtr()
		if !parser.TryRune(p, '~') {
			zc.Tilde = nil

			if zc.CheckRoot == nil && !chainHasCheck {
				return nil, &fancyerr.Error{
					Message: "zero coalescing: just a regular GoCode",
					Primary: quickanno.Expected(p, p.Pos(), "a zero coalescing"),
					Hints: []fancyerr.Hint{
						{Hint: "This is a sentinel error that you shouldn't see, please open an issue."},
					},
				}
			}
			return zc, nil
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())
		zc.Default, ok = parser.TryOk(p, NonZCExpression())
		if !ok {
			p.CaptureError(&fancyerr.Error{
				Message: "zero coalescing: missing default value",
				Primary: quickanno.Expected(p, p.Pos(), "a default value"),
				Hints: []fancyerr.Hint{
					{Hint: "If you don't want a default value, remove the tilde."},
				},
			})
		}

		if zc.CheckRoot == nil && !chainHasCheck {
			start, end := *zc.Tilde, ast.Position{Line: zc.Tilde.Line, Col: zc.Tilde.Col + 1}
			if zc.Default != nil {
				start, end = zc.Default.Pos(), zc.Default.End()
			}

			p.CaptureError(&fancyerr.Error{
				Message: "zero coalescing: redundant default value",
				Primary: []fancyerr.Annotation{
					anno.Range(p.File, start, end,
						"this default value is never used, because no element of the zero coalescing has a check"),
				},
				Hints: []fancyerr.Hint{
					{
						Hint:    "You can add a check to a node by appending a '?' to it.",
						Example: "`foo.bar?.baz` or `foo[1?]?`",
					},
				},
				Docs: "!zero-coalescing",
			})
		}
		return zc, nil
	}
}

func zeroCoalescingRoot() parser.Func[*ast.Expression] {
	return func(p *parser.Parser) (*ast.Expression, *fancyerr.Error) {
		if parser.MatchesAnyRune(p, '(') {
			c, err := parser.Try(p, goCode(true, false))
			if err != nil {
				return nil, err
			}
			return &ast.Expression{Code: c}, err
		}

		ident, ok := parser.TryOk(p, golang.Identifier())
		if ok {
			return &ast.Expression{
				Code: ast.Code{
					&ast.GoCode{Code: ident.Ident, Position: ident.Position},
				},
			}, nil
		}

		return nil, &fancyerr.Error{
			Message: "missing root",
			Primary: quickanno.Expected(p, p.Pos(), "a root"),
		}
	}
}

type zeroCoalescingNodeData[T ast.ZeroCoalescingNode] struct {
	node     T
	hasCheck bool
}

func zeroCoalescingNode() parser.Func[zeroCoalescingNodeData[ast.ZeroCoalescingNode]] {
	return func(p *parser.Parser) (zeroCoalescingNodeData[ast.ZeroCoalescingNode], *fancyerr.Error) {
		if ie, ok := parser.TryOk(p, zcIndexExpression()); ok {
			return zeroCoalescingNodeData[ast.ZeroCoalescingNode]{ie.node, ie.hasCheck}, nil
		} else if se, ok := parser.TryOk(p, zcSelectorExpression()); ok {
			return zeroCoalescingNodeData[ast.ZeroCoalescingNode]{se.node, se.hasCheck}, nil
		} else if pe, ok := parser.TryOk(p, zcParenExpression()); ok {
			return zeroCoalescingNodeData[ast.ZeroCoalescingNode]{pe.node, pe.hasCheck}, nil
		} else if tae, ok := parser.TryOk(p, zcTypeAssertionExpression()); ok {
			return zeroCoalescingNodeData[ast.ZeroCoalescingNode]{tae.node, tae.hasCheck}, nil
		}

		return zeroCoalescingNodeData[ast.ZeroCoalescingNode]{}, &fancyerr.Error{
			Message: "missing zero coalescing node",
			Primary: quickanno.Expected(p, p.Pos(), "a zero coalescing node"),
		}
	}
}

func ZCIndexExpression() parser.Func[*ast.ZCIndexExpression] {
	return func(p *parser.Parser) (*ast.ZCIndexExpression, *fancyerr.Error) {
		d, err := parser.Try(p, zcIndexExpression())
		return d.node, err
	}
}

func zcIndexExpression() parser.Func[zeroCoalescingNodeData[*ast.ZCIndexExpression]] {
	return func(p *parser.Parser) (zeroCoalescingNodeData[*ast.ZCIndexExpression], *fancyerr.Error) {
		ie := &ast.ZCIndexExpression{LBracket: p.Pos()}
		if !parser.TryRune(p, '[') {
			return zeroCoalescingNodeData[*ast.ZCIndexExpression]{}, &fancyerr.Error{
				Message: "missing '['",
				Primary: quickanno.Expected(p, p.Pos(), "["),
			}
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())
		ie.Index = parser.Must(p, NonZCExpression())

		parser.TrySkip(p, comment.OrHorizontalWhitespace())
		ie.CheckIndex = p.PosPtr()
		if parser.TryOptionalRune(p, '?') {
			parser.TrySkip(p, comment.OrHorizontalWhitespace())
		} else {
			ie.CheckIndex = nil
		}

		ie.Comma = p.PosPtr()
		if parser.TryOptionalRune(p, ',') {
			parser.TrySkip(p, comment.OrAnyWhitespace())
		} else {
			ie.Comma = nil
		}

		ie.RBracket = p.PosPtr()
		if !parser.TryRune(p, ']') {
			ie.RBracket = nil
			p.CaptureError(&fancyerr.Error{
				Message: "index expression: missing ']'",
				Primary: quickanno.Expected(p, ie.LBracket, "a closing ']' for the opening '[' here"),
			})
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())
		ie.CheckValue = p.PosPtr()
		if !parser.TryRune(p, '?') {
			ie.CheckValue = nil
		}

		return zeroCoalescingNodeData[*ast.ZCIndexExpression]{
			node:     ie,
			hasCheck: ie.CheckIndex != nil || ie.CheckValue != nil,
		}, nil
	}
}

func ZCSelectorExpression() parser.Func[*ast.ZCSelectorExpression] {
	return func(p *parser.Parser) (*ast.ZCSelectorExpression, *fancyerr.Error) {
		d, err := parser.Try(p, zcSelectorExpression())
		return d.node, err
	}
}

func zcSelectorExpression() parser.Func[zeroCoalescingNodeData[*ast.ZCSelectorExpression]] {
	return func(p *parser.Parser) (zeroCoalescingNodeData[*ast.ZCSelectorExpression], *fancyerr.Error) {
		se := &ast.ZCSelectorExpression{Dot: p.Pos()}
		if !parser.TryRune(p, '.') {
			return zeroCoalescingNodeData[*ast.ZCSelectorExpression]{}, &fancyerr.Error{
				Message: "missing selector expression",
				Primary: quickanno.Expected(p, p.Pos(), "a selector expression"),
			}
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())
		var ok bool
		se.Ident, ok = parser.TryOk(p, golang.Identifier())
		if !ok {
			return zeroCoalescingNodeData[*ast.ZCSelectorExpression]{}, &fancyerr.Error{
				Message: "selector expression: missing identifier",
				Primary: quickanno.Expected(p, p.Pos(), "an identifier"),
			}
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())
		se.Check = p.PosPtr()
		if !parser.TryRune(p, '?') {
			se.Check = nil
		}

		return zeroCoalescingNodeData[*ast.ZCSelectorExpression]{
			node:     se,
			hasCheck: se.Check != nil,
		}, nil
	}
}

func ZCParenExpression() parser.Func[*ast.ZCParenExpression] {
	return func(p *parser.Parser) (*ast.ZCParenExpression, *fancyerr.Error) {
		d, err := parser.Try(p, zcParenExpression())
		return d.node, err
	}
}

func zcParenExpression() parser.Func[zeroCoalescingNodeData[*ast.ZCParenExpression]] {
	return func(p *parser.Parser) (zeroCoalescingNodeData[*ast.ZCParenExpression], *fancyerr.Error) {
		l, ok := parser.TryOk(p, list.ParenList("arguments", NonZCExpression()))
		if !ok {
			return zeroCoalescingNodeData[*ast.ZCParenExpression]{}, &fancyerr.Error{
				Message: "missing paren expression",
				Primary: quickanno.Expected(p, p.Pos(), "a paren expression"),
			}
		}

		pe := &ast.ZCParenExpression{LParen: l.Open, Args: l.Elems, RParen: l.Close}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())
		pe.Check = p.PosPtr()
		if !parser.TryRune(p, '?') {
			pe.Check = nil
		}

		return zeroCoalescingNodeData[*ast.ZCParenExpression]{
			node:     pe,
			hasCheck: pe.Check != nil,
		}, nil
	}
}

func ZCTypeAssertionExpression() parser.Func[*ast.ZCTypeAssertionExpression] {
	return func(p *parser.Parser) (*ast.ZCTypeAssertionExpression, *fancyerr.Error) {
		d, err := parser.Try(p, zcTypeAssertionExpression())
		return d.node, err
	}
}

func zcTypeAssertionExpression() parser.Func[zeroCoalescingNodeData[*ast.ZCTypeAssertionExpression]] {
	return func(p *parser.Parser) (zeroCoalescingNodeData[*ast.ZCTypeAssertionExpression], *fancyerr.Error) {
		tae := &ast.ZCTypeAssertionExpression{Dot: p.Pos()}
		if !parser.TryRune(p, '.') {
			return zeroCoalescingNodeData[*ast.ZCTypeAssertionExpression]{}, &fancyerr.Error{
				Message: "missing type assertion expression",
				Primary: quickanno.Expected(p, p.Pos(), "a type assertion expression"),
			}
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())

		tae.LParen = p.PosPtr()
		if !parser.TryRune(p, '(') {
			return zeroCoalescingNodeData[*ast.ZCTypeAssertionExpression]{}, &fancyerr.Error{
				Message: "type assertion expression: missing '('",
				Primary: quickanno.Expected(p, tae.Dot, "a '(' after here"),
			}
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())
		tae.PointerCount = len(parser.TokenWhile(p, func() bool {
			return parser.MatchesAnyRune(p, '*')
		}))

		var ok bool
		tae.Type, ok = parser.TryOk(p, golang.FullIdent())
		if !ok {
			p.CaptureError(&fancyerr.Error{
				Message: "type assertion expression: missing type",
				Primary: quickanno.Expected(p, p.Pos(), "a type"),
			})
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())
		tae.CheckType = p.PosPtr()
		if parser.TryOptionalRune(p, '?') {
			parser.TrySkip(p, comment.OrHorizontalWhitespace())
		} else {
			tae.CheckType = nil
		}

		tae.RParen = p.PosPtr()
		if !parser.TryRune(p, ')') {
			tae.RParen = nil
			p.CaptureError(&fancyerr.Error{
				Message: "type assertion expression: missing ')'",
				Primary: quickanno.Expected(p, *tae.LParen, "a closing ')' for the opening '(' here"),
			})
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())
		tae.CheckValue = p.PosPtr()
		if !parser.TryRune(p, '?') {
			tae.CheckValue = nil
		}

		return zeroCoalescingNodeData[*ast.ZCTypeAssertionExpression]{
			node:     tae,
			hasCheck: tae.CheckType != nil || tae.CheckValue != nil,
		}, nil
	}
}
