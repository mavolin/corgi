package body

import (
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
		var s ast.Scope

		s.LBrace = parser.TryRuneAt(p, '{')
		if s.LBrace == nil {
			return nil, &fancyerr.Error{
				Message: "missing scope",
				Primary: quickanno.Expected(p, p.Pos(), "a opening brace"),
			}
		}

		s.Nodes = parser.Collect(p, ScopeNode(), 24, comment.OrAnyWhitespace())
		parser.TrySkip(p, comment.OrAnyWhitespace())

		s.RBrace = parser.TryRuneAt(p, '}')
		if s.RBrace == nil {
			p.CaptureError(&fancyerr.Error{
				Message: "unclosed scope",
				Primary: quickanno.Expected(p, *s.LBrace, "expected a `}` for the opening `{` here"),
			})
		}

		return &s, nil
	}
}

var scopeNode parser.Func[ast.ScopeNode]

func SetScopeNode(f parser.Func[ast.ScopeNode]) {
	scopeNode = f
}

func ScopeNode() parser.Func[ast.ScopeNode] {
	return func(p *parser.Parser) (ast.ScopeNode, *fancyerr.Error) {
		n, err := parser.TryErr(p, scopeNode)
		if err == nil {
			return n, nil
		}

		if n := parser.Try(p, BadScopeNode()); n != nil {
			p.CaptureError(&fancyerr.Error{
				Message: "bad scope node",
				Primary: []fancyerr.Annotation{
					anno.Range(p.File, n.From, n.Until, "unexpected tokens"),
				},
			})
			return n, nil
		}

		return nil, err
	}
}

func BadScopeNode() parser.Func[*ast.BadScopeNode] {
	return func(p *parser.Parser) (*ast.BadScopeNode, *fancyerr.Error) {
		var b ast.BadScopeNode
		b.From = p.Pos()

		for {
			_ = unexpected.UntilAnyRune(p, nil, '(', ')', '[', ']', '{', '}', ';')
			b.Until = p.Pos()
			parser.TrySkip(p, comment.OrHorizontalWhitespace())
			if parser.MatchesAnyRune(p, '}', ';') {
				break
			} else if parser.MatchesAnyRune(p, '{') {
				if parser.TryOptional(p, Scope(), nil) == nil {
					parser.TryRune(p, '{')
				}
			} else if parser.MatchesAnyRune(p, '[') {
				if parser.TryOptional(p, BracketText(), nil) == nil {
					parser.TryRune(p, '[')
				}
			}
			parser.TrySkip(p, comment.OrAnyWhitespace())
			if parser.MatchesAnyRune(p, '}') {
				break
			} else if parser.Matches(p, scopeNode) {
				break
			} else if parser.MatchesAnyRune(p, parser.EOF) {
				break
			}

			if parser.TryOptionalRune(p, ']', nil) {
			} else if parser.TryOptionalRune(p, '(', nil) {
			} else if parser.TryOptionalRune(p, ')', nil) {
			}
		}

		parser.RestoreWS(p)
		b.Until = p.Pos()
		if b.Until == b.From {
			return nil, &fancyerr.Error{
				Message: "empty bad scope node",
				Primary: quickanno.Expected(p, b.From, "a scope node"),
			}
		}
		return &b, nil
	}
}
