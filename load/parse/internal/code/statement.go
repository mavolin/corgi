package code

import (
	"fmt"
	"slices"

	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/golang"
	"github.com/mavolin/corgi/v2/load/parse/internal/list"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/unexpected"
)

func Statement() parser.Func[*ast.Statement] {
	return func(p *parser.Parser) *ast.Statement {
		s := parser.Try(p, parsedStatement())
		if s != nil && s.Parsed != nil {
			return s
		}

		// parsedStatement might also have returned an expression belonging to
		// the start of the statement

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		extra := parser.Try(p, enhancedStatementNodes())
		if s == nil || s.Nodes == nil {
			if len(extra) == 0 {
				return nil
			}
			return &ast.Statement{Nodes: extra}
		}

		if len(extra) == 0 {
			return s
		}
		s.Nodes = slices.Grow(s.Nodes, len(extra))
		s.Nodes = append(s.Nodes, extra...)
		return s
	}
}

func ParsedStatement() parser.Func[*ast.Statement] {
	return func(p *parser.Parser) *ast.Statement {
		s := parser.Try(p, parsedStatement())
		if s == nil || s.Parsed == nil {
			return nil
		}
		return s
	}
}

func parsedStatement() parser.Func[*ast.Statement] {
	return func(p *parser.Parser) *ast.Statement {
		start := p.Pos()

		if r := parser.Try(p, Return()); r != nil {
			return &ast.Statement{
				Nodes:  ReturnAsCode(r),
				Parsed: r,
			}
		} else if b := parser.Try(p, Break()); b != nil {
			return &ast.Statement{
				Nodes:  BreakAsCode(b),
				Parsed: b,
			}
		} else if c := parser.Try(p, Continue()); c != nil {
			return &ast.Statement{
				Nodes:  ContinueAsCode(c),
				Parsed: c,
			}
		} else if f := parser.Try(p, Fallthrough()); f != nil {
			return &ast.Statement{
				Nodes:  FallthroughAsCode(f),
				Parsed: f,
			}
		} else if d := parser.Try(p, Defer()); d != nil {
			return &ast.Statement{
				Nodes:  DeferAsCode(d),
				Parsed: d,
			}
		} else if cd := parser.Try(p, ConstDeclaration()); cd != nil {
			return &ast.Statement{
				Nodes:  ConstDeclarationAsCode(start, cd),
				Parsed: cd,
			}
		} else if vd := parser.Try(p, VarDeclaration()); vd != nil {
			return &ast.Statement{
				Nodes:  VarDeclarationAsCode(start, vd),
				Parsed: vd,
			}
		} else if l := parser.Try(p, Label()); l != nil {
			return &ast.Statement{
				Nodes:  LabelAsCode(l),
				Parsed: l,
			}
		}

		beforeExpr := p.CloneState()
		e := parser.Try(p, Expression())
		if e == nil {
			return nil
		}
		afterExpr := p.CloneState()

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		if zca := parser.TryOptional(p, zeroCoalescingAssignment(e), nil); zca != nil {
			return &ast.Statement{
				Nodes:  ZeroCoalescingAssignmentAsCode(zca),
				Parsed: zca,
			}
		} else if incDec := parser.TryOptional(p, incDec(e), nil); incDec != nil {
			return &ast.Statement{
				Nodes:  IncDecAsCode(incDec),
				Parsed: incDec,
			}
		} else if a := parser.Try(p, assignment(e)); a != nil {
			return &ast.Statement{
				Nodes:  AssignmentAsCode(e.Start(), a),
				Parsed: a,
			}
		}

		p.RestoreState(beforeExpr)
		if svd := parser.Try(p, ShortVarDeclaration()); svd != nil {
			return &ast.Statement{
				Nodes:  ShortVarDeclarationAsCode(e.Start(), svd),
				Parsed: svd,
			}
		}

		p.RestoreState(afterExpr)
		return &ast.Statement{Nodes: e.Nodes}
	}
}

func SimpleStatement() parser.Func[*ast.SimpleStatement] {
	return func(p *parser.Parser) *ast.SimpleStatement {
		ss, ok := parsedSimpleStatement(p)
		if ss != nil && ss.Parsed != nil {
			return ss
		}

		if !ok { // a non-simple statement is ahead
			return nil
		}

		// parsedSimpleStatement might also have returned an expression
		// belonging to the start of the statement

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		extra := parser.Try(p, enhancedStatementNodes())
		if ss == nil || ss.Nodes == nil {
			if len(extra) == 0 {
				return nil
			}
			return &ast.SimpleStatement{Nodes: extra}
		}

		if len(extra) == 0 {
			return ss
		}
		ss.Nodes = slices.Grow(ss.Nodes, len(extra))
		ss.Nodes = append(ss.Nodes, extra...)
		return ss
	}
}

func ParsedSimpleStatement() parser.Func[*ast.SimpleStatement] {
	return func(p *parser.Parser) *ast.SimpleStatement {
		ss, _ := parsedSimpleStatement(p)
		if ss == nil || ss.Parsed == nil {
			return nil
		}
		return ss
	}
}

func parsedSimpleStatement(p *parser.Parser) (*ast.SimpleStatement, bool) {
	s := parser.Try(p, parsedStatement())
	if s == nil {
		return nil, true
	} else if s.Parsed == nil {
		return &ast.SimpleStatement{Nodes: s.Nodes}, true
	}

	pss, _ := s.Parsed.(ast.ParsedSimpleStatement)
	if pss == nil {
		return nil, false
	}
	return &ast.SimpleStatement{
		Nodes:  s.Nodes,
		Parsed: pss,
	}, true
}

func Return() parser.Func[*ast.Return] {
	return func(p *parser.Parser) *ast.Return {
		ret := parser.TryKeywordAt(p, "return")
		if ret == nil {
			return nil
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		var r ast.Return
		r.Return = ret

		r.Error = parser.Try(p, Expression())
		return &r
	}
}

func ReturnAsCode(r *ast.Return) ast.Code {
	if r.Error == nil {
		return ast.Code{&ast.GoCode{Code: "return", Position: r.Return}}
	}

	c := make(ast.Code, 1+len(r.Error.Nodes))
	c[0] = &ast.GoCode{Code: "return", Position: r.Return}
	copy(c[1:], r.Error.Nodes)
	return c
}

func Break() parser.Func[*ast.Break] {
	return func(p *parser.Parser) *ast.Break {
		brk := parser.TryKeywordAt(p, "break")
		if brk == nil {
			return nil
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		var b ast.Break
		b.Break = brk

		b.Label = parser.Try(p, golang.Identifier())
		return &b
	}
}

func BreakAsCode(b *ast.Break) ast.Code {
	if b.Label == nil {
		return ast.Code{&ast.GoCode{Code: "break", Position: b.Break}}
	}
	return ast.Code{
		&ast.GoCode{Code: "break", Position: b.Break},
		&ast.GoCode{Code: b.Label.Name, Position: b.Label.Position},
	}
}

func Continue() parser.Func[*ast.Continue] {
	return func(p *parser.Parser) *ast.Continue {
		_continue := parser.TryKeywordAt(p, "continue")
		if _continue == nil {
			return nil
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		var c ast.Continue
		c.Continue = _continue

		c.Label = parser.Try(p, golang.Identifier())

		return &c
	}
}

func ContinueAsCode(c *ast.Continue) ast.Code {
	if c.Label == nil {
		return ast.Code{&ast.GoCode{Code: "continue", Position: c.Continue}}
	}
	return ast.Code{
		&ast.GoCode{Code: "continue", Position: c.Continue},
		&ast.GoCode{Code: c.Label.Name, Position: c.Label.Position},
	}
}

func Fallthrough() parser.Func[*ast.Fallthrough] {
	return func(p *parser.Parser) *ast.Fallthrough {
		_fallthrough := parser.TryKeywordAt(p, "fallthrough")
		if _fallthrough == nil {
			return nil
		}

		return &ast.Fallthrough{Fallthrough: _fallthrough}
	}
}

func FallthroughAsCode(f *ast.Fallthrough) ast.Code {
	return ast.Code{&ast.GoCode{Code: "fallthrough", Position: f.Fallthrough}}
}

func Defer() parser.Func[*ast.Defer] {
	return func(p *parser.Parser) *ast.Defer {
		deferKeyword := parser.TryKeywordAt(p, "defer")
		if deferKeyword == nil {
			return nil
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())

		var d ast.Defer
		d.Defer = deferKeyword

		pos := p.Pos()
		d.Expression = parser.Try(p, Expression())
		if d.Expression == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "defer: missing expression",
				Primary: quickanno.Expected(p, pos, "an expression after the `defer` keyword"),
			})
		}
		return &d
	}
}

func DeferAsCode(d *ast.Defer) ast.Code {
	if d.Expression == nil {
		return ast.Code{&ast.GoCode{Code: "defer", Position: d.Defer}}
	}

	c := make(ast.Code, 1+len(d.Expression.Nodes))
	c[0] = &ast.GoCode{Code: "defer", Position: d.Defer}
	copy(c[1:], d.Expression.Nodes)
	return c
}

func ZeroCoalescingAssignment() parser.Func[*ast.ZeroCoalescingAssignment] {
	return func(p *parser.Parser) *ast.ZeroCoalescingAssignment {
		valueExpr := parser.Try(p, Expression())
		if valueExpr == nil {
			return nil
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		return parser.Try(p, zeroCoalescingAssignment(valueExpr))
	}
}

func zeroCoalescingAssignment(valueExpr *ast.Expression) parser.Func[*ast.ZeroCoalescingAssignment] {
	return func(p *parser.Parser) *ast.ZeroCoalescingAssignment {
		var zca ast.ZeroCoalescingAssignment
		zca.ValueExpression = valueExpr

		parser.TrySkip(p, comment.OrHorizontalWhitespace())
		zca.VarComma = parser.TryOptionalRuneAt(p, ',', comment.OrAnyWhitespace())
		if zca.VarComma != nil {
			zca.OkExpression = parser.Try(p, Expression())
			if zca.OkExpression == nil {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "zero coalescing assignment: missing ok variable",
					Primary: quickanno.Expected(p, p.Pos(), "an ok variable after the comma"),
					Secondary: []diagnostic.Annotation{
						anno.Position(p.File, *zca.VarComma, "because of the comma here"),
					},
				})
			}
			parser.TrySkip(p, comment.OrHorizontalWhitespace())
		}

		zca.Colon = parser.TryOptionalRuneAt(p, ':', nil)

		zca.EqualSign = parser.TryRuneAt(p, '=')
		if zca.EqualSign == nil {
			return nil
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())

		zca.Expression = parser.Try(p, ZeroCoalescing())
		if zca.Expression == nil {
			return nil
		}

		return &zca
	}
}

func ZeroCoalescingAssignmentAsCode(zca *ast.ZeroCoalescingAssignment) ast.Code {
	var n int
	n += len(zca.ValueExpression.Nodes)
	if zca.VarComma != nil {
		n++
	}
	if zca.OkExpression != nil {
		n += len(zca.OkExpression.Nodes)
	}
	if zca.Colon != nil || zca.EqualSign != nil {
		n++
	}
	if zca.Expression != nil {
		n++
	}

	c := make(ast.Code, n)
	i := copy(c, zca.ValueExpression.Nodes)
	if zca.VarComma != nil {
		c[i] = &ast.GoCode{Code: ",", Position: zca.VarComma}
		i++
	}
	if zca.OkExpression != nil {
		i += copy(c[i:], zca.OkExpression.Nodes)
	}
	if zca.EqualSign != nil {
		if zca.Colon != nil {
			c[i] = &ast.GoCode{Code: ":=", Position: zca.Colon}
		} else {
			c[i] = &ast.GoCode{Code: "=", Position: zca.EqualSign}
		}
		i++
	} else if zca.Colon != nil {
		c[i] = &ast.GoCode{Code: ":", Position: zca.Colon}
		i++
	}
	if zca.Expression != nil {
		c[i] = zca.Expression
	}
	return c
}

func IncDec() parser.Func[*ast.IncDec] {
	return func(p *parser.Parser) *ast.IncDec {
		expr := parser.Try(p, Expression())
		if expr == nil {
			return nil
		}
		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		return parser.Try(p, incDec(expr))
	}
}

func incDec(expr *ast.Expression) parser.Func[*ast.IncDec] {
	return func(p *parser.Parser) *ast.IncDec {
		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		switch {
		case parser.TryOptionalToken(p, "++", nil):
			pos := p.Pos()
			pos.Col -= ast.Col(len("++"))
			return &ast.IncDec{Expression: expr, IncrPos: &pos}
		case parser.TryOptionalToken(p, "--", nil):
			pos := p.Pos()
			pos.Col -= ast.Col(len("--"))
			return &ast.IncDec{Expression: expr, DecrPos: &pos}
		default:
			return nil
		}
	}
}

func IncDecAsCode(incDec *ast.IncDec) ast.Code {
	if incDec.IncrPos == nil && incDec.DecrPos == nil {
		return incDec.Expression.Nodes
	}

	c := make(ast.Code, len(incDec.Expression.Nodes)+1)
	copy(c, incDec.Expression.Nodes)
	if incDec.IncrPos != nil {
		c[len(c)-1] = &ast.GoCode{Code: "++", Position: incDec.IncrPos}
	} else {
		c[len(c)-1] = &ast.GoCode{Code: "--", Position: incDec.DecrPos}
	}
	return c
}

func ConstDeclaration() parser.Func[*ast.ConstDeclaration] {
	return func(p *parser.Parser) *ast.ConstDeclaration {
		constKeyword := parser.TryKeywordAt(p, "const")
		if constKeyword == nil {
			return nil
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())

		var d ast.ConstDeclaration
		d.Const = constKeyword

		d.LParen = parser.TryOptionalRuneAt(p, '(', comment.OrAnyWhitespace())
		if d.LParen == nil {
			spec := parser.Try(p, ConstSpec())
			if spec != nil {
				d.Specs = []*ast.ConstSpec{spec}
			} else {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "const declaration: missing identifiers",
					Primary: quickanno.Expected(p, p.Pos(), "one or more identifiers"),
					Examples: []diagnostic.Example{
						{Example: "`const bark = \"woof\"`"},
					},
				})
			}
			return &d
		}

		for {
			parser.TrySkip(p, comment.OrAnyWhitespace())

			spec := parser.Try(p, ConstSpec())
			if spec == nil {
				break
			}
			d.Specs = append(d.Specs, spec)

			parser.TrySkip(p, comment.OrHorizontalWhitespace())
			if parser.MatchesAnyRune(p, ')') {
				break
			}

			parser.Try(p, comment.AndMustEOS())
		}
		if len(d.Specs) == 0 {
			d.Specs = nil
		} else {
			d.Specs = slices.Clip(d.Specs)
		}

		err := unexpected.UntilAnyRune(p, comment.OrAnyWhitespace(), ')')
		if err != nil {
			err.Message = "const declaration: unexpected runes"
			p.CaptureError(err)
		}

		d.RParen = parser.TryRuneAt(p, ')')
		if d.RParen == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "const declaration: missing ')'",
				Primary: quickanno.Expected(p, *d.LParen, "a closing ')' for the '(' here"),
			})
		}

		return &d
	}
}

func ConstSpec() parser.Func[*ast.ConstSpec] {
	return func(p *parser.Parser) *ast.ConstSpec {
		names := parser.Try(p, list.CommaList("const", "name", "names", golang.Identifier()))
		if names == nil {
			return nil
		}

		var s ast.ConstSpec
		s.Names = names

		if parser.TrySkip(p, comment.OrHorizontalWhitespace()) {
			s.Type = parser.TryOptional(p, golang.Type(), comment.OrHorizontalWhitespace())
		}

		err := unexpected.UntilAnyRune(p, comment.OrHorizontalWhitespace(), '=', ';', ')')
		if err != nil {
			err.Message = "const spec: unexpected runes"
			p.CaptureError(err)
		}

		s.EqualSign = parser.TryRuneAt(p, '=')
		if s.EqualSign == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "const spec: missing equal sign",
				Primary: quickanno.Expected(p, p.Pos(), "an equal sign"),
			})
			return &s
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())

		valuesStart := p.Pos()
		s.Values = parser.Try(p, list.CommaList("const", "value", "values", Expression()))
		valuesEnd := p.Pos()
		if s.Values == nil {
			if len(s.Names) == 1 {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "const spec: missing values",
					Primary: quickanno.Expected(p, p.Pos(), "an expression"),
				})
			} else {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "const spec: missing values",
					Primary: quickanno.Expected(p, p.Pos(), fmt.Sprint("one or a list of ", len(s.Names), " expressions")),
				})
			}
		}

		if len(s.Values) > 1 && len(s.Names) != len(s.Values) {
			if len(s.Names) == 1 {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "var spec: mismatched number of values and constants",
					Primary: []diagnostic.Annotation{
						anno.Range(p.File, valuesStart, valuesEnd,
							fmt.Sprint("expected a single expression, but found ", len(s.Values))),
					},
				})
			} else {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "var spec: mismatched number of values and constants",
					Primary: []diagnostic.Annotation{
						anno.Range(p.File, valuesStart, valuesEnd,
							fmt.Sprint("a single or ", len(s.Names), " expressions, but found ", len(s.Values))),
					},
				})
			}
		}

		return &s
	}
}

func ConstDeclarationAsCode(start ast.Position, d *ast.ConstDeclaration) ast.Code {
	var n int
	n++ // 'const'
	if d.LParen != nil {
		n++
	}
	for _, spec := range d.Specs {
		if spec == nil {
			continue
		}
		n += len(spec.Names) // commas included
		if spec.EqualSign != nil {
			n++
		}
		for _, val := range spec.Values {
			if val != nil {
				n += len(val.Nodes)
			}
		}
		n += max(0, len(spec.Values)-1) // commas
	}
	if d.RParen != nil {
		n++
	}

	lastEnd := start

	c := make(ast.Code, n)
	c[0] = &ast.GoCode{Code: "const", Position: d.Const}
	i := 1
	if d.LParen != nil {
		c[i] = &ast.GoCode{Code: "(", Position: d.LParen}
		i++
	}
	for _, spec := range d.Specs {
		if spec == nil {
			continue
		}
		for nameI, name := range spec.Names {
			if name == nil {
				pos := lastEnd
				if nameI < len(spec.Names)-1 { // not last
					c[i] = &ast.GoCode{Code: ",", Position: &pos}
				} else {
					c[i] = &ast.GoCode{Code: "", Position: &pos}
				}
				lastEnd = c[i].End()
				i++
				continue
			}

			if nameI < len(spec.Names)-1 { // not last
				c[i] = &ast.GoCode{Code: name.Name + ",", Position: name.Position}
			} else {
				c[i] = &ast.GoCode{Code: name.Name, Position: name.Position}
			}
			i++
		}
		if spec.EqualSign != nil {
			c[i] = &ast.GoCode{Code: "=", Position: spec.EqualSign}
			i++
		}
		lastEnd = c[i-1].End()
		for valueI, value := range spec.Values {
			if value == nil {
				if valueI < len(spec.Values)-1 { // not last
					pos := lastEnd
					c[i] = &ast.GoCode{Code: ",", Position: &pos}
					lastEnd = c[i].End()
					i++
				}
				continue
			}

			i += copy(c[i:], value.Nodes)
			if valueI < len(spec.Values)-1 { // not last
				comma := value.End()
				comma.Col++
				c[i] = &ast.GoCode{Code: ",", Position: &comma}
				i++
			}
		}
	}
	if d.RParen != nil {
		c[i] = &ast.GoCode{Code: ")", Position: d.RParen}
	}

	return c
}

func VarDeclaration() parser.Func[*ast.VarDeclaration] {
	return func(p *parser.Parser) *ast.VarDeclaration {
		varKeyword := parser.TryKeywordAt(p, "var")
		if varKeyword == nil {
			return nil
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())

		var d ast.VarDeclaration
		d.Var = varKeyword

		d.LParen = parser.TryOptionalRuneAt(p, '(', comment.OrAnyWhitespace())
		if d.LParen == nil {
			pos := p.Pos()
			spec := parser.Try(p, VarSpec())
			if spec == nil {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "var declaration: missing identifiers",
					Primary: quickanno.Expected(p, pos, "one or more identifiers"),
					Examples: []diagnostic.Example{
						{Example: "`var bark = \"woof\"`"},
					},
				})
			}
			d.Specs = []*ast.VarSpec{spec}
			return &d
		}

		for {
			parser.TrySkip(p, comment.OrAnyWhitespace())

			spec := parser.Try(p, VarSpec())
			if spec == nil {
				break
			}
			d.Specs = append(d.Specs, spec)

			parser.TrySkip(p, comment.OrHorizontalWhitespace())
			if parser.MatchesAnyRune(p, ')') {
				break
			}

			parser.Try(p, comment.AndMustEOS())
		}
		if len(d.Specs) == 0 {
			d.Specs = nil
		} else {
			d.Specs = slices.Clip(d.Specs)
		}

		err := unexpected.UntilAnyRune(p, comment.OrAnyWhitespace(), ')')
		if err != nil {
			err.Message = "var declaration: unexpected runes"
			p.CaptureError(err)
		}

		d.RParen = parser.TryRuneAt(p, ')')
		if d.RParen == nil {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "var declaration: missing ')'",
				Primary: quickanno.Expected(p, *d.LParen, "a closing ')' for the '(' here"),
			})
		}

		return &d
	}
}

func VarSpec() parser.Func[*ast.VarSpec] {
	return func(p *parser.Parser) *ast.VarSpec {
		names := parser.Try(p, list.CommaList("var spec", "name", "names", golang.Identifier()))
		if names == nil {
			return nil
		}

		var s ast.VarSpec
		s.Names = names

		pos := p.Pos()
		if parser.TrySkip(p, comment.OrHorizontalWhitespace()) {
			s.Type = parser.TryOptional(p, golang.Type(), comment.OrHorizontalWhitespace())
		}

		err := unexpected.UntilAnyRune(p, comment.OrHorizontalWhitespace(), '=', ';', ')')
		if err != nil {
			err.Message = "var spec: unexpected runes"
			p.CaptureError(err)
		}

		s.EqualSign = parser.TryRuneAt(p, '=')
		if s.EqualSign == nil {
			if s.Type == nil {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "var spec: missing type or value",
					Primary: quickanno.Expected(p, pos, "either a type or an equal sign"),
				})
			}
			return &s
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())

		valuesStart := p.Pos()
		s.Values = parser.Try(p, list.CommaList("var spec", "value", "values", Expression()))
		valuesEnd := p.Pos()
		if s.Values == nil {
			if len(s.Names) == 1 {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "var spec: missing values",
					Primary: quickanno.Expected(p, p.Pos(), "an expression"),
				})
			} else {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "var spec: missing values",
					Primary: quickanno.Expected(p, p.Pos(), fmt.Sprint("one or a list of ", len(s.Names), " expressions")),
				})
			}
		}

		if len(s.Values) > 1 && len(s.Names) != len(s.Values) {
			if len(s.Names) == 1 {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "var spec: mismatched number of values and variables",
					Primary: []diagnostic.Annotation{
						anno.Range(p.File, valuesStart, valuesEnd,
							fmt.Sprint("expected a single expression, but found ", len(s.Values))),
					},
				})
			} else {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "var spec: mismatched number of values and variables",
					Primary: []diagnostic.Annotation{
						anno.Range(p.File, valuesStart, valuesEnd,
							fmt.Sprint("a single or ", len(s.Names), " expressions, but found ", len(s.Values))),
					},
				})
			}
		}

		return &s
	}
}

func VarDeclarationAsCode(start ast.Position, d *ast.VarDeclaration) ast.Code {
	var n int
	n++ // 'var'
	if d.LParen != nil {
		n++
	}
	for _, spec := range d.Specs {
		if spec == nil {
			continue
		}
		n += len(spec.Names) // commas included
		if spec.EqualSign != nil {
			n++
		}
		for _, val := range spec.Values {
			if val != nil {
				n += len(val.Nodes)
			}
		}
		n += max(0, len(spec.Values)-1) // commas
	}
	if d.RParen != nil {
		n++
	}

	lastEnd := start

	c := make(ast.Code, n)
	c[0] = &ast.GoCode{Code: "var", Position: d.Var}
	i := 1
	if d.LParen != nil {
		c[i] = &ast.GoCode{Code: "(", Position: d.LParen}
		i++
	}
	for _, spec := range d.Specs {
		if spec == nil {
			continue
		}
		for nameI, name := range spec.Names {
			if name == nil {
				pos := lastEnd
				if nameI < len(spec.Names)-1 { // not last
					c[i] = &ast.GoCode{Code: ",", Position: &pos}
				} else {
					c[i] = &ast.GoCode{Code: "", Position: &pos}
				}
				lastEnd = c[i].End()
				i++
				continue
			}

			if nameI < len(spec.Names)-1 { // not last
				c[i] = &ast.GoCode{Code: name.Name + ",", Position: name.Position}
			} else {
				c[i] = &ast.GoCode{Code: name.Name, Position: name.Position}
			}
			i++
		}
		if spec.EqualSign != nil {
			c[i] = &ast.GoCode{Code: "=", Position: spec.EqualSign}
			i++
		}
		lastEnd = c[i-1].End()
		for valueI, value := range spec.Values {
			if value == nil {
				if valueI < len(spec.Values)-1 { // not last
					pos := lastEnd
					c[i] = &ast.GoCode{Code: ",", Position: &pos}
					lastEnd = c[i].End()
					i++
				}
				continue
			}

			i += copy(c[i:], value.Nodes)
			if valueI < len(spec.Values)-1 { // not last
				comma := value.End()
				comma.Col++
				c[i] = &ast.GoCode{Code: ",", Position: &comma}
				i++
			}
		}
	}
	if d.RParen != nil {
		c[i] = &ast.GoCode{Code: ")", Position: d.RParen}
	}

	return c
}

func ShortVarDeclaration() parser.Func[*ast.ShortVarDeclaration] {
	return func(p *parser.Parser) *ast.ShortVarDeclaration {
		names := parser.Try(p, list.CommaList("short var declaration", "identifier", "identifiers", golang.Identifier()))
		if names == nil {
			return nil
		}

		var d ast.ShortVarDeclaration
		d.Names = names

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		d.ColonEqualSign = parser.TryTokenAt(p, ":=")
		if d.ColonEqualSign == nil {
			return nil
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())

		valuesStart := p.Pos()
		d.Values = parser.Try(p, list.CommaList("short var declaration", "value", "values", Expression()))
		valuesEnd := p.Pos()
		if d.Values == nil {
			if len(d.Names) == 1 {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "short var declaration: missing values",
					Primary: quickanno.Expected(p, p.Pos(), "an expression"),
				})
			} else {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "short var declaration: missing values",
					Primary: quickanno.Expected(p, p.Pos(), fmt.Sprint("one or a list of ", len(d.Names), " expressions")),
				})
			}
		}

		if len(d.Values) > 1 && len(d.Names) != len(d.Values) {
			if len(d.Names) == 1 {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "short var declaration: mismatched number of values and variables",
					Primary: []diagnostic.Annotation{
						anno.Range(p.File, valuesStart, valuesEnd,
							fmt.Sprint("expected a single expression, but found ", len(d.Values))),
					},
				})
			} else {
				p.CaptureError(&diagnostic.Diagnostic{
					Message: "short var declaration: mismatched number of values and variables",
					Primary: []diagnostic.Annotation{
						anno.Range(p.File, valuesStart, valuesEnd,
							fmt.Sprint("a single or ", len(d.Names), " expressions, but found ", len(d.Values))),
					},
				})
			}
		}

		return &d
	}
}

func ShortVarDeclarationAsCode(start ast.Position, d *ast.ShortVarDeclaration) ast.Code {
	var n int
	n += len(d.Names) // commas included
	if d.ColonEqualSign != nil {
		n++
	}
	for _, val := range d.Values {
		if val != nil {
			n += len(val.Nodes)
		}
	}
	n += max(0, len(d.Values)-1) // commas

	lastEnd := start

	c := make(ast.Code, n)
	i := 0
	for nameI, name := range d.Names {
		if name == nil {
			pos := lastEnd
			lastEnd.Col++
			if nameI < len(d.Names)-1 { // not last
				c[i] = &ast.GoCode{Code: ",", Position: &pos}
			} else {
				c[i] = &ast.GoCode{Code: "", Position: &pos}
			}
			lastEnd = c[i].End()
			i++
			continue
		}

		if nameI < len(d.Names)-1 { // not last
			c[i] = &ast.GoCode{Code: name.Name + ",", Position: name.Position}
		} else {
			c[i] = &ast.GoCode{Code: name.Name, Position: name.Position}
		}
		i++
	}
	if d.ColonEqualSign != nil {
		c[i] = &ast.GoCode{Code: ":=", Position: d.ColonEqualSign}
		i++
	}

	lastEnd = c[i-1].End()
	for valueI, value := range d.Values {
		if value == nil {
			if valueI < len(d.Values)-1 { // not last
				pos := lastEnd
				c[i] = &ast.GoCode{Code: ",", Position: &pos}
				lastEnd = c[i].End()
				i++
			}
			continue
		}

		i += copy(c[i:], value.Nodes)
		if valueI < len(d.Values)-1 { // not last
			comma := value.End()
			c[i] = &ast.GoCode{Code: ",", Position: &comma}
			lastEnd = c[i].End()
			i++
		}
	}

	return c
}

func Label() parser.Func[*ast.Label] {
	return func(p *parser.Parser) *ast.Label {
		name := parser.Try(p, golang.Identifier())
		if name == nil {
			return nil
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		colon := parser.TryRuneAt(p, ':')
		if colon == nil || parser.MatchesAnyRune(p, '=') {
			return nil
		}

		return &ast.Label{Name: name, Colon: colon}
	}
}

func LabelAsCode(l *ast.Label) ast.Code {
	if l.Colon == nil {
		return ast.Code{&ast.GoCode{Code: l.Name.Name, Position: l.Name.Position}}
	}
	return ast.Code{
		&ast.GoCode{Code: l.Name.Name, Position: l.Name.Position},
		&ast.GoCode{Code: ":", Position: l.Colon},
	}
}

func Assignment() parser.Func[*ast.Assignment] {
	return assignment(nil)
}

func assignment(e *ast.Expression) parser.Func[*ast.Assignment] {
	return func(p *parser.Parser) *ast.Assignment {
		var lhs []*ast.Expression
		if e != nil {
			parser.TrySkip(p, comment.OrHorizontalWhitespace())
			if parser.TryOptionalRune(p, ',', nil) {
				parser.TrySkip(p, comment.OrAnyWhitespace())

				es := parser.Try(p, list.CommaList("assignment", "assignee", "assignee", Expression()))
				if es == nil {
					p.CaptureError(&diagnostic.Diagnostic{
						Message: "assignment: missing assignees",
						Primary: quickanno.Expected(p, p.Pos(), "one or more assignees being assigned to"),
					})
				}
				lhs = append([]*ast.Expression{e}, es...)
			} else {
				lhs = []*ast.Expression{e}
			}
		} else {
			lhs = parser.Try(p, list.CommaList("assignment", "assignee", "assignee", Expression()))
			if lhs == nil {
				return nil
			}
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		op := parser.Try(p, golang.AssignOp())
		if op == "" {
			return nil
		}

		var a ast.Assignment
		a.LHS = lhs
		a.Operator = op
		a.OperatorPosition = p.PosPtr()
		a.OperatorPosition.Col -= ast.Col(len(op))

		parser.TrySkip(p, comment.OrAnyWhitespace())
		rhsStart := p.Pos()
		a.RHS = parser.Try(p, list.CommaList("assignment", "value", "values", Expression()))
		rhsEnd := p.Pos()
		if len(a.RHS) > 0 {
			if len(a.RHS) > 1 && len(a.LHS) != len(a.RHS) {
				if len(a.LHS) == 1 {
					p.CaptureError(&diagnostic.Diagnostic{
						Message: "assigment: mismatched number of values and assignees",
						Primary: []diagnostic.Annotation{
							anno.Range(p.File, rhsStart, rhsEnd,
								fmt.Sprint("expected a single expression, but found ", len(a.RHS))),
						},
					})
				} else {
					p.CaptureError(&diagnostic.Diagnostic{
						Message: "assigment: mismatched number of values and assignees",
						Primary: []diagnostic.Annotation{
							anno.Range(p.File, rhsStart, rhsEnd,
								fmt.Sprint("a single or ", len(a.LHS), " expressions, but found ", len(a.RHS))),
						},
					})
				}
			}
		} else {
			p.CaptureError(&diagnostic.Diagnostic{
				Message: "assignment: missing values",
				Primary: quickanno.Expected(p, rhsStart, fmt.Sprint("one or a list of ", len(a.LHS), " expressions being assigned to")),
			})
		}

		return &a
	}
}

func AssignmentAsCode(start ast.Position, a *ast.Assignment) ast.Code {
	var n int
	for _, e := range a.LHS {
		if e != nil {
			n += len(e.Nodes)
		}
	}
	n += max(0, len(a.LHS)-1) // commas
	if a.OperatorPosition != nil {
		n++ // operator
	}
	for _, e := range a.RHS {
		if e != nil {
			n += len(e.Nodes)
		}
	}
	n += max(0, len(a.RHS)-1) // commas

	lastEnd := start

	c := make(ast.Code, n)
	var i int
	for eI, e := range a.LHS {
		if e == nil {
			if eI < len(a.LHS)-1 { // not last
				pos := lastEnd
				lastEnd.Col++
				c[i] = &ast.GoCode{Code: ",", Position: &pos}
				i++
			}
			continue
		}

		i += copy(c[i:], e.Nodes)
		if eI < len(a.LHS)-1 { // not last
			pos := e.End()
			c[i] = &ast.GoCode{Code: ",", Position: &pos}
			lastEnd = c[i].End()
			i++
		}
	}
	if a.OperatorPosition != nil {
		c[i] = &ast.GoCode{Code: a.Operator, Position: a.OperatorPosition}
		i++
	}

	lastEnd = c[i-1].End()
	for eI, e := range a.RHS {
		if e == nil {
			if eI < len(a.RHS)-1 { // not last
				pos := lastEnd
				c[i] = &ast.GoCode{Code: ",", Position: &pos}
				lastEnd = c[i].End()
				i++
			}
			continue
		}
		i += copy(c[i:], e.Nodes)
		if eI < len(a.RHS)-1 { // not last
			pos := e.End()
			c[i] = &ast.GoCode{Code: ",", Position: &pos}
			lastEnd = c[i].End()
			i++
		}
	}
	return c
}
