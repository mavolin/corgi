package code

import (
	"slices"

	"github.com/mavolin/corgi/v2/fancyerr"
	"github.com/mavolin/corgi/v2/fancyerr/anno"
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/golang"
	"github.com/mavolin/corgi/v2/load/parse/internal/list"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

func Code(statement bool) parser.Func[ast.Code] {
	return func(p *parser.Parser) (ast.Code, *fancyerr.Error) {
		if zc, ok := parser.TryOk(p, ZeroCoalescing()); ok {
			return ast.Code{zc}, nil
		}

		return parser.Try(p, NonZCCode(statement))
	}
}

func NonZCCode(statement bool) parser.Func[ast.Code] {
	return func(p *parser.Parser) (ast.Code, *fancyerr.Error) {
		c := make(ast.Code, 0, 24)
		for {
			n, ok := parser.TryOk(p, nonZCNode(statement))
			if !ok {
				break
			}
			c = append(c, n...)
			parser.TrySkip(p, comment.OrHorizontalWhitespace())
		}
		if len(c) == 0 {
			return nil, &fancyerr.Error{
				Message: "missing code node",
				Primary: quickanno.Expected(p, p.Pos(), "a code node"),
			}
		}

		c = slices.Clip(c)
		return c, nil
	}
}

// nonZCNode tries to capture as few as possible []ast.CodeNode.
// See the doc of [GoCode] on why it may return more than one node.
func nonZCNode(statement bool) parser.Func[[]ast.CodeNode] {
	return func(p *parser.Parser) ([]ast.CodeNode, *fancyerr.Error) {
		if gc, ok := parser.TryOk(p, GoCode(statement)); ok {
			return gc, nil
		} else if bf, ok := parser.TryOk(p, BlockFunction()); ok {
			return []ast.CodeNode{bf}, nil
		} else if s, ok := parser.TryOk(p, String()); ok {
			return []ast.CodeNode{s}, nil
		} else if t, ok := parser.TryOk(p, Ternary()); ok {
			return []ast.CodeNode{t}, nil
		}
		return nil, &fancyerr.Error{
			Message: "missing code node",
			Primary: quickanno.Expected(p, p.Pos(), "a code node"),
		}
	}
}

// GoCode, if it matches, returns a slice of one or more Expression[]ast.CodeNode, the
// first of which is a pointer to a [ast.GoCode].
// The only case in which more than one ExpressionNode is returned, is when
// the parsed code contains corgi language extensions within parenthesis.
func GoCode(statement bool) parser.Func[[]ast.CodeNode] {
	return func(p *parser.Parser) ([]ast.CodeNode, *fancyerr.Error) {
		ns, err := parser.Try(p, goCode(false, statement))
		if err != nil {
			return nil, err
		}
		return ns, nil
	}
}

func goCode(tilParenClose, statement bool) parser.Func[[]ast.CodeNode] {
	return func(p *parser.Parser) ([]ast.CodeNode, *fancyerr.Error) {
		c := &ast.GoCode{Position: p.Pos()}

		exps := make([]ast.CodeNode, 0, 8)

		start := p.Index()
		type paren struct {
			open byte
			pos  ast.Position
		}
		parenStack := make([]paren, 0, 12)

		var canSkipAnyWS bool
		for {
			state := p.CloneState()
			if canSkipAnyWS {
				parser.TrySkip(p, comment.OrAnyWhitespace())
				canSkipAnyWS = false
			} else {
				parser.TrySkip(p, comment.OrHorizontalWhitespace())
			}
			parser.CommitWS(p)

			pos := p.Pos()
			if r := parser.TryAnyRune(p, '(', '{', '['); r > 0 {
				parenStack = append(parenStack, paren{open: byte(r), pos: pos})
				continue
			} else if parser.MatchesAnyRune(p, ')', '}', ']') {
				if len(parenStack) == 0 {
					break
				}
				close := parser.TryRunePredicate(p, func(rune) bool { return true })
				open := parenStack[len(parenStack)-1]
				switch {
				case open.open == '(' && close == ')',
					open.open == '{' && close == '}',
					open.open == '[' && close == ']':
					parenStack = parenStack[:len(parenStack)-1]
					if tilParenClose && len(parenStack) == 0 {
						break
					}
				case close == ')':
					p.CaptureError(&fancyerr.Error{
						Message: "go code: mismatched parentheses",
						Primary: []fancyerr.Annotation{
							anno.Position(p.File, pos, "this closing parenthesis does not have a matching opening parenthesis"),
						},
					})
				case close == '}':
					p.CaptureError(&fancyerr.Error{
						Message: "go code: mismatched braces",
						Primary: []fancyerr.Annotation{
							anno.Position(p.File, pos, "this closing brace does not have a matching opening brace"),
						},
					})
				case close == ']':
					p.CaptureError(&fancyerr.Error{
						Message: "go code: mismatched brackets",
						Primary: []fancyerr.Annotation{
							anno.Position(p.File, pos, "this closing bracket does not have a matching opening bracket"),
						},
					})
				}
				continue
			} else if _, ok := parser.TryOk(p, golang.RuneLit()); ok {
				continue
			} else if parser.MatchesToken(p, "block") && parser.Matches(p, BlockFunction()) {
				if len(parenStack) > 0 {
					c.Code = p.Raw[start:p.Index()]
					exps = append(exps, c, parser.Must(p, BlockFunction()))
					start = p.Index()
					c = &ast.GoCode{Position: p.Pos()}
					continue
				}
				break
			} else if parser.MatchesAnyRune(p, '"', '`') {
				if len(parenStack) > 0 {
					c.Code = p.Raw[start:p.Index()]
					exps = append(exps, c, parser.Must(p, String()))
					start = p.Index()
					c = &ast.GoCode{Position: p.Pos()}
					continue
				}
				break
			} else if parser.MatchesAnyRune(p, '?') {
				if len(parenStack) > 0 {
					c.Code = p.Raw[start:p.Index()]
					t, ok := parser.TryOk(p, Ternary())
					if ok {
						exps = append(exps, c, t)
						start = p.Index()
						c = &ast.GoCode{Position: p.Pos()}
						continue
					}
				}
			} else if parser.MatchesAnyRune(p, parser.EOF) {
				break
			}
			if len(parenStack) == 0 {
				if !statement &&
					(parser.MatchesAnyRune(p, ',') || parser.MatchesToken(p, ":=") ||
						(!parser.MatchesToken(p, "==") && parser.Matches(p, golang.AssignOp()))) {
					p.RestoreState(state)
					break
				} else if parser.MatchesAnyRune(p, ';', '?') || parser.MatchesToken(p, "--") || parser.MatchesToken(p, "++") {
					p.RestoreState(state)
					break
				} else if parser.MatchesWS(p, comment.AndEOS()) {
					p.RestoreState(state)
					break
				}
			}

			r := parser.TryRunePredicate(p, func(r rune) bool { return true })
			if r == '.' {
				canSkipAnyWS = true
			}
		}
		if start == p.Index() {
			if len(exps) == 0 {
				return nil, &fancyerr.Error{
					Message: "missing go code",
					Primary: quickanno.Expected(p, p.Pos(), "go code"),
				}
			}
		} else {
			c.Code = p.Raw[start:p.Index()]
			exps = append(exps, c)
		}
		exps = slices.Clip(exps)

		for _, open := range slices.Backward(parenStack) {
			switch open.open {
			case '(':
				p.CaptureError(&fancyerr.Error{
					Message: "go code: unclosed parenthesis",
					Primary: quickanno.Expected(p, open.pos, "expected a `)` for the opening `(` here"),
				})
			case '{':
				p.CaptureError(&fancyerr.Error{
					Message: "go code: unclosed brace",
					Primary: quickanno.Expected(p, open.pos, "expected a `}` for the opening `{` here"),
				})
			case '[':
				p.CaptureError(&fancyerr.Error{
					Message: "go code: unclosed bracket",
					Primary: quickanno.Expected(p, open.pos, "expected a `]` for the opening `[` here"),
				})
			}
		}

		return exps, nil
	}
}

func BlockFunction() parser.Func[*ast.BlockFunction] {
	return func(p *parser.Parser) (*ast.BlockFunction, *fancyerr.Error) {
		bf := &ast.BlockFunction{
			Position: p.Pos(),
		}

		if !parser.TryToken(p, "block") {
			return nil, &fancyerr.Error{
				Message: "missing block function",
				Primary: quickanno.Expected(p, p.Pos(), "a block function"),
			}
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		l, err := parser.Try(p, list.ParenList("block function arguments", golang.Identifier()))
		if err != nil {
			return nil, err
		}

		bf.LParen, bf.RParen = &l.Open, l.Close

		if len(l.Elems) == 0 {
			p.CaptureError(&fancyerr.Error{
				Message: "block function: missing block name",
				Primary: quickanno.Expected(p, l.Open, "a block name"),
			})
		} else {
			bf.Block = l.Elems[0]
			if len(l.Elems) > 1 {
				p.CaptureError(&fancyerr.Error{
					Message: "block function: too many arguments",
					Primary: []fancyerr.Annotation{
						anno.Range(p.File, l.Elems[1].Pos(), l.Elems[len(l.Elems)-1].End(),
							"unexpected arguments, expected only a single block name"),
					},
				})
			}
		}

		return bf, nil
	}
}

func Ternary() parser.Func[*ast.Ternary] {
	return func(p *parser.Parser) (*ast.Ternary, *fancyerr.Error) {
		t := &ast.Ternary{
			QuestionMark: p.Pos(),
		}
		if !parser.TryRune(p, '?') {
			return nil, &fancyerr.Error{
				Message: "missing ternary function",
				Primary: quickanno.Expected(p, p.Pos(), "a ternary function"),
				Examples: []fancyerr.Example{
					{
						Example: "?(condition, ifTrue, ifFalse)",
					},
				},
			}
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		l, err := parser.Try(p, list.ParenList("ternary function arguments", NonZCExpression()))
		if err != nil {
			return nil, &fancyerr.Error{
				Message: "missing ternary function",
				Primary: quickanno.Expected(p, p.Pos(), "a ternary function"),
				Examples: []fancyerr.Example{
					{
						Example: "?(condition, ifTrue, ifFalse)",
					},
				},
			}
		}

		t.LParen, t.RParen = &l.Open, l.Close
		switch {
		case len(l.Elems) == 0:
			p.CaptureError(&fancyerr.Error{
				Message: "ternary function: missing arguments",
				Primary: quickanno.Expected(p, l.Open,
					"a condition, a value for if the condition is true, and a value for if the condition is false"),
				Examples: []fancyerr.Example{
					{
						Example: "?(condition, ifTrue, ifFalse)",
					},
				},
			})
		case len(l.Elems) == 1:
			t.Condition = l.Elems[0]

			pos := l.Open
			if l.Close != nil {
				pos = *l.Close
			}

			p.CaptureError(&fancyerr.Error{
				Message: "ternary function: missing if-true and if-false values",
				Primary: quickanno.Expected(p, pos,
					"a value for if the condition is true and a value for if the condition is false, after the condition"),
				Examples: []fancyerr.Example{
					{
						Example: "?(condition, ifTrue, ifFalse)",
					},
				},
			})
		case len(l.Elems) == 2:
			t.Condition, t.TrueVal = l.Elems[0], l.Elems[1]

			pos := l.Open
			if l.Close != nil {
				pos = *l.Close
			}
			p.CaptureError(&fancyerr.Error{
				Message: "ternary function: missing if-false value",
				Primary: quickanno.Expected(p, pos, "a value for if the condition is false"),
				Examples: []fancyerr.Example{
					{
						Example: "?(condition, ifTrue, ifFalse)",
					},
				},
			})
		default:
			t.Condition, t.TrueVal, t.FalseVal = l.Elems[0], l.Elems[1], l.Elems[2]
		}

		if len(l.Elems) > 3 {
			p.CaptureError(&fancyerr.Error{
				Message: "ternary function: too many arguments",
				Primary: []fancyerr.Annotation{
					anno.Range(p.File, l.Elems[3].Pos(), l.Elems[len(l.Elems)-1].End(),
						"unexpected arguments: expected only a condition, "+
							"a value for if the condition is true, and a value for if the condition is false"),
				},
			})
		}

		return t, nil
	}
}
