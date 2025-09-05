package code

import (
	"fmt"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/golang"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

func GoCode(o Options) parser.Func[ast.Code] {
	return func(p *parser.Parser) ast.Code {
		state := &goCodeState{
			start:    p.Index(),
			startPos: *p.PosPtr(),
			hasWS:    parser.TrySkip(p, comment.OrHorizontalWhitespace()),
		}

		goCode(p, state, o)
		state.saveCaptured(p)
		parser.RestoreWS(p)

		return state.nodes
	}
}

type goCodeState struct {
	nodes ast.Code

	start    int
	end      int
	startPos ast.Position
	hasWS    bool
}

func (s *goCodeState) saveCaptured(p *parser.Parser) {
	if s.start >= s.end {
		return
	}

	pos := s.startPos
	s.nodes = append(s.nodes, &ast.GoCode{
		Code:     p.AST.Raw[s.start:s.end],
		Position: &pos,
	})
}

type comparableNode interface {
	comparable
	ast.CodeNode
}

func capture[N comparableNode](p *parser.Parser, s *goCodeState, n N, ws parser.WhitespaceFunc) {
	s.saveCaptured(p)
	var zero N
	if n == zero {
		panic(fmt.Sprintf("parse: %s: was sure %T was ahead, but its parser did not match", p.Pos(), n))
	}
	s.nodes = append(s.nodes, n)
	s.skipWS(p, ws)
	s.resetStart(p)
}

func (s *goCodeState) resetStart(p *parser.Parser) {
	s.start = p.Index()
	s.startPos = *p.PosPtr()
}

func (s *goCodeState) skipWS(p *parser.Parser, ws parser.WhitespaceFunc) {
	s.end = p.Index()
	s.hasWS = parser.TrySkip(p, ws)
}

func (s *goCodeState) isEmpty() bool {
	return len(s.nodes) == 0 && s.start >= s.end
}

// goCode parses go code until it encounters a closing paren, bracket or brace.
func goCode(p *parser.Parser, s *goCodeState, o Options) {
	for {
		if !o.inParen() && parser.Matches(p, comment.AndEOS()) {
			break
		} else if !o.inParen() && !o.statements() {
			if parser.MatchesAnyRune(p, ',', ':') { //nolint:gocritic
				break
			} else if parser.MatchesToken(p, "--") || parser.MatchesToken(p, "++") {
				break
			} else if parser.Matches(p, golang.AssignOp()) {
				break
			}
		}

		if parser.MatchesAnyRune(p, '(', '{', '[') { //nolint:gocritic
			r := parser.PeekRune(p)
			if !o.inParen() && (s.hasWS || s.isEmpty()) && (r == '{' || r == '[') {
				// so we don't parse the body after an if or a sole {...} or [...]
				return
			}

			parenGoCode(p, s)
			continue
		} else if parser.MatchesAnyRune(p, ')', '}', ']') {
			return
		} else if parser.Matches(p, BlockFunction()) {
			capture(p, s, parser.Try(p, BlockFunction()), comment.OrHorizontalWhitespace())
			continue
		} else if parser.MatchesAnyRune(p, '"', '`') {
			capture(p, s, parser.Try(p, String()), comment.OrHorizontalWhitespace())
			continue
		} else if parser.MatchesAnyRune(p, '?') {
			t := parser.Try(p, Ternary())
			if t == nil { // for zero coalescing
				return
			}
			capture(p, s, t, comment.OrHorizontalWhitespace())
			continue
		} else if parser.MatchesAnyRune(p, parser.EOF) {
			break
		}

		switch {
		case parser.TryOptional(p, golang.RuneLit(), nil):
			s.skipWS(p, comment.OrHorizontalWhitespace())
			continue
		case parser.TryAnyOptionalToken(p, nil, "==", "!=", ">=", ">", "<=", "<") != "":
			s.skipWS(p, comment.OrAnyWhitespace())
			continue
		case parser.TryAnyOptionalRune(p, nil, '.', ':', '=', '+', '-', '*', '/', '%', '&', '|', '^') > 0:
			s.skipWS(p, comment.OrAnyWhitespace())
			continue
		}

		parser.NextRune(p)
		s.skipWS(p, comment.OrHorizontalWhitespace())
	}
}

func parenGoCode(p *parser.Parser, s *goCodeState) {
	openingPos := p.Pos()
	openingRune := parser.TryAnyRune(p, '(', '{', '[')
	if openingRune == 0 {
		return
	}

	s.skipWS(p, comment.OrAnyWhitespace())
	goCode(p, s, inParen)

	var singular string
	var closingRune rune
	switch openingRune {
	case '(':
		singular, closingRune = "parenthesis", ')'
	case '{':
		singular, closingRune = "brace", '}'
	case '[':
		singular, closingRune = "bracket", ']'
	default:
		panic("unreachable")
	}

	closingPos := parser.TryRuneAt(p, closingRune)
	if closingPos != nil {
		s.skipWS(p, comment.OrHorizontalWhitespace())
		return
	}

	p.CaptureError(&diagnostic.Diagnostic{
		Message: "go code: unclosed " + singular,
		Primary: quickanno.Expected(p, openingPos, "expected a `"+string(closingRune)+"` for the opening `"+string(openingRune)+"` here"),
	})
}
