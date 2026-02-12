package code

import (
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/golang"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

func enhancedStatementNodes() parser.Func[ast.Code] {
	return func(p *parser.Parser) ast.Code {
		// statement can't start with a paren, brace or bracket
		if parser.MatchesAnyRune(p, '(', '{', '[') {
			return nil
		}

		return parseEnhanced(p, enhancedStatementParser)
	}
}

type enhancedParser struct {
	*parser.Parser

	nodes ast.Code

	start, end parser.ByteIndex
	startPos   ast.Position
	hasWS      bool
}

func parseEnhanced(p *parser.Parser, f func(*enhancedParser)) ast.Code {
	gcp := &enhancedParser{
		Parser:   p,
		start:    p.ByteIndex(),
		startPos: *p.PosPtr(),
	}

	f(gcp)
	gcp.commitGoCode()
	parser.RestoreWS(p)

	if len(gcp.nodes) == 0 {
		return nil
	}
	return gcp.nodes
}

func (p *enhancedParser) commitGoCode() {
	if p.start >= p.end {
		return
	}

	pos := p.startPos
	p.nodes = append(p.nodes, &ast.GoCode{
		Code:     p.TokenAt(p.start, p.end),
		Position: &pos,
	})
}

func captureNode(p *enhancedParser, n ast.CodeNode) {
	p.commitGoCode()
	p.nodes = append(p.nodes, n)
	p.skipWS(comment.OrHorizontalWhitespace())
	p.resetStart()
}

type comparableNode interface {
	comparable
	ast.CodeNode
}

func (p *enhancedParser) resetStart() {
	p.start = p.ByteIndex()
	p.startPos = *p.PosPtr()
}

func (p *enhancedParser) skipWS(ws parser.WhitespaceFunc) {
	p.end = p.ByteIndex()
	p.hasWS = parser.TrySkip(p.Parser, ws)
}

func enhancedExpressionParser(p *enhancedParser) {
	for {
		switch {
		case parser.MatchesAnyRune(p.Parser, parser.EOF, ')', '}', ']'):
			return
		case parser.Matches(p.Parser, comment.AndEOS()):
			return
		case parser.MatchesAnyRune(p.Parser, ',', ':'):
			return
		case parser.MatchesToken(p.Parser, "--") || parser.MatchesToken(p.Parser, "++"):
			return
		case parser.Matches(p.Parser, golang.AssignOp()):
			return
		}

		if parser.MatchesAnyRune(p.Parser, '(', '{', '[') {
			r := parser.PeekRune(p.Parser)
			if p.hasWS && (r == '{' || r == '[') {
				// so we don't parse the body after an if or a sole {...} or [...]
				return
			}

			enhancedParenExpressionParser(p)
		} else if ex := parser.TryOptional(p.Parser, Enhancement(), comment.OrHorizontalWhitespace()); ex != nil {
			captureNode(p, ex)
		} else if parser.MatchesRune(p.Parser, '?') { // this is a check for optional chaining
			return
		} else {
			unenhancedRune(p)
		}
	}
}

func enhancedStatementParser(p *enhancedParser) {
	for {
		switch {
		case parser.MatchesAnyRune(p.Parser, parser.EOF, ')', '}', ']'):
			return
		case parser.Matches(p.Parser, comment.AndEOS()):
			return
		}

		if parser.MatchesAnyRune(p.Parser, '(', '{', '[') {
			r := parser.PeekRune(p.Parser)
			if p.hasWS && (r == '{' || r == '[') {
				// so we don't parse the body after an if or a sole {...} or [...]
				return
			}

			enhancedParenExpressionParser(p)
		} else if ex := parser.TryOptional(p.Parser, Enhancement(), comment.OrHorizontalWhitespace()); ex != nil {
			captureNode(p, ex)
		} else {
			unenhancedRune(p)
		}
	}
}

func enhancedParenExpressionParser(p *enhancedParser) {
	openingPos := p.Pos()
	openingRune := parser.TryAnyRune(p.Parser, '(', '{', '[')
	if openingRune == 0 {
		return
	}

	p.skipWS(comment.OrAnyWhitespace())
	inEnhancedParen(p)

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

	closingPos := parser.TryRuneAt(p.Parser, closingRune)
	if closingPos != nil {
		p.skipWS(comment.OrHorizontalWhitespace())
		return
	}

	p.CaptureError(&diagnostic.Diagnostic{
		Message: "go code: unclosed " + singular,
		Primary: quickanno.Expected(p.Parser, openingPos,
			"expected a `"+string(closingRune)+"` for the opening `"+string(openingRune)+"` here"),
	})
}

func inEnhancedParen(p *enhancedParser) {
	for {
		switch {
		case parser.MatchesAnyRune(p.Parser, parser.EOF, ')', '}', ']'):
			return
		}

		if parser.MatchesAnyRune(p.Parser, '(', '{', '[') {
			enhancedParenExpressionParser(p)
		} else if ex := parser.TryOptional(p.Parser, Enhancement(), comment.OrHorizontalWhitespace()); ex != nil {
			captureNode(p, ex)
		} else {
			unenhancedRune(p)
		}
	}
}

func unenhancedRune(p *enhancedParser) {
	switch {
	case parser.TryOptional(p.Parser, golang.RuneLit(), nil):
		p.skipWS(comment.OrHorizontalWhitespace())
	case parser.TryAnyOptionalToken(p.Parser, nil, "==", "!=", ">=", ">", "<=", "<") != "":
		p.skipWS(comment.OrAnyWhitespace())
	case parser.TryAnyOptionalRune(p.Parser, nil, '.', ':', '=', '+', '-', '*', '/', '%', '&', '|', '^') > 0:
		p.skipWS(comment.OrAnyWhitespace())
	default:
		parser.NextRune(p.Parser)
		p.skipWS(comment.OrHorizontalWhitespace())
	}
}
