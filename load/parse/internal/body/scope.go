package body

import (
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/unexpected"
)

func Scope() parser.Func[*ast.Scope] {
	return func(p *parser.Parser) *ast.Scope {
		lBrace := parser.TryRuneAt(p, '{')
		if lBrace == nil {
			return nil
		}

		var s ast.Scope
		s.LBrace = lBrace

		for {
			parser.TrySkip(p, comment.OrAnyWhitespace())
			if parser.MatchesRune(p, '}') {
				break
			}

			n := parser.TryOptional(p, ScopeNode(), nil)
			if n != nil {
				s.Nodes = append(s.Nodes, n)
				parser.Try(p, comment.AndForceEOS())
				continue
			}

			bn := parser.TryOptional(p, BadNode(), nil)
			if bn == nil {
				break
			}
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "bad scope node",
				Primary: []diagnostic.Annotation{
					anno.Range(p.File, bn.From, bn.Until, "unexpected tokens"),
				},
			})
			s.Nodes = append(s.Nodes, bn)
			parser.Try(p, comment.AndForceEOS())
		}

		s.RBrace = parser.TryRuneAt(p, '}')
		if s.RBrace == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "unclosed scope",
				Primary: quickanno.Expected(p, *s.LBrace, "expected a `}` for the opening `{` here"),
			})
		}

		return &s
	}
}

var scopeNode parser.Func[ast.ScopeNode]

func SetScopeNode(f parser.Func[ast.ScopeNode]) {
	scopeNode = f
}

func ScopeNode() parser.Func[ast.ScopeNode] {
	return func(p *parser.Parser) ast.ScopeNode {
		n := parser.Try(p, scopeNode)
		if n != nil {
			return n
		}

		return nil
	}
}

func BadNode() parser.Func[*ast.BadNode] {
	return func(p *parser.Parser) *ast.BadNode {
		var b ast.BadNode
		b.From = p.Pos()

		for {
			_ = unexpected.UntilAnyRune(p, nil, '(', ')', '[', ']', '{', '}', ';')
			b.Until = p.Pos()
			parser.TrySkip(p, comment.OrHorizontalWhitespace())
			if parser.MatchesAnyRune(p, '}', ';') { //nolint:gocritic
				break
			} else if parser.MatchesRune(p, '{') {
				if parser.TryOptional(p, Scope(), nil) == nil {
					parser.TryRune(p, '{')
				}
			} else if parser.MatchesRune(p, '[') {
				if parser.TryOptional(p, BracketText(), nil) == nil {
					parser.TryRune(p, '[')
				}
			} else if parser.Matches(p, comment.AndEOS()) {
				break
			} else {
				parser.TryAnyOptionalRune(p, nil, ']', '(', ')')
			}
		}

		parser.RestoreWS(p)
		b.Until = p.Pos()
		if b.Until == b.From {
			return nil
		}
		return &b
	}
}
