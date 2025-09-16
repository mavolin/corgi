package control

import (
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/body"
	"github.com/mavolin/corgi/v2/load/parse/internal/code"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
)

func Conditional() parser.Func[*ast.Conditional] {
	return func(p *parser.Parser) *ast.Conditional {
		ifNode := parser.Try(p, If())
		if ifNode == nil {
			return nil
		}

		var c ast.Conditional
		c.If = ifNode
		c.ElseIfs = parser.Collect(p, ElseIf(), comment.OrAnyWhitespace())
		parser.TrySkip(p, comment.OrAnyWhitespace())
		c.Else = parser.Try(p, Else())

		return &c
	}
}

func If() parser.Func[*ast.If] {
	return func(p *parser.Parser) *ast.If {
		ifKw := parser.TryKeywordAt(p, "if", comment.OrAnyWhitespace())
		if ifKw == nil {
			return nil
		}

		var i ast.If
		i.If = ifKw
		i.Header = parser.Try(p, IfHeader())
		if i.Header == nil {
			return nil
		}
		parser.TrySkip(p, comment.OrHorizontalWhitespace())
		i.Then = parser.Try(p, body.Body())
		if i.Then == nil {
			return nil
		}

		return &i
	}
}

func ElseIf() parser.Func[*ast.ElseIf] {
	return func(p *parser.Parser) *ast.ElseIf {
		elseKw := parser.TryKeywordAt(p, "else", comment.OrAnyWhitespace())
		ifKw := parser.TryKeywordAt(p, "if", comment.OrAnyWhitespace())
		if elseKw == nil || ifKw == nil {
			return nil
		}

		var ei ast.ElseIf
		ei.Else = elseKw
		ei.If = ifKw
		ei.Header = parser.Try(p, IfHeader())
		if ei.Header == nil {
			return nil
		}
		parser.TrySkip(p, comment.OrHorizontalWhitespace())
		ei.Then = parser.Try(p, body.Body())
		if ei.Then == nil {
			return nil
		}

		return &ei
	}
}

func Else() parser.Func[*ast.Else] {
	return func(p *parser.Parser) *ast.Else {
		elseKw := parser.TryKeywordAt(p, "else", comment.OrAnyWhitespace())
		if elseKw == nil {
			return nil
		}

		var e ast.Else
		e.Else = elseKw
		e.Then = parser.Try(p, body.Body())
		if e.Then == nil {
			return nil
		}
		return &e
	}
}

func IfHeader() parser.Func[*ast.IfHeader] {
	return func(p *parser.Parser) *ast.IfHeader {
		return parser.TryInOrder(p, ifHeaderWithStatement(), ifHeaderWithoutStatement())
	}
}

func ifHeaderWithStatement() parser.Func[*ast.IfHeader] {
	return func(p *parser.Parser) *ast.IfHeader {
		stmt := parser.Try(p, code.SimpleStatement())
		if stmt == nil {
			return nil
		}

		if !parser.Try(p, comment.AndEOS()) {
			return nil
		}
		parser.TrySkip(p, comment.OrAnyWhitespace())

		var h ast.IfHeader
		h.Statement = stmt

		h.Condition = parser.Try(p, code.Expression())
		if h.Condition == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "if header: missing condition",
				Primary: quickanno.Expected(p, p.Pos(), "a condition expression"),
				Secondary: []diagnostic.Annotation{
					anno.Node(p.File, h.Statement, "because of the statement here"),
				},
			})
		}

		return &h
	}
}

func ifHeaderWithoutStatement() parser.Func[*ast.IfHeader] {
	return func(p *parser.Parser) *ast.IfHeader {
		cond := parser.Try(p, code.Expression())
		if cond == nil {
			return nil
		}

		return &ast.IfHeader{Condition: cond}
	}
}
