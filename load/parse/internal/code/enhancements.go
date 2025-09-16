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

func BlockFunction() parser.Func[*ast.BlockFunction] {
	return func(p *parser.Parser) *ast.BlockFunction {
		block := parser.TryTokenAt(p, "block")
		if block == nil {
			return nil
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		argStart := p.Pos()
		l := parser.Try(p, list.ParenList("argument", "block function arguments", golang.Identifier()))
		argEnd := p.Pos()
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
			excessStart := argStart
			if l.Elems[1] != nil {
				excessStart = l.Elems[1].Start()
			}
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "block function: too many arguments",
				Primary: []diagnostic.Annotation{
					anno.Range(p.File, excessStart, argEnd,
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

		argsStart := p.Pos()
		l := parser.Try(p, list.ParenList("argument", "ternary function arguments", SimpleExpression()))
		argsEnd := p.Pos()
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
			excessStart := argsStart
			if l.Elems[3] != nil {
				excessStart = l.Elems[3].Start()
			}
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "ternary function: too many arguments",
				Primary: []diagnostic.Annotation{
					anno.Range(p.File, excessStart, argsEnd,
						"unexpected arguments: expected only a condition, "+
							"a value for if the condition is true, and a value for if the condition is false"),
				},
			})
		}

		return &t
	}
}
