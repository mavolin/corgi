package write

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/typeinfer"
)

// ============================================================================
// Generated Expression
// ======================================================================================

// generateExpression writes the passed expression using the passed escapers.
//
// If txtEsc is set, this expression is assumed to required contextual escaping.
//
// If that is the case, string text is escaped with txtEsc, all else is escaped
// using exprEsc.
//
// If expr is contains a single ExpressionItem of type [file.ChainExpression],
// generateExpression generates code that performs the checks of the chain
// expression, calling writer with a function that generated the resolved
// expression at the point in code where all checks pass.
//
// If the chain expression has a default, writer is instead called immediately,
// then when genExpr is called, either the chain expression result or the
// default is written.
//
// If expr contains a single [file.TernaryExpression], generateExpression similarly
// writes an if else, calling write twice, once for the ifTrue and once for the
// ifFalse expression
//
// If expr is any other expression, writer is called immediately with a
// function generating the expression.
//
// If writer is nil, generateExpression calls genExpr immediately, instead of
// passing it to writer.
//
// writer must call genExpr only once.
func generateExpression(ctx *ctx, expr file.Expression, txtEsc *escaper, exprEsc *escaper, writer func(genExpr func())) {
	txtEscName := "nil"
	if txtEsc != nil {
		txtEscName = txtEsc.name
	}
	exprEscName := "nil"
	if exprEsc != nil {
		exprEscName = exprEsc.name
	}
	ctx.debugItem(expr, fmt.Sprintf("(txt escaper: %s, expr esc: %s) (see below)", txtEscName, exprEscName))

	if writer == nil {
		writer = func(genExpr func()) {
			genExpr()
		}
	}

	if len(expr.Expressions) == 1 {
		switch exprItm := expr.Expressions[0].(type) {
		case file.ChainExpression:
			generateChainExpression(ctx, exprItm, nil, exprEsc, writer)
			return
		case file.TernaryExpression:
			writer(func() {
				generateTernaryExpression(ctx, exprItm, txtEsc, exprEsc)
			})
			return
		case file.GoExpression:
			writer(func() {
				generateGoExpression(ctx, exprItm, exprEsc)
			})
			return
		case file.StringExpression:
			writer(func() {
				generateStringExpression(ctx, exprItm, txtEsc, exprEsc)
			})
			return
		}
	}

	if txtEsc != nil {
		panic(fmt.Errorf("generateExpression (%d:%d): attempting to generated contextual escaped complex expression"+
			" (this should've been caught during validation)", expr.Pos().Line, expr.Pos().Col))
	}

	writer(func() {
		ctx.generate(inlineExpression(ctx, expr), exprEsc)
	})
}

const (
	chainValVar   = "chainVal"
	chainIndexVar = "chainIndex"
)

func generateChainExpression(ctx *ctx, cexpr file.ChainExpression, defaultExpr *file.Expression, esc *escaper, writer func(func())) {
	if cexpr.Default != nil {
		defaultExpr = cexpr.Default
	}

	var valBuilder strings.Builder
	valBuilder.WriteString(inlineExpression(ctx, file.Expression{Expressions: []file.ExpressionItem{cexpr.Root}}))

	if defaultExpr != nil {
		writer(func() {
			ctx.flushGenerate()
			if cexpr.CheckRoot {
				setChainValVar(ctx, &valBuilder)
				ctx.write("if ")
				ctx.writeln("; !" + ctx.woofFunc("IsZero", ctx.ident(chainValVar)) + " {")
			}

			checksPassedGoto := ctx.nextGotoIdent()

			generateChainExprItems(ctx, cexpr.Chain, &valBuilder, func() {
				ctx.generate(valBuilder.String(), esc)
				ctx.flushGenerate()
				ctx.writeln("goto " + checksPassedGoto)
			})
			if cexpr.CheckRoot {
				ctx.writeln("}")
			}

			ctx.generate(inlineExpression(ctx, *defaultExpr), esc)
			ctx.flushGenerate()
			ctx.writeln(checksPassedGoto + ":")
		})
		return
	}

	ctx.flushGenerate()
	if cexpr.CheckRoot {
		setChainValVar(ctx, &valBuilder)
		ctx.write("if ")
		ctx.writeln("; !" + ctx.woofFunc("IsZero", ctx.ident(chainValVar)) + " {")
		defer ctx.writeln("}")
	}

	generateChainExprItems(ctx, cexpr.Chain, &valBuilder, func() {
		writer(func() {
			ctx.generate(valBuilder.String(), esc)
		})
		ctx.flushGenerate()
	})
}

func generateChainExprItems(ctx *ctx, cexprs []file.ChainExpressionItem, valBuilder *strings.Builder, checksPassed func()) {
	for _, cexpr := range cexprs {
		switch cexpr := cexpr.(type) {
		case file.IndexExpression:
			if cexpr.CheckIndex {
				setChainValVar(ctx, valBuilder) // set now, in case chainIndexVar is in the valBuilder
				ctx.writeln("")
				ctx.writeln(ctx.ident(chainIndexVar) + " := " + inlineExpression(ctx, cexpr.Index))

				switch typeinfer.Infer(cexpr.Index) {
				case "int", "": // either a map, or a slice
					ctx.writeln("if " + ctx.woofFunc("CanIndex", ctx.ident(chainValVar), ctx.ident(chainIndexVar)) + " {")

					valBuilder.WriteByte('[')
					valBuilder.WriteString(ctx.ident(chainIndexVar))
					valBuilder.WriteByte(']')

					//goland:noinspection GoDeferInLoop
					defer ctx.writeln("}")
				default: // this is a map
					ctx.writeln("if " + ctx.ident(chainValVar) + ", " + ctx.ident("ok") + " := " + ctx.ident(chainValVar) + "[" + ctx.ident(chainIndexVar) + "]; " + ctx.ident("ok") + " {")
					//goland:noinspection GoDeferInLoop
					defer ctx.writeln("}")
				}
			} else {
				valBuilder.WriteByte('[')
				valBuilder.WriteString(inlineExpression(ctx, cexpr.Index))
				valBuilder.WriteByte(']')
			}

			if cexpr.CheckValue {
				ctx.write("if ")
				setChainValVar(ctx, valBuilder)
				ctx.writeln("; !" + ctx.woofFunc("IsZero", ctx.ident(chainValVar)) + " {")
				//goland:noinspection GoDeferInLoop
				defer ctx.writeln("}")
			}
		case file.DotIdentExpression:
			valBuilder.WriteByte('.')
			valBuilder.WriteString(cexpr.Ident.Ident)

			if cexpr.Check {
				ctx.write("if ")
				setChainValVar(ctx, valBuilder)
				ctx.writeln("; !" + ctx.woofFunc("IsZero", ctx.ident(chainValVar)) + " {")
				//goland:noinspection GoDeferInLoop
				defer ctx.writeln("}")
			}
		case file.ParenExpression:
			valBuilder.WriteByte('(')
			for i, arg := range cexpr.Args {
				if i > 0 {
					valBuilder.WriteString(", ")
				}

				valBuilder.WriteString(inlineExpression(ctx, arg))
			}
			valBuilder.WriteByte(')')

			if cexpr.Check {
				ctx.write("if ")
				setChainValVar(ctx, valBuilder)
				ctx.writeln("; !" + ctx.woofFunc("IsZero", ctx.ident(chainValVar)) + " {")
				//goland:noinspection GoDeferInLoop
				defer ctx.writeln("}")
			}
		case file.TypeAssertionExpression:
			var typBuilder strings.Builder
			typBuilder.WriteString(strings.Repeat("*", cexpr.PointerCount))
			if cexpr.Package != nil {
				typBuilder.WriteString(cexpr.Package.Ident)
				typBuilder.WriteByte('.')
			}
			typBuilder.WriteString(cexpr.Type.Ident)
			typ := typBuilder.String()

			if cexpr.Check {
				setChainValVar(ctx, valBuilder)
				ctx.writeln("if " + ctx.ident(chainValVar) + ", " + ctx.ident("ok") + " := " + ctx.ident(chainValVar) + ".(" + typ + "); " + ctx.ident("ok") + " {")
				//goland:noinspection GoDeferInLoop
				defer ctx.writeln("}")
			} else {
				valBuilder.WriteString(".(")
				valBuilder.WriteString(typ)
				valBuilder.WriteByte(')')
			}
		}
	}

	checksPassed()
}

func setChainValVar(ctx *ctx, sb *strings.Builder) {
	ctx.write(ctx.ident(chainValVar) + " := " + sb.String())
	sb.Reset()
	sb.WriteString(ctx.ident(chainValVar))
}

func generateTernaryExpression(ctx *ctx, texpr file.TernaryExpression, txtEsc *escaper, exprEsc *escaper) {
	ctx.debugItem(texpr, "(generated) (see below)")
	ctx.flushGenerate()
	ctx.write("if ")
	ctx.write(inlineCondition(ctx, texpr.Condition))
	ctx.writeln(" {")
	generateExpression(ctx, texpr.IfTrue, txtEsc, exprEsc, nil)
	ctx.flushGenerate()
	ctx.writeln("} else {")
	generateExpression(ctx, texpr.IfFalse, txtEsc, exprEsc, nil)
	ctx.flushGenerate()
	ctx.writeln("}")
}

func generateGoExpression(ctx *ctx, gexpr file.GoExpression, esc *escaper) {
	ctx.debugItem(gexpr, gexpr.Expression)
	ctx.generateExpr(gexpr.Expression, esc)
}

func generateStringExpression(ctx *ctx, sexpr file.StringExpression, txtEsc *escaper, exprEsc *escaper) {
	ctx.debugItem(sexpr, "(generated) (see below)")

	for _, exprItm := range sexpr.Contents {
		switch exprItm := exprItm.(type) {
		case file.StringExpressionText:
			s := unquoteStringExpressionText(sexpr, exprItm)

			ctx.debugItem(exprItm, s)

			if txtEsc != nil {
				ctx.generate(s, txtEsc)
			} else {
				ctx.generate(s, exprEsc)
			}
		case file.StringExpressionInterpolation:
			if exprItm.FormatDirective == "" {
				ctx.debugItem(exprItm, "(see sub expressions)")
				generateExpression(ctx, exprItm.Expression, nil, exprEsc, nil)
				continue
			}

			ctx.debugItem(exprItm, "[%"+exprItm.FormatDirective+"] (see sub expressions)")
			fmtString := strconv.Quote("%" + exprItm.FormatDirective)
			ctx.generateExpr(ctx.fmtFunc("Sprintf", fmtString, inlineExpression(ctx, exprItm.Expression)), exprEsc)
		}
	}
}

// ============================================================================
// Value Chain Expression
// ======================================================================================

func valueChainExpression(ctx *ctx, cexpr file.ChainExpression, onValue func(expr string)) {
	var valBuilder strings.Builder
	valBuilder.WriteString(inlineExpression(ctx, file.Expression{Expressions: []file.ExpressionItem{cexpr.Root}}))

	ctx.flushGenerate()
	if cexpr.CheckRoot {
		setChainValVar(ctx, &valBuilder)
		ctx.write("if ")
		ctx.writeln("; !" + ctx.woofFunc("IsZero", ctx.ident(chainValVar)) + " {")
	}

	var checksPassedGoto string
	if cexpr.Default != nil {
		checksPassedGoto = ctx.nextGotoIdent()
	}

	generateChainExprItems(ctx, cexpr.Chain, &valBuilder, func() {
		onValue(valBuilder.String())
		ctx.flushGenerate()
		if cexpr.Default != nil {
			ctx.writeln("goto " + checksPassedGoto)
		}
	})
	if cexpr.CheckRoot {
		ctx.writeln("}")
	}

	if cexpr.Default != nil {
		onValue(inlineExpression(ctx, *cexpr.Default))
		ctx.flushGenerate()
		ctx.writeln(checksPassedGoto + ":")
	}

	ctx.flushGenerate()
}

// ============================================================================
// Inline Expression
// ======================================================================================

// yields an unescaped expression
func inlineExpression(ctx *ctx, expr file.Expression) string {
	var sb strings.Builder

	for _, exprItm := range expr.Expressions {
		switch exprItm := exprItm.(type) {
		case file.TernaryExpression:
			inlineTernaryExpression(ctx, &sb, exprItm)
		case file.GoExpression:
			inlineGoExpression(ctx, &sb, exprItm)
		case file.StringExpression:
			inlineStringExpression(ctx, &sb, exprItm)
		}
	}

	return sb.String()
}

func inlineChainExpression(ctx *ctx, sb *strings.Builder, cexpr file.ChainExpression, defaultExpr *file.Expression, typ string) {
	if cexpr.Default != nil {
		defaultExpr = cexpr.Default
	}

	sb.WriteString("func () ")
	sb.WriteString(typ)
	sb.WriteString(" {\n")
	defer sb.WriteString("}()")

	var valBuilder strings.Builder
	valBuilder.WriteString(inlineExpression(ctx, file.Expression{Expressions: []file.ExpressionItem{cexpr.Root}}))

	if cexpr.CheckRoot {
		setChainValVar(ctx, &valBuilder)
		sb.WriteString("if ")
		sb.WriteString("; !" + ctx.woofFunc("IsZero", ctx.ident(chainValVar)) + " {\n")
	}

	inlineChainExprItems(ctx, sb, cexpr.Chain, &valBuilder, func() {
		sb.WriteString("return ")
		sb.WriteString(valBuilder.String())
		sb.WriteByte('\n')
	})
	if cexpr.CheckRoot {
		sb.WriteString("}\n")
	}

	sb.WriteString("return ")
	sb.WriteString(inlineExpression(ctx, *defaultExpr))
	sb.WriteByte('\n')
}

func inlineChainExprItems(ctx *ctx, exprBuilder *strings.Builder, cexprs []file.ChainExpressionItem, valBuilder *strings.Builder, checksPassed func()) {
	for _, cexpr := range cexprs {
		switch cexpr := cexpr.(type) {
		case file.IndexExpression:
			if cexpr.CheckIndex {
				inlineSetChainValVar(ctx, exprBuilder, valBuilder) // set now, in case chainIndexVar is in the valBuilder
				exprBuilder.WriteByte('\n')
				exprBuilder.WriteString(ctx.ident(chainIndexVar) + " := " + inlineExpression(ctx, cexpr.Index) + "\n")

				switch typeinfer.Infer(cexpr.Index) {
				case "int", "": // either a map, or a slice
					exprBuilder.WriteString("if " + ctx.woofFunc("CanIndex", ctx.ident(chainValVar), ctx.ident(chainIndexVar)) + " {\n")

					valBuilder.WriteByte('[')
					valBuilder.WriteString(ctx.ident(chainIndexVar))
					valBuilder.WriteByte(']')

					//goland:noinspection GoDeferInLoop
					defer exprBuilder.WriteString("}\n")
				default: // this is a map
					exprBuilder.WriteString("if " + ctx.ident(chainValVar) + ", " + ctx.ident("ok") + " := " + ctx.ident(chainValVar) + "[" + ctx.ident(chainIndexVar) + "]; " + ctx.ident("ok") + " {\n")
					//goland:noinspection GoDeferInLoop
					defer exprBuilder.WriteString("}\n")
				}
			} else {
				valBuilder.WriteByte('[')
				valBuilder.WriteString(inlineExpression(ctx, cexpr.Index))
				valBuilder.WriteByte(']')
			}

			if cexpr.CheckValue {
				ctx.write("if ")
				inlineSetChainValVar(ctx, exprBuilder, valBuilder)
				exprBuilder.WriteString("; !" + ctx.woofFunc("IsZero", ctx.ident(chainValVar)) + " {\n")
				//goland:noinspection GoDeferInLoop
				defer exprBuilder.WriteString("}\n")
			}
		case file.DotIdentExpression:
			valBuilder.WriteByte('.')
			valBuilder.WriteString(cexpr.Ident.Ident)

			if cexpr.Check {
				ctx.write("if ")
				inlineSetChainValVar(ctx, exprBuilder, valBuilder)
				exprBuilder.WriteString("; !" + ctx.woofFunc("IsZero", ctx.ident(chainValVar)) + " {\n")
				//goland:noinspection GoDeferInLoop
				defer exprBuilder.WriteString("}\n")
			}
		case file.ParenExpression:
			valBuilder.WriteByte('(')
			for i, arg := range cexpr.Args {
				if i > 0 {
					valBuilder.WriteString(", ")
				}

				valBuilder.WriteString(inlineExpression(ctx, arg))
			}
			valBuilder.WriteByte(')')

			if cexpr.Check {
				ctx.write("if ")
				inlineSetChainValVar(ctx, exprBuilder, valBuilder)
				exprBuilder.WriteString("; !" + ctx.woofFunc("IsZero", ctx.ident(chainValVar)) + " {\n")
				//goland:noinspection GoDeferInLoop
				defer exprBuilder.WriteString("}\n")
			}
		case file.TypeAssertionExpression:
			var typBuilder strings.Builder
			typBuilder.WriteString(strings.Repeat("*", cexpr.PointerCount))
			if cexpr.Package != nil {
				typBuilder.WriteString(cexpr.Package.Ident)
				typBuilder.WriteByte('.')
			}
			typBuilder.WriteString(cexpr.Type.Ident)
			typ := typBuilder.String()

			if cexpr.Check {
				inlineSetChainValVar(ctx, exprBuilder, valBuilder)
				exprBuilder.WriteString("if " + ctx.ident(chainValVar) + ", " + ctx.ident("ok") + " := " + ctx.ident(chainValVar) + ".(" + typ + "); " + ctx.ident("ok") + " {\n")
				//goland:noinspection GoDeferInLoop
				defer exprBuilder.WriteString("}\n")
			} else {
				valBuilder.WriteString(".(")
				valBuilder.WriteString(typ)
				valBuilder.WriteByte(')')
			}
		}
	}

	checksPassed()
}

func inlineSetChainValVar(ctx *ctx, exprBuilder, valBuilder *strings.Builder) {
	exprBuilder.WriteString(ctx.ident(chainValVar) + " := " + valBuilder.String())
	valBuilder.Reset()
	valBuilder.WriteString(ctx.ident(chainValVar))
}

func inlineTernaryExpression(ctx *ctx, sb *strings.Builder, texpr file.TernaryExpression) {
	ctx.debugItemInline(texpr, "(generated) (see below)")
	sb.WriteString(ctx.woofFunc("Ternary", inlineCondition(ctx, texpr.Condition), inlineExpression(ctx, texpr.IfTrue), inlineExpression(ctx, texpr.IfFalse)))
}

func inlineGoExpression(ctx *ctx, sb *strings.Builder, gexpr file.GoExpression) {
	ctx.debugItemInline(gexpr, gexpr.Expression)
	sb.WriteString(gexpr.Expression)
}

func inlineStringExpression(ctx *ctx, sb *strings.Builder, sexpr file.StringExpression) {
	ctx.debugItemInline(sexpr, "(see below)")

	for i, exprItm := range sexpr.Contents {
		if i > 0 {
			sb.WriteString(" + ")
		}

		switch exprItm := exprItm.(type) {
		case file.StringExpressionText:
			s := string(sexpr.Quote) + exprItm.Text + string(sexpr.Quote)
			ctx.debugItemInline(exprItm, s)
			sb.WriteString(s)
		case file.StringExpressionInterpolation:
			if exprItm.FormatDirective == "" {
				ctx.debugItemInline(exprItm, "(see sub expressions)")
				sb.WriteString(ctx.woofFunc("Must", ctx.ident(ctxVar), ctx.woofQual("Stringify"), inlineExpression(ctx, exprItm.Expression)))
				continue
			}

			ctx.debugItemInline(exprItm, "[%"+exprItm.FormatDirective+"] (see below)")
			fmtString := strconv.Quote("%" + exprItm.FormatDirective)
			sb.WriteString(ctx.fmtFunc("Sprintf", fmtString, inlineExpression(ctx, exprItm.Expression)))
		}
	}
}

// ============================================================================
// Condition
// ======================================================================================

// inlineCondition writes a condition of, for example, an if.
//
// It is the exact same as inlineExpression, except for the following case:
//
// If condition is a [file.ChainExpression], writeCondition writes a function
// literal that is directly invoked and which returns true or false, depending
// on whether the chain expression's checks pass.
func inlineCondition(ctx *ctx, condition file.Expression) string {
	var sb strings.Builder

	for _, exprItm := range condition.Expressions {
		switch exprItm := exprItm.(type) {
		case file.ChainExpression:
			inlineConditionChainExpression(ctx, &sb, exprItm)
		case file.TernaryExpression:
			inlineTernaryExpression(ctx, &sb, exprItm)
		case file.GoExpression:
			inlineGoExpression(ctx, &sb, exprItm)
		case file.StringExpression:
			inlineStringExpression(ctx, &sb, exprItm)
		}
	}

	return sb.String()
}

func inlineConditionChainExpression(ctx *ctx, sb *strings.Builder, cexpr file.ChainExpression) {
	sb.WriteString("func () bool {\n")
	defer sb.WriteString("}()")

	var valBuilder strings.Builder
	valBuilder.WriteString(inlineExpression(ctx, file.Expression{Expressions: []file.ExpressionItem{cexpr.Root}}))

	if cexpr.CheckRoot {
		setChainValVar(ctx, &valBuilder)
		sb.WriteString("if ")
		sb.WriteString("; !" + ctx.woofFunc("IsZero", ctx.ident(chainValVar)) + " {\n")
	}

	inlineChainExprItems(ctx, sb, cexpr.Chain, &valBuilder, func() {
		sb.WriteString("return true\n")
	})
	if cexpr.CheckRoot {
		sb.WriteString("}\n")
	}

	sb.WriteString("return false\n")
}

// ============================================================================
// For Expression
// ======================================================================================

func forChainExpression(ctx *ctx, cexpr file.ChainExpression, writer func(string)) {
	var valBuilder strings.Builder
	valBuilder.WriteString(inlineExpression(ctx, file.Expression{Expressions: []file.ExpressionItem{cexpr.Root}}))

	if cexpr.Default != nil {
		ctx.inContext(func() {
			ctx.writeln(ctx.ident("ranger") + " := " + inlineExpression(ctx, *cexpr.Default))

			if cexpr.CheckRoot {
				setChainValVar(ctx, &valBuilder)
				ctx.write("if ")
				ctx.writeln("; !" + ctx.woofFunc("IsZero", ctx.ident(chainValVar)) + " {")
			}

			generateChainExprItems(ctx, cexpr.Chain, &valBuilder, func() {
				ctx.writeln(ctx.ident("ranger") + " := " + valBuilder.String())
			})
			if cexpr.CheckRoot {
				ctx.writeln("}")
			}

			writer(ctx.ident("ranger"))
		})

		return
	}

	if cexpr.CheckRoot {
		setChainValVar(ctx, &valBuilder)
		ctx.write("if ")
		ctx.writeln("; !" + ctx.woofFunc("IsZero", ctx.ident(chainValVar)) + " {")
		defer ctx.writeln("}")
	}

	generateChainExprItems(ctx, cexpr.Chain, &valBuilder, func() {
		writer(valBuilder.String())
	})
}

// ============================================================================
// contextExpression
// ======================================================================================

func generateContextExpression(ctx *ctx, expr file.Expression, escFunc, safeTyp string, normalizer func(string) string, writer func(func())) {
	if len(expr.Expressions) == 1 {
		switch exprItm := expr.Expressions[0].(type) {
		case file.ChainExpression:
			generateChainExpression(ctx, exprItm, nil, &escaper{name: escFunc}, writer)
			return
		case file.TernaryExpression:
			writer(func() {
				generateContextTernaryExpression(ctx, exprItm, escFunc, safeTyp, normalizer)
			})
			return
		case file.GoExpression:
			writer(func() {
				generateGoExpression(ctx, exprItm, &escaper{name: escFunc})
			})
			return
		case file.StringExpression:
			writer(func() {
				generateContextStringExpression(ctx, exprItm, escFunc, safeTyp, normalizer)
			})
			return
		}
	}

	panic(fmt.Errorf("generateContextExpression (%d:%d): attempting to generated contextual escaped complex expression"+
		" (this should've been caught during validation)", expr.Pos().Line, expr.Pos().Col))
}

func generateContextTernaryExpression(ctx *ctx, texpr file.TernaryExpression, escFunc, safeTyp string, normalizer func(string) string) {
	ctx.debugItem(texpr, "(generated) (see below)")
	ctx.flushGenerate()
	ctx.write("if ")
	ctx.write(inlineCondition(ctx, texpr.Condition))
	ctx.writeln(" {")
	generateContextExpression(ctx, texpr.IfTrue, escFunc, safeTyp, normalizer, nil)
	ctx.flushGenerate()
	ctx.writeln("} else {")
	generateContextExpression(ctx, texpr.IfFalse, escFunc, safeTyp, normalizer, nil)
	ctx.flushGenerate()
	ctx.writeln("}")
}

func generateContextStringExpression(ctx *ctx, sexpr file.StringExpression, escFunc, safeTyp string, normalizer func(string) string) {
	ctx.debugItem(sexpr, "(generated) (see below)")

	var b strings.Builder
	b.WriteString(ctx.woofQual(escFunc))
	b.WriteByte('(')

	for i, exprItm := range sexpr.Contents {
		if i > 0 {
			b.WriteString(", ")
		}

		switch exprItm := exprItm.(type) {
		case file.StringExpressionText:
			s := unquoteStringExpressionText(sexpr, exprItm)
			ctx.debugItem(exprItm, s)
			s = normalizer(s)
			b.WriteString(ctx.woofFunc(safeTyp, strconv.Quote(s)))
		case file.StringExpressionInterpolation:
			ctx.debugItem(exprItm, "(see sub expressions)")
			b.WriteString(inlineExpression(ctx, exprItm.Expression))
			continue
		}
	}

	ctx.write(ctx.contextFunc("Write", ctx.woofFunc("MustFunc", ctx.ident(ctxVar), "func() ("+safeTyp+", error) { return "+b.String()+" }")))
}

func unquoteStringExpressionText(sexpr file.StringExpression, txt file.StringExpressionText) string {
	s, _ := strconv.Unquote(string(sexpr.Quote) + txt.Text + string(sexpr.Quote))
	return s
}
