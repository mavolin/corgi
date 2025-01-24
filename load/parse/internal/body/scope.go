package body

import (
	"slices"

	"github.com/mavolin/corgi/v2/fancyerr"
	"github.com/mavolin/corgi/v2/fancyerr/anno"
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/unexpected"
)

func Scope() parser.Func[*ast.Scope] {
	return func(p *parser.Parser) (*ast.Scope, *fancyerr.Error) {
		s := &ast.Scope{LBrace: p.Pos()}
		if !parser.TryRune(p, '{') {
			return nil, &fancyerr.Error{
				Message: "missing scope",
				Primary: quickanno.Expected(p, p.Pos(), "a opening brace"),
			}
		}

		s.Nodes = make([]ast.ScopeNode, 0, 24)
		for {
			parser.TrySkip(p, comment.OrAnyWhitespace())

			n, ok := parser.TryOk(p, ScopeNode())
			if !ok {
				break
			}
			s.Nodes = append(s.Nodes, n)
		}
		if len(s.Nodes) == 0 {
			s.Nodes = nil
		} else {
			s.Nodes = slices.Clip(s.Nodes)
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())

		s.RBrace = p.PosPtr()
		if !parser.TryRune(p, '}') {
			s.RBrace = nil
			p.CaptureError(&fancyerr.Error{
				Message: "unclosed scope",
				Primary: quickanno.Expected(p, s.LBrace, "expected a `}` for the opening `{` here"),
			})
		}

		return s, nil
	}
}

var scopeNode parser.Func[ast.ScopeNode]

func SetScopeNode(f parser.Func[ast.ScopeNode]) {
	scopeNode = f
}

func ScopeNode() parser.Func[ast.ScopeNode] {
	return scopeNode
}

func BadScopeNode() parser.Func[*ast.BadScopeNode] {
	return func(p *parser.Parser) (*ast.BadScopeNode, *fancyerr.Error) {
		b := &ast.BadScopeNode{Start: p.Pos()}
		for {
			unexpected.UntilAnyRune(p, nil, '(', ')', '[', ']', '{', '}', ';')
			b.Until = p.Pos()
			parser.TrySkip(p, comment.OrHorizontalWhitespace())
			if parser.MatchesAnyRune(p, '}', ';') {
				break
			} else if parser.MatchesAnyRune(p, '{') {
				if _, ok := parser.TryOptionalOk(p, Scope()); !ok {
					parser.TryRune(p, '{')
				}
			} else if parser.MatchesAnyRune(p, '[') {
				if _, ok := parser.TryOptionalOk(p, BracketText()); !ok {
					parser.TryRune(p, '[')
				}
			}
			parser.TrySkip(p, comment.OrAnyWhitespace())
			if parser.MatchesAnyRune(p, '}') {
				break
			} else if parser.Matches(p, ScopeNode()) {
				break
			}

			if parser.TryOptionalRune(p, ']') {
			} else if parser.TryOptionalRune(p, '(') {
			} else if parser.TryOptionalRune(p, ')') {
			}
		}

		parser.RestoreWS(p)
		b.Until = p.Pos()
		p.CaptureError(&fancyerr.Error{
			Message: "bad scope node",
			Primary: []fancyerr.Annotation{
				anno.Range(p.File, b.Start, b.Until, "unexpected tokens"),
			},
		})
		return b, nil
	}
}
