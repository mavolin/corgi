package code

import (
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/golang"
	"github.com/mavolin/corgi/v2/load/parse/internal/list"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

var componentCall parser.Func[*ast.ComponentCall]

func SetComponentCall(f parser.Func[*ast.ComponentCall]) {
	componentCall = f
}

type Options uint8

const (
	// Regular applies no special options.
	Regular Options = iota
	// Statements parses statements, not expressions.
	Statements Options = 1 << iota
	// FirstParen only parses until the first parenthesis is closed.
	FirstParen Options = 1 << iota
	// inParen indicates we are inside parentheses/brackets/braces.
	// This relaxes some parsing rules, e.g. allows commas.
	inParen Options = 1 << iota
)

func (o Options) statements() bool { return o&Statements != 0 }
func (o Options) firstParen() bool { return o&FirstParen != 0 }
func (o Options) inParen() bool    { return o&inParen != 0 }

func Code(o Options) parser.Func[ast.Code] {
	return func(p *parser.Parser) ast.Code {
		if zc := parser.Try(p, ZeroCoalescing()); zc != nil {
			return ast.Code{zc}
		} else if cc := parser.Try(p, componentCall); cc != nil {
			return ast.Code{cc}
		}

		return parser.Try(p, GoCode(o))
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
