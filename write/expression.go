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
// If txtEsc is set and expr contains a single ExpressionItem of type
// [file.StringExpression], then the text parts of the string will be escaped
// using txtEsc, everything else will be escaped using exprEsc.
//
// If expr contains a single ExpressionItem of type [file.ChainExpression],
// generateExpression generates code that performs the checks of the chain
// expression, calling writer with a function that generates the resolved
// expression at the point in code where all checks pass.
//
// If the chain expression has a default, writer is instead called immediately,
// then when genExpr is called, either the chain expression result or the
// default is written.
//
// If expr contains a single [file.TernaryExpression], generateExpression
// similarly writes an if else, calling write twice, once for the ifTrue and
// once for the ifFalse expression.
//
// If expr is any other expression, writer is called immediately with a
// function generating the expression.
//
// If writer is nil, generateExpression calls genExpr directly, instead of
// passing it to writer.
//
// writer must call genExpr only once.
func generateExpression(ctx *ctx, expr file.Expression, txtEsc *textEscaper, exprEsc *expressionEscaper, writer func(genExpr func())) {
	if ctx.debugEnabled {
		txtEscName := "none"
		if txtEsc != nil {
			txtEscName = txtEsc.name
		}
		exprEscName := "none"
		if exprEsc != nil {
			exprEscName = exprEsc.funcName
		}
		ctx.debugItem(expr, fmt.Sprintf("(txt escaper: %s, expr esc: %s) (see below)", txtEscName, exprEscName))
	}

	if len(expr.Expressions) == 0 {
		return
	}

	if writer == nil {
		writer = func(genExpr func()) { genExpr() }
	}

	if len(expr.Expressions) == 1 {
		switch exprItm := expr.Expressions[0].(type) {
		case file.ChainExpression:
			generateChainExpression(ctx, exprItm, nil, exprEsc, writer)
			return
		case file.TernaryExpression:
			writer(func() {
				ctx.flushGenerate()
				generateTernaryExpression(ctx, exprItm, txtEsc, exprEsc)
			})
			return
		case file.GoExpression:
			if txtEsc != nil {
				if num, err := strconv.ParseInt(exprItm.Expression, 10, 64); err == nil {
					writer(func() {
						ctx.generate(ctx.stringify(num), txtEsc)
					})
					return
				} else if num, err := strconv.ParseUint(exprItm.Expression, 10, 64); err == nil {
					writer(func() {
						ctx.generate(ctx.stringify(num), txtEsc)
					})
					return
				} else if num, err := strconv.ParseFloat(exprItm.Expression, 64); err == nil {
					writer(func() {
						ctx.generate(ctx.stringify(num), txtEsc)
					})
					return
				}
			}

			writer(func() {
				ctx.flushGenerate()
				generateGoExpression(ctx, exprItm, exprEsc)
			})
			return
		case file.StringExpression:
			if txtEsc != nil && len(exprItm.Contents) == 1 {
				txt, ok := exprItm.Contents[0].(file.StringExpressionText)
				if ok {
					s := unquoteStringExpressionText(exprItm, txt)
					s = strings.ReplaceAll(s, "##", "#")

					writer(func() {
						ctx.generate(s, txtEsc)
					})
					return
				}
			}

			writer(func() {
				ctx.flushGenerate()
				generateStringExpression(ctx, exprItm, txtEsc, exprEsc)
			})
			return
		}
	}

	writer(func() {
		ctx.flushGenerate()
		ctx.generateExpr(inlineExpression(ctx, expr), exprEsc)
	})
}

const (
	chainValVar   = "chainVal"
	chainIndexVar = "chainIndex"
)

func generateChainExpression(ctx *ctx, cexpr file.ChainExpression, defaultExpr *file.Expression, esc *expressionEscaper, writer func(func())) {
	ctx.debugItem(cexpr, "(see below)")

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
				ctx.generateExpr(valBuilder.String(), esc)
				ctx.flushGenerate()
				ctx.writeln("goto " + checksPassedGoto)
			})
			if cexpr.CheckRoot {
				ctx.writeln("}")
			}

			ctx.generateExpr(inlineExpression(ctx, *defaultExpr), esc)
			ctx.flushGenerate()
			ctx.writeln(checksPassedGoto + ":")
		})
		return
	}

	ctx.flushGenerate()
	ctx.flushClasses()
	if cexpr.CheckRoot {
		ctx.write("if ")
		setChainValVar(ctx, &valBuilder)
		ctx.writeln("; !" + ctx.woofFunc("IsZero", ctx.ident(chainValVar)) + " {")
	} else {
		ctx.writeln("{")
	}
	defer ctx.writeln("}")

	generateChainExprItems(ctx, cexpr.Chain, &valBuilder, func() {
		writer(func() {
			ctx.flushGenerate()
			ctx.generateExpr(strings.Repeat("*", cexpr.DerefCount)+valBuilder.String(), esc)
		})
		ctx.flushGenerate()
		ctx.flushClasses()
	})
}

func generateChainExprItems(ctx *ctx, cexprItms []file.ChainExpressionItem, valBuilder *strings.Builder, checksPassed func()) {
	for _, cexprItm := range cexprItms {
		switch cexpr := cexprItm.(type) {
		case file.IndexExpression:
			if cexpr.CheckIndex {
				ctx.writeln("{")
				//goland:noinspection GoDeferInLoop
				defer ctx.writeln("}")
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
				ctx.writeln("")
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

func generateTernaryExpression(ctx *ctx, texpr file.TernaryExpression, txtEsc *textEscaper, exprEsc *expressionEscaper) {
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

func generateGoExpression(ctx *ctx, gexpr file.GoExpression, esc *expressionEscaper) {
	ctx.debugItem(gexpr, gexpr.Expression)
	ctx.generateExpr(gexpr.Expression, esc)
}

func generateStringExpression(ctx *ctx, sexpr file.StringExpression, txtEsc *textEscaper, exprEsc *expressionEscaper) {
	ctx.debugItem(sexpr, "(generated) (see below)")

	for _, exprItm := range sexpr.Contents {
		switch exprItm := exprItm.(type) {
		case file.StringExpressionText:
			s := unquoteStringExpressionText(sexpr, exprItm)
			s = strings.ReplaceAll(s, "##", "#")

			ctx.debugItem(exprItm, s)

			if txtEsc != nil {
				ctx.generate(s, txtEsc)
			} else {
				ctx.generateExpr(strconv.Quote(exprItm.Text), exprEsc)
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
		ctx.write("if ")
		setChainValVar(ctx, &valBuilder)
		ctx.writeln("; !" + ctx.woofFunc("IsZero", ctx.ident(chainValVar)) + " {")
	}

	var checksPassedGoto string
	if cexpr.Default != nil {
		checksPassedGoto = ctx.nextGotoIdent()
	}

	generateChainExprItems(ctx, cexpr.Chain, &valBuilder, func() {
		ctx.flushGenerate()
		onValue(strings.Repeat("*", cexpr.DerefCount) + valBuilder.String())
		ctx.flushGenerate()
		if cexpr.Default != nil {
			ctx.writeln("goto " + checksPassedGoto)
		}
	})
	if cexpr.CheckRoot {
		ctx.writeln("}")
	}

	if cexpr.Default != nil {
		ctx.flushGenerate()
		onValue(inlineExpression(ctx, *cexpr.Default))
		ctx.flushGenerate()
		ctx.writeln(checksPassedGoto + ":")
	}

	ctx.flushGenerate()
}

// ============================================================================
// Inline Expression
// ======================================================================================

// yields an unescaped expression.
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
		default:
			ctx.youShouldntSeeThisError(fmt.Errorf("unknown expression item %T", exprItm))
		}
	}

	return sb.String()
}

// ============================================================================
// Mixin Arg Expression
// ======================================================================================

func mixinArgExpression(ctx *ctx, param file.MixinParam, arg file.MixinArg) {
	if len(arg.Value.Expressions) == 1 {
		cExpr, ok := arg.Value.Expressions[0].(file.ChainExpression)
		if ok {
			var typ string
			if param.Type != nil {
				typ = param.Type.Type
			} else {
				typ = param.InferredType
			}

			ctx.write(mixinArgChainExpression(ctx, cExpr, typ, param.Default != nil))
			return
		}
	}

	if param.Default != nil {
		ctx.write(ctx.woofFunc("Ptr", inlineExpression(ctx, arg.Value)))
	} else {
		ctx.write(inlineExpression(ctx, arg.Value))
	}
}

func mixinArgChainExpression(ctx *ctx, cexpr file.ChainExpression, typ string, attrHasDefault bool) string {
	var sb strings.Builder

	sb.WriteString(`func () `)
	if attrHasDefault {
		sb.WriteByte('*')
	}
	sb.WriteString(typ)
	sb.WriteString(" {\n")

	var valBuilder strings.Builder
	valBuilder.WriteString(inlineExpression(ctx, file.Expression{Expressions: []file.ExpressionItem{cexpr.Root}}))

	if cexpr.CheckRoot {
		sb.WriteString("if ")
		setChainValVar(ctx, &valBuilder)
		sb.WriteString("; !" + ctx.woofFunc("IsZero", ctx.ident(chainValVar)) + " {\n")
	}

	mixinArgChainExprItms(ctx, &sb, cexpr.Chain, &valBuilder, func() {
		sb.WriteString("return ")
		if attrHasDefault {
			sb.WriteString(ctx.woofFunc("Ptr", strings.Repeat("*", cexpr.DerefCount)+valBuilder.String()))
		} else {
			sb.WriteString(valBuilder.String())
		}
		sb.WriteByte('\n')
	})
	if cexpr.CheckRoot {
		sb.WriteString("}\n")
	}

	if cexpr.Default != nil {
		sb.WriteString("return ")
		if attrHasDefault {
			sb.WriteString(ctx.woofFunc("Ptr", inlineExpression(ctx, *cexpr.Default)))
		} else {
			sb.WriteString(inlineExpression(ctx, *cexpr.Default))
		}
		sb.WriteByte('\n')
	} else {
		sb.WriteString("return nil\n")
	}

	sb.WriteString("}()")
	return sb.String()
}

func mixinArgChainExprItms(ctx *ctx, exprBuilder *strings.Builder, cexprItms []file.ChainExpressionItem, valBuilder *strings.Builder, checksPassed func()) {
	for _, cexprItm := range cexprItms {
		switch cexpr := cexprItm.(type) {
		case file.IndexExpression:
			if cexpr.CheckIndex {
				exprBuilder.WriteString("{\n")
				//goland:noinspection GoDeferInLoop
				defer exprBuilder.WriteString("}\n")

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
				exprBuilder.WriteString("if ")
				inlineSetChainValVar(ctx, exprBuilder, valBuilder)
				exprBuilder.WriteString("; !" + ctx.woofFunc("IsZero", ctx.ident(chainValVar)) + " {\n")
				//goland:noinspection GoDeferInLoop
				defer exprBuilder.WriteString("}\n")
			}
		case file.DotIdentExpression:
			valBuilder.WriteByte('.')
			valBuilder.WriteString(cexpr.Ident.Ident)

			if cexpr.Check {
				exprBuilder.WriteString("if ")
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
				exprBuilder.WriteString("if ")
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
				exprBuilder.WriteByte('\n')
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
			s := string(sexpr.Quote) + strings.ReplaceAll(exprItm.Text, "##", "#") + string(sexpr.Quote)
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
		sb.WriteString("if ")
		setChainValVar(ctx, &valBuilder)
		sb.WriteString("; !" + ctx.woofFunc("IsZero", ctx.ident(chainValVar)) + " {\n")
	}

	mixinArgChainExprItms(ctx, sb, cexpr.Chain, &valBuilder, func() {
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
				ctx.write("if ")
				setChainValVar(ctx, &valBuilder)
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
		ctx.write("if ")
		setChainValVar(ctx, &valBuilder)
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

func generateContextExpression(ctx *ctx, expr file.Expression, ctxEsc contextEscaper, writer func(func())) {
	if writer == nil {
		writer = func(genExpr func()) { genExpr() }
	}

	if len(expr.Expressions) == 1 {
		switch exprItm := expr.Expressions[0].(type) {
		case file.ChainExpression:
			valueChainExpression(ctx, exprItm, func(expr string) {
				writer(func() {
					ctx.flushGenerate()
					ctx.writeln(ctx.woofFunc("WriteAnys", ctx.ident(ctxVar), ctx.woofQual(ctxEsc.funcName), expr))
				})
			})
			return
		case file.TernaryExpression:
			writer(func() {
				ctx.flushGenerate()
				generateContextTernaryExpression(ctx, exprItm, ctxEsc)
			})
			return
		case file.StringExpression:
			writer(func() {
				ctx.flushGenerate()
				generateContextStringExpression(ctx, exprItm, ctxEsc)
			})
			return
		}
	}

	writer(func() {
		ctx.flushGenerate()
		ctx.writeln(ctx.woofFunc("WriteAnys", ctx.ident(ctxVar), ctx.woofQual(ctxEsc.funcName), inlineExpression(ctx, expr)))
	})
}

func generateContextTernaryExpression(ctx *ctx, texpr file.TernaryExpression, ctxEsc contextEscaper) {
	ctx.debugItem(texpr, "(generated) (see below)")
	ctx.flushGenerate()
	ctx.flushClasses()
	ctx.write("if ")
	ctx.write(inlineCondition(ctx, texpr.Condition))
	ctx.writeln(" {")
	generateContextExpression(ctx, texpr.IfTrue, ctxEsc, nil)
	ctx.flushGenerate()
	ctx.flushClasses()
	ctx.writeln("} else {")
	generateContextExpression(ctx, texpr.IfFalse, ctxEsc, nil)
	ctx.flushGenerate()
	ctx.flushClasses()
	ctx.writeln("}")
}

func generateContextStringExpression(ctx *ctx, sexpr file.StringExpression, ctxEsc contextEscaper) {
	ctx.debugItem(sexpr, "(generated) (see below)")

	var b strings.Builder
	for i, exprItm := range sexpr.Contents {
		if i > 0 {
			b.WriteString(", ")
		}

		switch exprItm := exprItm.(type) {
		case file.StringExpressionText:
			s := unquoteStringExpressionText(sexpr, exprItm)
			s = strings.ReplaceAll(s, "##", "#")
			ctx.debugItem(exprItm, s)
			if ctxEsc.normalizer != nil {
				s = ctxEsc.normalizer(s)
			}
			b.WriteString(ctx.woofFunc(ctxEsc.safeType, strconv.Quote(s)))
		case file.StringExpressionInterpolation:
			ctx.debugItem(exprItm, "(see sub expressions)")
			b.WriteString(inlineExpression(ctx, exprItm.Expression))
			continue
		}
	}

	ctx.writeln(ctx.woofFunc("WriteAnys", ctx.ident(ctxVar), ctx.woofQual(ctxEsc.funcName), b.String()))
}

func unquoteStringExpressionText(sexpr file.StringExpression, txt file.StringExpressionText) string {
	s, _ := strconv.Unquote(string(sexpr.Quote) + txt.Text + string(sexpr.Quote))
	return s
}
