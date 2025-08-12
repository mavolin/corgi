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
	"github.com/mavolin/corgi/v2/load/parse/internal/whitespace"
)

type Options uint8

const (
	// Regular applies no special options.
	Regular Options = iota
	// Statements parses statements, not expressions.
	Statements Options = 1 << iota
	// FirstParen only parses until the first parenthesis is closed.
	FirstParen Options = 1 << iota
)

func (o Options) statements() bool { return o&Statements != 0 }
func (o Options) firstParen() bool { return o&FirstParen != 0 }

func Code(o Options) parser.Func[ast.Code] {
	return func(p *parser.Parser) ast.Code {
		if zc := parser.Try(p, ZeroCoalescing()); zc != nil {
			return ast.Code{zc}
		}

		return parser.Try(p, NonZCCode(o))
	}
}

func NonZCCode(o Options) parser.Func[ast.Code] {
	return func(p *parser.Parser) ast.Code {
		var c ast.Code
		for {
			n := parser.Try(p, nonZCNode(o))
			if n == nil {
				break
			}
			if c == nil {
				c = n.Nodes
			} else {
				c = append(c, n.Nodes...)
			}
			if n.Stop {
				break
			}
			parser.TrySkip(p, comment.OrHorizontalWhitespace())
		}
		if len(c) == 0 {
			return nil
		}

		c = slices.Clip(c)
		return c
	}
}

type codeResult struct {
	Nodes []ast.CodeNode
	Stop  bool
}

// nonZCNode tries to capture as few as possible []ast.CodeNode.
// See the doc of [GoCode] on why it may return more than one node.
func nonZCNode(o Options) parser.Func[*codeResult] {
	return func(p *parser.Parser) *codeResult {
		if gc := parser.Try(p, goCode(o)); gc != nil {
			return gc
		} else if bf := parser.Try(p, BlockFunction()); bf != nil {
			return &codeResult{Nodes: []ast.CodeNode{bf}}
		} else if s := parser.Try(p, String()); s != nil {
			return &codeResult{Nodes: []ast.CodeNode{s}}
		} else if t := parser.Try(p, Ternary()); t != nil {
			return &codeResult{Nodes: []ast.CodeNode{t}}
		}
		return nil
	}
}

// GoCode, if it matches, returns a slice of one or more Expression[]ast.CodeNode, the
// first of which is a pointer to a [ast.GoCode].
// The only case in which more than one ExpressionNode is returned, is when
// the parsed code contains corgi language extensions within parenthesis.
func GoCode(o Options) parser.Func[[]ast.CodeNode] {
	return func(p *parser.Parser) []ast.CodeNode {
		res := parser.Try(p, goCode(o))
		if res == nil {
			return nil
		}
		return res.Nodes
	}
}

func goCode(o Options) parser.Func[*codeResult] {
	return func(p *parser.Parser) *codeResult {
		var c *ast.GoCode

		var exps []ast.CodeNode

		start := p.Index()
		type paren struct {
			opening byte
			pos     ast.Position
		}
		var parenStack []paren
		var stop bool

		var canSkipAnyWS bool
		for {
			end := p.Index()
			var hasWS bool
			if canSkipAnyWS || len(parenStack) > 0 {
				hasWS = parser.TrySkip(p, comment.OrAnyWhitespace())
				canSkipAnyWS = false
			} else {
				hasWS = parser.TrySkip(p, comment.OrHorizontalWhitespace())
			}

			if len(parenStack) == 0 && parser.MatchesAnyRune(p, ';') {
				stop = true
				break
			} else if len(parenStack) == 0 && !o.statements() {
				if parser.MatchesAnyRune(p, ',', ':') { //nolint:gocritic
					stop = true
					break
				} else if parser.MatchesToken(p, "--") || parser.MatchesToken(p, "++") {
					stop = true
					break
				} else if parser.Matches(p, golang.AssignOp()) {
					stop = true
					break
				}
			}

			pos := p.Pos()
			if parser.MatchesAnyRune(p, '(', '{', '[') { //nolint:gocritic
				r := parser.PeekRune(p)
				if (hasWS || (len(exps) == 0 && start == end)) && (r == '{' || r == '[') {
					// so we don't parse the body after an if
					stop = true
					break
				}

				if c == nil {
					c = &ast.GoCode{Position: p.PosPtr()}
				}

				parser.NextRune(p)
				parenStack = append(parenStack, paren{opening: byte(r), pos: pos})
				continue
			} else if parser.MatchesAnyRune(p, ')', '}', ']') {
				if len(parenStack) == 0 {
					stop = true
					break
				}

				if c == nil {
					c = &ast.GoCode{Position: p.PosPtr()}
				}

				closing := parser.NextRune(p)
				open := parenStack[len(parenStack)-1]
				switch {
				case open.opening == '(' && closing == ')',
					open.opening == '{' && closing == '}',
					open.opening == '[' && closing == ']':
					parenStack = parenStack[:len(parenStack)-1]
					if o.firstParen() && len(parenStack) == 0 {
						break
					}
				case closing == ')':
					p.CaptureError(&diagnostic.Diagnostic{
						Message: "go code: mismatched parentheses",
						Primary: []diagnostic.Annotation{
							anno.Position(p.File, pos, "this closing parenthesis does not have a matching opening parenthesis"),
						},
					})
				case closing == '}':
					p.CaptureError(&diagnostic.Diagnostic{
						Message: "go code: mismatched braces",
						Primary: []diagnostic.Annotation{
							anno.Position(p.File, pos, "this closing brace does not have a matching opening brace"),
						},
					})
				case closing == ']':
					p.CaptureError(&diagnostic.Diagnostic{
						Message: "go code: mismatched brackets",
						Primary: []diagnostic.Annotation{
							anno.Position(p.File, pos, "this closing bracket does not have a matching opening bracket"),
						},
					})
				}
				continue
			} else if parser.MatchesToken(p, "block") && parser.Matches(p, BlockFunction()) {
				if len(parenStack) > 0 {
					c.Code = p.AST.Raw[start:end]
					exps = append(exps, c, parser.Try(p, BlockFunction()))
					parser.TrySkip(p, comment.OrHorizontalWhitespace())
					start = p.Index()
					c = nil
					continue
				}
				break
			} else if parser.MatchesAnyRune(p, '"', '`') {
				if len(parenStack) > 0 {
					c.Code = p.AST.Raw[start:end]
					exps = append(exps, c, parser.Try(p, String()))
					parser.TrySkip(p, comment.OrHorizontalWhitespace())
					start = p.Index()
					c = nil
					continue
				}
				break
			} else if parser.MatchesAnyRune(p, '?') {
				if len(parenStack) > 0 {
					c.Code = p.AST.Raw[start:end]
					t := parser.Try(p, Ternary())
					if t != nil {
						exps = append(exps, c, t)
						parser.TrySkip(p, comment.OrHorizontalWhitespace())
						start = p.Index()
						c = nil
						continue
					}
				}
				break
			} else if parser.MatchesAnyRune(p, parser.EOF) {
				stop = true
				break
			}

			if c == nil {
				c = &ast.GoCode{Position: p.PosPtr()}
			}
			switch {
			case parser.TryOptional(p, golang.RuneLit(), nil) != "":
				continue
			case parser.TryAnyOptionalToken(p, nil, "==", "!=", ">", ">=", "<", "<=") != "":
				canSkipAnyWS = true
				continue
			case parser.TryAnyOptionalRune(p, nil, '.', ':', '=', '+', '-', '*', '/', '%', '&', '|', '^') > 0:
				canSkipAnyWS = true
				continue
			}

			if parser.MatchesAnyRune(p, whitespace.Runes...) {
				stop = true
				break
			}
			parser.NextRune(p)
		}

		parser.RestoreWS(p)
		if start == p.Index() {
			if len(exps) == 0 {
				return nil
			}
		} else {
			c.Code = p.AST.Raw[start:p.Index()]
			exps = append(exps, c)
		}

		for _, open := range slices.Backward(parenStack) {
			switch open.opening {
			case '(':
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "go code: unclosed parenthesis",
					Primary: quickanno.Expected(p, open.pos, "expected a `)` for the opening `(` here"),
				})
			case '{':
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "go code: unclosed brace",
					Primary: quickanno.Expected(p, open.pos, "expected a `}` for the opening `{` here"),
				})
			case '[':
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "go code: unclosed bracket",
					Primary: quickanno.Expected(p, open.pos, "expected a `]` for the opening `[` here"),
				})
			}
		}

		return &codeResult{Nodes: exps, Stop: stop}
	}
}

func BlockFunction() parser.Func[*ast.BlockFunction] {
	return func(p *parser.Parser) *ast.BlockFunction {
		block := parser.TryTokenAt(p, "block")
		if block == nil {
			return nil
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		l := parser.Try(p, list.ParenList("argument", "block function arguments", golang.Identifier()))
		if l == nil {
			return nil
		}

		var bf ast.BlockFunction
		bf.Block = block
		bf.LParen, bf.RParen = l.Open, l.Close

		if len(l.Elems) > 0 {
			bf.BlockName = l.Elems[0]
		}
		if len(l.Elems) > 1 {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "block function: too many arguments",
				Primary: []diagnostic.Annotation{
					anno.Range(p.File, l.Elems[1].Start(), l.Elems[len(l.Elems)-1].End(),
						"unexpected arguments, expected only a single block name"),
				},
			})
		}

		return &bf
	}
}

func Ternary() parser.Func[*ast.Ternary] {
	return func(p *parser.Parser) *ast.Ternary {
		questionMark := parser.TryRuneAt(p, '?')
		if questionMark == nil {
			return nil
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		l := parser.Try(p, list.ParenList("argument", "ternary function arguments", NonZCExpression(Regular)))
		if l == nil {
			return nil
		}

		var t ast.Ternary
		t.QuestionMark = questionMark
		t.LParen, t.RParen = l.Open, l.Close

		switch {
		case len(l.Elems) == 0:
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "ternary function: missing arguments",
				Primary: quickanno.Expected(p, *l.Open,
					"a condition, a value for if the condition is true, and a value for if the condition is false"),
				Examples: []diagnostic.Example{{Example: "?(condition, ifTrue, ifFalse)"}},
			})
		case len(l.Elems) == 1:
			t.Condition = l.Elems[0]

			pos := *l.Open
			if l.Close != nil {
				pos = *l.Close
			}

			p.CaptureError(&diagnostic.Diagnostic{
				Message: "ternary function: missing if-true and if-false values",
				Primary: quickanno.Expected(p, pos,
					"a value for if the condition is true and a value for if the condition is false, after the condition"),
				Examples: []diagnostic.Example{{Example: "?(condition, ifTrue, ifFalse)"}},
			})
		case len(l.Elems) == 2:
			t.Condition, t.TrueVal = l.Elems[0], l.Elems[1]

			pos := *l.Open
			if l.Close != nil {
				pos = *l.Close
			}
			p.CaptureError(&diagnostic.Diagnostic{
				Message:  "ternary function: missing if-false value",
				Primary:  quickanno.Expected(p, pos, "a value for if the condition is false"),
				Examples: []diagnostic.Example{{Example: "?(condition, ifTrue, ifFalse)"}},
			})
		default:
			t.Condition, t.TrueVal, t.FalseVal = l.Elems[0], l.Elems[1], l.Elems[2]
		}

		if len(l.Elems) > 3 {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "ternary function: too many arguments",
				Primary: []diagnostic.Annotation{
					anno.Range(p.File, l.Elems[3].Start(), l.Elems[len(l.Elems)-1].End(),
						"unexpected arguments: expected only a condition, "+
							"a value for if the condition is true, and a value for if the condition is false"),
				},
			})
		}

		return &t
	}
}
