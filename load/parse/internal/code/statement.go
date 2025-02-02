package code

import (
	"fmt"
	"slices"

	"github.com/mavolin/corgi/v2/fancyerr"
	"github.com/mavolin/corgi/v2/fancyerr/anno"
	"github.com/mavolin/corgi/v2/file/ast"
	parser "github.com/mavolin/corgi/v2/load/parse/internal"
	"github.com/mavolin/corgi/v2/load/parse/internal/comment"
	"github.com/mavolin/corgi/v2/load/parse/internal/golang"
	"github.com/mavolin/corgi/v2/load/parse/internal/list"
	"github.com/mavolin/corgi/v2/load/parse/internal/quickanno"
	"github.com/mavolin/corgi/v2/load/parse/internal/unexpected"
)

func Statement() parser.Func[*ast.Statement] {
	return func(p *parser.Parser) (*ast.Statement, *fancyerr.Error) {
		s, _ := parser.Try(p, parsedStatement())
		if s != nil && s.Parsed != nil {
			return s, nil
		}

		c, _ := parser.TryOk(p, NonZCCode(true))
		parser.MustSkip(p, comment.AndMustEOS())
		if s == nil {
			if len(c) == 0 {
				return nil, &fancyerr.Error{
					Message: "missing simple statement",
					Primary: quickanno.Expected(p, p.Pos(), "a simple statement"),
				}
			}
			return &ast.Statement{Code: c}, nil
		}

		if len(c) == 0 {
			return s, nil
		}
		c2 := make(ast.Code, len(s.Code)+len(c))
		copy(c2, s.Code)
		copy(c2[len(s.Code):], c)
		return &ast.Statement{Code: c2}, nil
	}
}

func ParsedStatement() parser.Func[*ast.Statement] {
	return func(p *parser.Parser) (*ast.Statement, *fancyerr.Error) {
		s, err := parser.Try(p, parsedStatement())
		if err != nil {
			return nil, err
		} else if s.Parsed == nil {
			return nil, &fancyerr.Error{
				Message: "missing parsed statement",
				Primary: quickanno.Expected(p, p.Pos(), "a parsed statement"),
			}
		}
		return s, nil
	}
}

func parsedStatement() parser.Func[*ast.Statement] {
	return func(p *parser.Parser) (*ast.Statement, *fancyerr.Error) {
		if r, ok := parser.TryOk(p, Return()); ok {
			return &ast.Statement{
				Code:   ReturnAsCode(r),
				Parsed: r,
			}, nil
		} else if b, ok := parser.TryOk(p, Break()); ok {
			return &ast.Statement{
				Code:   BreakAsCode(b),
				Parsed: b,
			}, nil
		} else if c, ok := parser.TryOk(p, Continue()); ok {
			return &ast.Statement{
				Code:   ContinueAsCode(c),
				Parsed: c,
			}, nil
		} else if f, ok := parser.TryOk(p, Fallthrough()); ok {
			return &ast.Statement{
				Code:   FallthroughAsCode(f),
				Parsed: f,
			}, nil
		} else if d, ok := parser.TryOk(p, Defer()); ok {
			return &ast.Statement{
				Code:   DeferAsCode(d),
				Parsed: d,
			}, nil
		} else if cd, ok := parser.TryOk(p, ConstDeclaration()); ok {
			return &ast.Statement{
				Code:   ConstDeclarationAsCode(cd),
				Parsed: cd,
			}, nil
		} else if vd, ok := parser.TryOk(p, VarDeclaration()); ok {
			return &ast.Statement{
				Code:   VarDeclarationAsCode(vd),
				Parsed: vd,
			}, nil
		} else if l, ok := parser.TryOk(p, Label()); ok {
			return &ast.Statement{
				Code:   LabelAsCode(l),
				Parsed: l,
			}, nil
		}

		beforeExpr := p.CloneState()
		e, ok := parser.TryOk(p, Expression())
		if !ok {
			return nil, &fancyerr.Error{
				Message: "missing statement",
				Primary: quickanno.Expected(p, p.Pos(), "a statement"),
			}
		}
		afterExpr := p.CloneState()

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		if zca, ok := parser.TryOptionalOk(p, zeroCoalescingAssignment(e)); ok {
			return &ast.Statement{
				Code:   ZeroCoalescingAssignmentAsCode(zca),
				Parsed: zca,
			}, nil
		} else if incDec, ok := parser.TryOptionalOk(p, incDec(e)); ok {
			return &ast.Statement{
				Code:   IncDecAsCode(incDec),
				Parsed: incDec,
			}, nil
		} else if a, ok := parser.TryOk(p, assignment(e)); ok {
			return &ast.Statement{
				Code:   AssignmentAsCode(a),
				Parsed: a,
			}, nil
		}

		p.RestoreState(beforeExpr)
		if svd, ok := parser.TryOk(p, ShortVarDeclaration()); ok {
			return &ast.Statement{
				Code:   ShortVarDeclarationAsCode(svd),
				Parsed: svd,
			}, nil
		}

		p.RestoreState(afterExpr)
		return &ast.Statement{Code: e.Code}, nil
	}
}

func SimpleStatement() parser.Func[*ast.SimpleStatement] {
	return func(p *parser.Parser) (*ast.SimpleStatement, *fancyerr.Error) {
		ss, _ := parser.Try(p, parsedSimpleStatement())
		if ss != nil && ss.Parsed != nil {
			return ss, nil
		}

		c, _ := parser.TryOk(p, NonZCCode(true))
		parser.MustSkip(p, comment.AndMustEOS())
		if ss == nil {
			if len(c) == 0 {
				return nil, &fancyerr.Error{
					Message: "missing simple statement",
					Primary: quickanno.Expected(p, p.Pos(), "a simple statement"),
				}
			}
			return &ast.SimpleStatement{Code: c}, nil
		}

		if len(c) == 0 {
			return ss, nil
		}
		c2 := make(ast.Code, len(ss.Code)+len(c))
		copy(c2, ss.Code)
		copy(c2[len(ss.Code):], c)
		return &ast.SimpleStatement{Code: c2}, nil
	}
}

func ParsedSimpleStatement() parser.Func[*ast.SimpleStatement] {
	return func(p *parser.Parser) (*ast.SimpleStatement, *fancyerr.Error) {
		ss, err := parser.Try(p, parsedSimpleStatement())
		if err != nil {
			return nil, err
		} else if ss.Parsed == nil {
			return nil, &fancyerr.Error{
				Message: "missing parsed simple statement",
				Primary: quickanno.Expected(p, p.Pos(), "a parsed simple statement"),
			}
		}
		return ss, nil
	}
}

func parsedSimpleStatement() parser.Func[*ast.SimpleStatement] {
	return func(p *parser.Parser) (*ast.SimpleStatement, *fancyerr.Error) {
		beforeExpr := p.CloneState()

		e, ok := parser.TryOk(p, Expression())
		if !ok {
			return nil, &fancyerr.Error{
				Message: "missing simple statement",
				Primary: quickanno.Expected(p, p.Pos(), "a simple statement"),
			}
		}
		afterExpr := p.CloneState()

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		if zca, ok := parser.TryOptionalOk(p, zeroCoalescingAssignment(e)); ok {
			return &ast.SimpleStatement{
				Code:   ZeroCoalescingAssignmentAsCode(zca),
				Parsed: zca,
			}, nil
		} else if incDec, ok := parser.TryOptionalOk(p, incDec(e)); ok {
			return &ast.SimpleStatement{
				Code:   IncDecAsCode(incDec),
				Parsed: incDec,
			}, nil
		} else if a, ok := parser.TryOk(p, assignment(e)); ok {
			return &ast.SimpleStatement{
				Code:   AssignmentAsCode(a),
				Parsed: a,
			}, nil
		}

		p.RestoreState(beforeExpr)
		if svd, ok := parser.TryOk(p, ShortVarDeclaration()); ok {
			return &ast.SimpleStatement{
				Code:   ShortVarDeclarationAsCode(svd),
				Parsed: svd,
			}, nil
		}

		p.RestoreState(afterExpr)
		return &ast.SimpleStatement{Code: e.Code}, nil
	}
}

func Return() parser.Func[*ast.Return] {
	return func(p *parser.Parser) (*ast.Return, *fancyerr.Error) {
		r := &ast.Return{Position: p.Pos()}
		if !parser.TryToken(p, "return") {
			return nil, &fancyerr.Error{
				Message: "missing return statement",
				Primary: quickanno.Expected(p, p.Pos(), "a return statement"),
			}
		}

		if parser.TrySkipOk(p, comment.AndEOS()) {
			return r, nil
		}

		if !parser.TrySkipOk(p, comment.OrAnyWhitespace()) {
			return nil, &fancyerr.Error{
				Message: "missing return statement",
				Primary: quickanno.Expected(p, r.Position, "a return statement"),
				Hints: []fancyerr.Hint{
					{Hint: "You might be missing a space, newline, or semicolon."},
				},
			}
		}

		var ok bool
		r.Error, ok = parser.TryOptionalOk(p, Expression())
		if !ok {
			p.CaptureError(&fancyerr.Error{
				Message: "return statement: missing expression or end of statement",
				Primary: quickanno.Expected(p, p.Pos(), "an expression or the end of the statement"),
			})
		}

		parser.MustSkip(p, comment.AndMustEOS())
		return r, nil
	}
}

func ReturnAsCode(r *ast.Return) ast.Code {
	if r.Error == nil {
		return ast.Code{&ast.GoCode{Code: "return", Position: r.Position}}
	}

	c := make(ast.Code, 1+len(r.Error.Code))
	c[0] = &ast.GoCode{Code: "return", Position: r.Position}
	copy(c[1:], r.Error.Code)
	return c
}

func Break() parser.Func[*ast.Break] {
	return func(p *parser.Parser) (*ast.Break, *fancyerr.Error) {
		b := &ast.Break{Position: p.Pos()}
		if !parser.TryToken(p, "break") {
			return nil, &fancyerr.Error{
				Message: "missing break statement",
				Primary: quickanno.Expected(p, p.Pos(), "a break statement"),
			}
		}

		if parser.TrySkipOk(p, comment.AndEOS()) {
			return b, nil
		}

		if !parser.TrySkipOk(p, comment.OrAnyWhitespace()) {
			return nil, &fancyerr.Error{
				Message: "missing break statement",
				Primary: quickanno.Expected(p, b.Position, "a break statement"),
				Hints: []fancyerr.Hint{
					{Hint: "You might be missing a space, newline, or semicolon."},
				},
			}
		}

		var ok bool
		b.Label, ok = parser.TryOptionalOk(p, golang.Identifier())
		if !ok {
			p.CaptureError(&fancyerr.Error{
				Message: "return statement: missing label or end of statement",
				Primary: quickanno.Expected(p, p.Pos(), "a label or the end of the statement"),
			})
		}

		parser.MustSkip(p, comment.AndMustEOS())
		return b, nil
	}
}

func BreakAsCode(b *ast.Break) ast.Code {
	if b.Label == nil {
		return ast.Code{&ast.GoCode{Code: "break", Position: b.Position}}
	}
	return ast.Code{
		&ast.GoCode{Code: "break", Position: b.Position},
		&ast.GoCode{Code: b.Label.Ident, Position: b.Label.Position},
	}
}

func Continue() parser.Func[*ast.Continue] {
	return func(p *parser.Parser) (*ast.Continue, *fancyerr.Error) {
		b := &ast.Continue{Position: p.Pos()}
		if !parser.TryToken(p, "continue") {
			return nil, &fancyerr.Error{
				Message: "missing continue statement",
				Primary: quickanno.Expected(p, p.Pos(), "a continue statement"),
			}
		}

		if parser.TrySkipOk(p, comment.AndEOS()) {
			return b, nil
		}

		if !parser.TrySkipOk(p, comment.OrAnyWhitespace()) {
			return nil, &fancyerr.Error{
				Message: "missing continue statement",
				Primary: quickanno.Expected(p, b.Position, "a continue statement"),
				Hints: []fancyerr.Hint{
					{Hint: "You might be missing a space, newline, or semicolon."},
				},
			}
		}

		var ok bool
		b.Label, ok = parser.TryOptionalOk(p, golang.Identifier())
		if !ok {
			p.CaptureError(&fancyerr.Error{
				Message: "return statement: missing label or end of statement",
				Primary: quickanno.Expected(p, p.Pos(), "a label or the end of the statement"),
			})
		}

		parser.MustSkip(p, comment.AndMustEOS())
		return b, nil
	}
}

func ContinueAsCode(c *ast.Continue) ast.Code {
	if c.Label == nil {
		return ast.Code{&ast.GoCode{Code: "continue", Position: c.Position}}
	}
	return ast.Code{
		&ast.GoCode{Code: "continue", Position: c.Position},
		&ast.GoCode{Code: c.Label.Ident, Position: c.Label.Position},
	}
}

func Fallthrough() parser.Func[*ast.Fallthrough] {
	return func(p *parser.Parser) (*ast.Fallthrough, *fancyerr.Error) {
		b := &ast.Fallthrough{Position: p.Pos()}
		if !parser.TryToken(p, "fallthrough") {
			return nil, &fancyerr.Error{
				Message: "missing fallthrough statement",
				Primary: quickanno.Expected(p, p.Pos(), "a fallthrough statement"),
			}
		}

		if parser.TrySkipOk(p, comment.AndEOS()) {
			return b, nil
		}

		if !parser.TrySkipOk(p, comment.OrAnyWhitespace()) {
			return nil, &fancyerr.Error{
				Message: "missing fallthrough statement",
				Primary: quickanno.Expected(p, b.Position, "a fallthrough statement"),
				Hints: []fancyerr.Hint{
					{Hint: "You might be missing a space, newline, or semicolon."},
				},
			}
		}

		var ok bool
		b.Label, ok = parser.TryOptionalOk(p, golang.Identifier())
		if !ok {
			p.CaptureError(&fancyerr.Error{
				Message: "return statement: missing label or end of statement",
				Primary: quickanno.Expected(p, p.Pos(), "a label or the end of the statement"),
			})
		}

		parser.MustSkip(p, comment.AndMustEOS())
		return b, nil
	}
}

func FallthroughAsCode(f *ast.Fallthrough) ast.Code {
	if f.Label == nil {
		return ast.Code{&ast.GoCode{Code: "fallthrough", Position: f.Position}}
	}
	return ast.Code{
		&ast.GoCode{Code: "fallthrough", Position: f.Position},
		&ast.GoCode{Code: f.Label.Ident, Position: f.Label.Position},
	}
}

func Defer() parser.Func[*ast.Defer] {
	return func(p *parser.Parser) (*ast.Defer, *fancyerr.Error) {
		d := &ast.Defer{Position: p.Pos()}
		if !parser.TryToken(p, "defer") {
			return nil, &fancyerr.Error{
				Message: "missing defer statement",
				Primary: quickanno.Expected(p, p.Pos(), "a defer statement"),
			}
		}

		if !parser.TrySkipOk(p, comment.OrAnyWhitespace()) {
			return nil, &fancyerr.Error{
				Message: "missing defer statement",
				Primary: quickanno.Expected(p, d.Position, "a defer statement"),
				Hints: []fancyerr.Hint{
					{Hint: "You might be missing a space, newline, or semicolon."},
				},
			}
		}

		d.Expression = parser.Must(p, Expression())
		parser.MustSkip(p, comment.AndMustEOS())
		return d, nil
	}
}

func DeferAsCode(d *ast.Defer) ast.Code {
	if d.Expression == nil {
		return ast.Code{&ast.GoCode{Code: "defer", Position: d.Position}}
	}

	c := make(ast.Code, 1+len(d.Expression.Code))
	c[0] = &ast.GoCode{Code: "defer", Position: d.Position}
	copy(c[1:], d.Expression.Code)
	return c
}

func ZeroCoalescingAssignment() parser.Func[*ast.ZeroCoalescingAssignment] {
	return func(p *parser.Parser) (*ast.ZeroCoalescingAssignment, *fancyerr.Error) {
		valueExpr, ok := parser.TryOk(p, Expression())
		if !ok {
			return nil, &fancyerr.Error{
				Message: "missing zero coalescing assignment",
				Primary: quickanno.Expected(p, p.Pos(), "a variable name or expression"),
			}
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		return parser.Try(p, zeroCoalescingAssignment(valueExpr))
	}
}

func zeroCoalescingAssignment(valueExpr *ast.Expression) parser.Func[*ast.ZeroCoalescingAssignment] {
	return func(p *parser.Parser) (*ast.ZeroCoalescingAssignment, *fancyerr.Error) {
		zca := &ast.ZeroCoalescingAssignment{ValueExpression: *valueExpr}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		zca.VarComma = p.PosPtr()
		if parser.TryOptionalRune(p, ',') {
			parser.TrySkip(p, comment.OrAnyWhitespace())
			zca.OkExpression = parser.Must(p, Expression())
			parser.TrySkip(p, comment.OrHorizontalWhitespace())
		} else {
			zca.VarComma = nil
		}

		zca.Colon = p.PosPtr()
		if !parser.TryOptionalRune(p, ':') {
			zca.Colon = nil
		}

		zca.EqualSign = p.PosPtr()
		if !parser.TryRune(p, '=') {
			return nil, &fancyerr.Error{
				Message: "missing zero coalescing assignment",
				Primary: quickanno.Expected(p, *zca.EqualSign, "an equal sign"),
			}
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())

		var ok bool
		zca.Expression, ok = parser.TryOk(p, ZeroCoalescing())
		if !ok {
			return nil, &fancyerr.Error{
				Message: "missing zero coalescing assignment",
				Primary: quickanno.Expected(p, p.Pos(), "a zero coalescing expression"),
			}
		}

		parser.MustSkip(p, comment.AndMustEOS())
		return zca, nil
	}

}

func ZeroCoalescingAssignmentAsCode(zca *ast.ZeroCoalescingAssignment) ast.Code {
	n := len(zca.ValueExpression.Code)
	if zca.VarComma != nil {
		n++
	}
	if zca.OkExpression != nil {
		n += len(zca.OkExpression.Code)
	}
	if zca.Colon != nil || zca.EqualSign != nil {
		n++
	}
	if zca.Expression != nil {
		n++
	}

	c := make(ast.Code, n)
	i := copy(c, zca.ValueExpression.Code)
	if zca.VarComma != nil {
		c[i] = &ast.GoCode{Code: ",", Position: *zca.VarComma}
		i++
	}
	if zca.OkExpression != nil {
		i += copy(c[i:], zca.OkExpression.Code)
	}
	if zca.EqualSign != nil {
		if zca.Colon != nil {
			c[i] = &ast.GoCode{Code: ":=", Position: *zca.Colon}
		} else {
			c[i] = &ast.GoCode{Code: "=", Position: *zca.EqualSign}
		}
		i++
	} else if zca.Colon != nil {
		c[i] = &ast.GoCode{Code: ":", Position: *zca.Colon}
		i++
	}
	if zca.Expression != nil {
		c[i] = zca.Expression
	}
	return c
}

func IncDec() parser.Func[*ast.IncDec] {
	return func(p *parser.Parser) (*ast.IncDec, *fancyerr.Error) {
		expr, ok := parser.TryOk(p, Expression())
		if !ok {
			return nil, &fancyerr.Error{
				Message: "missing inc/dec",
				Primary: quickanno.Expected(p, p.Pos(), "an expression followed by `++` or `--`"),
			}
		}
		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		return parser.Try(p, incDec(expr))
	}
}

func incDec(expr *ast.Expression) parser.Func[*ast.IncDec] {
	return func(p *parser.Parser) (*ast.IncDec, *fancyerr.Error) {
		incDec := &ast.IncDec{Expression: *expr}

		pos := p.Pos()
		if parser.TryOptionalToken(p, "++") {
			incDec.IncrPos = &pos
		} else if parser.TryToken(p, "--") {
			incDec.DecrPos = &pos
		} else {
			return nil, &fancyerr.Error{
				Message: "missing inc/dec",
				Primary: quickanno.Expected(p, pos, "`++` or `--`"),
			}
		}

		parser.MustSkip(p, comment.AndMustEOS())
		return incDec, nil
	}
}

func IncDecAsCode(incDec *ast.IncDec) ast.Code {
	if incDec.IncrPos == nil && incDec.DecrPos == nil {
		return incDec.Expression.Code
	}

	c := make(ast.Code, len(incDec.Expression.Code)+1)
	copy(c, incDec.Expression.Code)
	if incDec.IncrPos != nil {
		c[len(c)-1] = &ast.GoCode{Code: "++", Position: *incDec.IncrPos}
	} else {
		c[len(c)-1] = &ast.GoCode{Code: "--", Position: *incDec.DecrPos}
	}
	return c
}

func ConstDeclaration() parser.Func[*ast.ConstDeclaration] {
	return func(p *parser.Parser) (*ast.ConstDeclaration, *fancyerr.Error) {
		d := &ast.ConstDeclaration{Const: p.Pos()}
		if !parser.TryToken(p, "const") {
			return nil, &fancyerr.Error{
				Message: "missing const declaration",
				Primary: quickanno.Expected(p, d.Pos(), "a const declaration"),
			}
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())

		d.LParen = p.PosPtr()
		if !parser.TryOptionalRune(p, '(') {
			d.LParen = nil

			d.Specs = make([]*ast.ConstSpec, 1)
			d.Specs[0] = parser.Must(p, ConstSpec())

			parser.MustSkip(p, comment.AndMustEOS())
			return d, nil
		}

		d.Specs = make([]*ast.ConstSpec, 0, 18)
		for {
			parser.TrySkip(p, comment.OrAnyWhitespace())

			spec, ok := parser.TryOk(p, ConstSpec())
			if !ok {
				break
			}
			d.Specs = append(d.Specs, spec)

			parser.TrySkip(p, comment.OrHorizontalWhitespace())
			if parser.MatchesAnyRune(p, ')') {
				break
			}

			parser.MustSkip(p, comment.AndEOS())
		}
		d.Specs = slices.Clip(d.Specs)

		err := unexpected.UntilAnyRune(p, comment.OrAnyWhitespace(), ')')
		if err != nil {
			err.Message = "const declaration: unexpected runes"
			p.CaptureError(err)
		}

		d.RParen = p.PosPtr()
		if !parser.TryRune(p, ')') {
			d.RParen = nil
			p.CaptureError(&fancyerr.Error{
				Message: "const declaration: missing ')'",
				Primary: quickanno.Expected(p, *d.LParen, "a closing ')' for the '(' here"),
			})
		}

		parser.MustSkip(p, comment.AndMustEOS())
		return d, nil
	}
}

func ConstSpec() parser.Func[*ast.ConstSpec] {
	return func(p *parser.Parser) (*ast.ConstSpec, *fancyerr.Error) {
		names, ok := parser.TryOk(p, list.CommaList("const name", "const names", golang.Identifier()))
		if !ok {
			return nil, &fancyerr.Error{
				Message: "missing const spec",
				Primary: quickanno.Expected(p, p.Pos(), "one or more identifiers"),
				Examples: []fancyerr.Example{
					{Example: "`bark = \"woof\"`"},
				},
			}
		}

		s := &ast.ConstSpec{Names: names}

		if parser.TrySkipOk(p, comment.OrHorizontalWhitespace()) {
			s.Type, ok = parser.TryOptionalOk(p, golang.Type())
			if ok {
				parser.TrySkip(p, comment.OrHorizontalWhitespace())
			}
		}

		err := unexpected.UntilAnyRune(p, comment.OrHorizontalWhitespace(), '=', ';', ')')
		if err != nil {
			err.Message = "const spec: unexpected runes"
			p.CaptureError(err)
		}

		s.EqualSign = p.PosPtr()
		if !parser.TryRune(p, '=') {
			p.CaptureError(&fancyerr.Error{
				Message: "const spec: missing equal sign",
				Primary: quickanno.Expected(p, p.Pos(), "an equal sign"),
			})
			return s, nil
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())

		s.Values, ok = parser.TryOk(p, list.CommaList("const value", "const values", Expression()))
		if !ok {
			if len(names) == 1 {
				p.CaptureError(&fancyerr.Error{
					Message: "const spec: missing values",
					Primary: quickanno.Expected(p, p.Pos(), "an expression"),
				})
			} else {
				p.CaptureError(&fancyerr.Error{
					Message: "const spec: missing values",
					Primary: quickanno.Expected(p, p.Pos(), fmt.Sprint("one or a list of ", len(names), " expressions")),
				})
			}
		}

		if len(s.Values) > 1 && len(s.Names) != len(s.Values) {
			if len(names) == 1 {
				p.CaptureError(&fancyerr.Error{
					Message: "var spec: mismatched number of values and constants",
					Primary: []fancyerr.Annotation{
						anno.Range(p.File, s.Values[0].Pos(), s.Values[len(s.Values)-1].End(),
							fmt.Sprint("expected a single expression, but found ", len(s.Values))),
					},
				})
			} else {
				p.CaptureError(&fancyerr.Error{
					Message: "var spec: mismatched number of values and constants",
					Primary: []fancyerr.Annotation{
						anno.Range(p.File, s.Values[0].Pos(), s.Values[len(s.Values)-1].End(),
							fmt.Sprint("a single or ", len(names), " expressions, but found ", len(s.Values))),
					},
				})
			}
		}

		return s, nil
	}
}

func ConstDeclarationAsCode(d *ast.ConstDeclaration) ast.Code {
	n := 1
	if d.LParen != nil {
		n++
	}
	for _, spec := range d.Specs {
		n += len(spec.Names)
		if spec.EqualSign != nil {
			n++
		}
		for _, val := range spec.Values {
			n += len(val.Code)
		}
		n += max(0, len(spec.Values)-1)
	}
	if d.RParen != nil {
		n++
	}

	c := make(ast.Code, n)
	c[0] = &ast.GoCode{Code: "const", Position: d.Const}
	i := 1
	if d.LParen != nil {
		c[i] = &ast.GoCode{Code: "(", Position: *d.LParen}
		i++
	}
	for _, spec := range d.Specs {
		for nameI, name := range spec.Names {
			if nameI < len(spec.Names)-1 {
				c[i] = &ast.GoCode{Code: name.Ident + ",", Position: name.Position}
			} else {
				c[i] = &ast.GoCode{Code: name.Ident, Position: name.Position}
			}
			i++
		}
		if spec.EqualSign != nil {
			c[i] = &ast.GoCode{Code: "=", Position: *spec.EqualSign}
		}
		for valueI, value := range spec.Values {
			i += copy(c[i:], value.Code)
			if valueI < len(spec.Values)-1 {
				comma := value.End()
				comma.Col++
				c[i] = &ast.GoCode{Code: ",", Position: comma}
				i++
			}
		}
	}
	if d.RParen != nil {
		c[i] = &ast.GoCode{Code: ")", Position: *d.RParen}
	}

	return c
}

func VarDeclaration() parser.Func[*ast.VarDeclaration] {
	return func(p *parser.Parser) (*ast.VarDeclaration, *fancyerr.Error) {
		d := &ast.VarDeclaration{Var: p.Pos()}
		if !parser.TryToken(p, "var") {
			return nil, &fancyerr.Error{
				Message: "missing var declaration",
				Primary: quickanno.Expected(p, d.Pos(), "a var declaration"),
			}
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())

		d.LParen = p.PosPtr()
		if !parser.TryOptionalRune(p, '(') {
			d.LParen = nil

			d.Specs = make([]*ast.VarSpec, 1)
			d.Specs[0] = parser.Must(p, VarSpec())

			parser.MustSkip(p, comment.AndMustEOS())
			return d, nil
		}

		d.Specs = make([]*ast.VarSpec, 0, 18)
		for {
			parser.TrySkip(p, comment.OrAnyWhitespace())

			spec, ok := parser.TryOk(p, VarSpec())
			if !ok {
				break
			}
			d.Specs = append(d.Specs, spec)

			parser.TrySkip(p, comment.OrHorizontalWhitespace())
			if parser.MatchesAnyRune(p, ')') {
				break
			}

			parser.MustSkip(p, comment.AndEOS())
		}
		d.Specs = slices.Clip(d.Specs)

		err := unexpected.UntilAnyRune(p, comment.OrAnyWhitespace(), ')')
		if err != nil {
			err.Message = "var declaration: unexpected runes"
			p.CaptureError(err)
		}

		d.RParen = p.PosPtr()
		if !parser.TryRune(p, ')') {
			d.RParen = nil
			p.CaptureError(&fancyerr.Error{
				Message: "var declaration: missing ')'",
				Primary: quickanno.Expected(p, *d.LParen, "a closing ')' for the '(' here"),
			})
		}

		parser.MustSkip(p, comment.AndMustEOS())
		return d, nil
	}
}

func VarSpec() parser.Func[*ast.VarSpec] {
	return func(p *parser.Parser) (*ast.VarSpec, *fancyerr.Error) {
		names, ok := parser.TryOk(p, list.CommaList("var name", "var names", golang.Identifier()))
		if !ok {
			return nil, &fancyerr.Error{
				Message: "missing var spec",
				Primary: quickanno.Expected(p, p.Pos(), "one or more identifiers"),
				Examples: []fancyerr.Example{
					{Example: "`bark = \"woof\"`"},
				},
			}
		}

		s := &ast.VarSpec{Names: names}

		var pos ast.Position
		if parser.TrySkipOk(p, comment.OrHorizontalWhitespace()) {
			pos = p.Pos()

			s.Type, ok = parser.TryOptionalOk(p, golang.Type())
			if ok {
				parser.TrySkip(p, comment.OrHorizontalWhitespace())
			}
		}

		err := unexpected.UntilAnyRune(p, comment.OrHorizontalWhitespace(), '=', ';', ')')
		if err != nil {
			err.Message = "var spec: unexpected runes"
			p.CaptureError(err)
		}

		s.EqualSign = p.PosPtr()
		if !parser.TryRune(p, '=') {
			s.EqualSign = nil
			if s.Type == nil {
				p.CaptureError(&fancyerr.Error{
					Message: "var spec: missing type or value",
					Primary: quickanno.Expected(p, pos, "either a type or an equal sign"),
				})
			}
			return s, nil
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())

		s.Values, ok = parser.TryOk(p, list.CommaList("var value", "var values", Expression()))
		if !ok {
			if len(names) == 1 {
				p.CaptureError(&fancyerr.Error{
					Message: "var spec: missing values",
					Primary: quickanno.Expected(p, p.Pos(), "an expression"),
				})
			} else {
				p.CaptureError(&fancyerr.Error{
					Message: "var spec: missing values",
					Primary: quickanno.Expected(p, p.Pos(), fmt.Sprint("one or a list of ", len(names), " expressions")),
				})
			}
		}

		if len(s.Values) > 1 && len(s.Names) != len(s.Values) {
			if len(names) == 1 {
				p.CaptureError(&fancyerr.Error{
					Message: "var spec: mismatched number of values and variables",
					Primary: []fancyerr.Annotation{
						anno.Range(p.File, s.Values[0].Pos(), s.Values[len(s.Values)-1].End(),
							fmt.Sprint("expected a single expression, but found ", len(s.Values))),
					},
				})
			} else {
				p.CaptureError(&fancyerr.Error{
					Message: "var spec: mismatched number of values and variables",
					Primary: []fancyerr.Annotation{
						anno.Range(p.File, s.Values[0].Pos(), s.Values[len(s.Values)-1].End(),
							fmt.Sprint("a single or ", len(names), " expressions, but found ", len(s.Values))),
					},
				})
			}
		}

		return s, nil
	}
}

func VarDeclarationAsCode(d *ast.VarDeclaration) ast.Code {
	n := 1
	if d.LParen != nil {
		n++
	}
	for _, spec := range d.Specs {
		n += len(spec.Names)
		if spec.EqualSign != nil {
			n++
		}
		for _, val := range spec.Values {
			n += len(val.Code)
		}
		n += max(0, len(spec.Values)-1)
	}
	if d.RParen != nil {
		n++
	}

	c := make(ast.Code, n)
	c[0] = &ast.GoCode{Code: "var", Position: d.Var}
	i := 1
	if d.LParen != nil {
		c[i] = &ast.GoCode{Code: "(", Position: *d.LParen}
		i++
	}
	for _, spec := range d.Specs {
		for nameI, name := range spec.Names {
			if nameI < len(spec.Names)-1 {
				c[i] = &ast.GoCode{Code: name.Ident + ",", Position: name.Position}
			} else {
				c[i] = &ast.GoCode{Code: name.Ident, Position: name.Position}
			}
			i++
		}
		if spec.EqualSign != nil {
			c[i] = &ast.GoCode{Code: "=", Position: *spec.EqualSign}
		}
		for valueI, value := range spec.Values {
			i += copy(c[i:], value.Code)
			if valueI < len(spec.Values)-1 {
				comma := value.End()
				comma.Col++
				c[i] = &ast.GoCode{Code: ",", Position: comma}
				i++
			}
		}
	}
	if d.RParen != nil {
		c[i] = &ast.GoCode{Code: ")", Position: *d.RParen}
	}

	return c
}

func ShortVarDeclaration() parser.Func[*ast.ShortVarDeclaration] {
	return func(p *parser.Parser) (*ast.ShortVarDeclaration, *fancyerr.Error) {
		names, ok := parser.TryOk(p, list.CommaList("identifier", "identifiers", golang.Identifier()))
		if !ok {
			return nil, &fancyerr.Error{
				Message: "missing short var declaration",
				Primary: quickanno.Expected(p, p.Pos(), "one or more identifiers"),
			}
		}

		d := &ast.ShortVarDeclaration{Names: names}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		d.ColonEqual = p.PosPtr()
		if !parser.TryToken(p, ":=") {
			return nil, &fancyerr.Error{
				Message: "missing short var declaration",
				Primary: quickanno.Expected(p, p.Pos(), "a `:=`"),
			}
		}

		parser.TrySkip(p, comment.OrAnyWhitespace())

		d.Values, ok = parser.TryOk(p, list.CommaList("value", "values", Expression()))
		if !ok {
			if len(names) == 1 {
				p.CaptureError(&fancyerr.Error{
					Message: "short var declaration: missing values",
					Primary: quickanno.Expected(p, p.Pos(), "an expression"),
				})
			} else {
				p.CaptureError(&fancyerr.Error{
					Message: "short var declaration: missing values",
					Primary: quickanno.Expected(p, p.Pos(), fmt.Sprint("one or a list of ", len(names), " expressions")),
				})
			}
		}

		if len(d.Values) > 1 && len(d.Names) != len(d.Values) {
			if len(names) == 1 {
				p.CaptureError(&fancyerr.Error{
					Message: "short var declaration: mismatched number of values and variables",
					Primary: []fancyerr.Annotation{
						anno.Range(p.File, d.Values[0].Pos(), d.Values[len(d.Values)-1].End(),
							fmt.Sprint("expected a single expression, but found ", len(d.Values))),
					},
				})
			} else {
				p.CaptureError(&fancyerr.Error{
					Message: "short var declaration: mismatched number of values and variables",
					Primary: []fancyerr.Annotation{
						anno.Range(p.File, d.Values[0].Pos(), d.Values[len(d.Values)-1].End(),
							fmt.Sprint("a single or ", len(names), " expressions, but found ", len(d.Values))),
					},
				})
			}
		}

		parser.MustSkip(p, comment.AndMustEOS())
		return d, nil
	}
}

func ShortVarDeclarationAsCode(d *ast.ShortVarDeclaration) ast.Code {
	n := len(d.Names)
	if d.ColonEqual != nil {
		n++
	}
	for _, val := range d.Values {
		n += len(val.Code)
	}
	n += max(0, len(d.Values)-1)

	c := make(ast.Code, n)
	i := 0
	for nameI, name := range d.Names {
		if nameI < len(d.Names)-1 {
			c[i] = &ast.GoCode{Code: name.Ident + ",", Position: name.Position}
		} else {
			c[i] = &ast.GoCode{Code: name.Ident, Position: name.Position}
		}
		i++
	}
	if d.ColonEqual != nil {
		c[i] = &ast.GoCode{Code: ":=", Position: *d.ColonEqual}
		i++
	}
	for valueI, value := range d.Values {
		i += copy(c[i:], value.Code)
		if valueI < len(d.Values)-1 {
			comma := value.End()
			comma.Col++
			c[i] = &ast.GoCode{Code: ",", Position: comma}
			i++
		}
	}

	return c
}

func Label() parser.Func[*ast.Label] {
	return func(p *parser.Parser) (*ast.Label, *fancyerr.Error) {
		name, ok := parser.TryOk(p, golang.Identifier())
		if !ok {
			return nil, &fancyerr.Error{
				Message: "missing label",
				Primary: quickanno.Expected(p, p.Pos(), "an identifier followed by a colon"),
			}
		}

		l := &ast.Label{Name: *name}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		l.Colon = p.PosPtr()
		if !parser.TryRune(p, ':') || parser.MatchesAnyRune(p, '=') {
			return nil, &fancyerr.Error{
				Message: "missing label",
				Primary: quickanno.Expected(p, p.Pos(), "a colon and the end of the statement"),
			}
		}

		parser.MustSkip(p, comment.AndMustEOS())
		return l, nil
	}
}

func LabelAsCode(l *ast.Label) ast.Code {
	if l.Colon == nil {
		return ast.Code{&ast.GoCode{Code: l.Name.Ident, Position: l.Name.Position}}
	}
	return ast.Code{
		&ast.GoCode{Code: l.Name.Ident, Position: l.Name.Position},
		&ast.GoCode{Code: ":", Position: *l.Colon},
	}
}

func Assignment() parser.Func[*ast.Assignment] {
	return assignment(nil)
}

func assignment(e *ast.Expression) parser.Func[*ast.Assignment] {
	return func(p *parser.Parser) (*ast.Assignment, *fancyerr.Error) {
		a := &ast.Assignment{}
		if e != nil {
			parser.TrySkip(p, comment.OrHorizontalWhitespace())
			if parser.TryOptionalRune(p, ',') {
				parser.TrySkip(p, comment.OrAnyWhitespace())

				es := parser.Must(p, list.CommaList("expression", "expressions", Expression()))
				a.LHS = append([]*ast.Expression{e}, es...)
			} else {
				a.LHS = []*ast.Expression{e}
			}
		} else {
			var ok bool
			a.LHS, ok = parser.TryOk(p, list.CommaList("expression", "expressions", Expression()))
			if !ok {
				return nil, &fancyerr.Error{
					Message: "missing assignment",
					Primary: quickanno.Expected(p, p.Pos(), "one or more expressions being assigned to"),
				}
			}
		}

		parser.TrySkip(p, comment.OrHorizontalWhitespace())

		a.OperatorPosition = p.PosPtr()
		var ok bool
		a.SpecialOperator, ok = parser.TryOk(p, golang.AssignOp())
		if !ok {
			return nil, &fancyerr.Error{
				Message: "missing assignment",
				Primary: quickanno.Expected(p, p.Pos(), "an assignment operator, e.g. `=` or `+=`"),
			}
		}
		a.SpecialOperator = a.SpecialOperator[:len(a.SpecialOperator)-1]

		parser.TrySkip(p, comment.OrAnyWhitespace())

		a.RHS = parser.Must(p, list.CommaList("expression", "expressions", Expression()))
		if len(a.RHS) > 0 {
			if len(a.RHS) > 1 && len(a.LHS) != len(a.RHS) {
				if len(a.LHS) == 1 {
					p.CaptureError(&fancyerr.Error{
						Message: "assigment: mismatched number of values and assignees",
						Primary: []fancyerr.Annotation{
							anno.Range(p.File, a.RHS[0].Pos(), a.RHS[len(a.RHS)-1].End(),
								fmt.Sprint("expected a single expression, but found ", len(a.RHS))),
						},
					})
				} else {
					p.CaptureError(&fancyerr.Error{
						Message: "assigment: mismatched number of values and assignees",
						Primary: []fancyerr.Annotation{
							anno.Range(p.File, a.RHS[0].Pos(), a.RHS[len(a.RHS)-1].End(),
								fmt.Sprint("a single or ", len(a.LHS), " expressions, but found ", len(a.RHS))),
						},
					})
				}
			}
		}

		parser.MustSkip(p, comment.AndMustEOS())
		return a, nil
	}
}

func AssignmentAsCode(a *ast.Assignment) ast.Code {
	n := 2*len(a.LHS) - 1
	if a.OperatorPosition != nil {
		n++
	}
	n += 2*len(a.RHS) - 1

	c := make(ast.Code, n)
	var i int
	for eI, e := range a.LHS {
		i += copy(c[i:], e.Code)
		if eI < len(a.LHS)-1 {
			p := e.End()
			p.Col++
			c[i] = &ast.GoCode{Code: ",", Position: p}
			i++
		}
	}
	if a.OperatorPosition != nil {
		c[i] = &ast.GoCode{Code: a.SpecialOperator + "=", Position: *a.OperatorPosition}
		i++
	}
	for eI, e := range a.RHS {
		i += copy(c[i:], e.Code)
		if eI < len(a.RHS)-1 {
			p := e.End()
			p.Col++
			c[i] = &ast.GoCode{Code: ",", Position: p}
			i++
		}
	}
	return c
}
