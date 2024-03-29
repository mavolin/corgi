package write

import (
	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/fileutil"
)

// ============================================================================
// If
// ======================================================================================

func _if(ctx *ctx, _if file.If) {
	ctx.flushGenerate()
	ctx.flushClasses()
	ctx.callUnclosedIfUnclosed()

	allClosed := true

	ctx.startScope(false)
	ctx.write("if ")
	ctx.write(inlineCondition(ctx, _if.Condition))
	ctx.writeln(" {")
	scope(ctx, _if.Then, false)
	ctx.flushGenerate()
	ctx.flushClasses()
	ctx.callClosedIfClosed()

	if ctx.endScope().startClosed != closed {
		allClosed = false
	}

	for _, elseIf := range _if.ElseIfs {
		ctx.startScope(false)

		ctx.write("} else if ")
		ctx.write(inlineCondition(ctx, elseIf.Condition))
		ctx.writeln(" {")
		scope(ctx, elseIf.Then, false)
		ctx.flushGenerate()
		ctx.flushClasses()
		ctx.callClosedIfClosed()

		if ctx.endScope().startClosed != closed {
			allClosed = false
		}
	}

	if _if.Else != nil {
		ctx.startScope(false)

		ctx.writeln("} else {")
		scope(ctx, _if.Else.Then, false)
		ctx.flushGenerate()
		ctx.flushClasses()
		ctx.callClosedIfClosed()

		if ctx.endScope().startClosed != closed {
			allClosed = false
		}
	}

	if ctx.scope().startClosed != closed {
		if allClosed && _if.Else != nil {
			ctx.scope().startClosed = closed
		} else {
			ctx.scope().startClosed = maybeClosed
		}
	}

	ctx.writeln("}")
}

// ============================================================================
// If Block
// ======================================================================================

func ifBlock(ctx *ctx, ifb file.IfBlock) {
	if ctx.mixin == nil {
		ifTemplateBlock(ctx, ifb)
		return
	}

	ifMixinBlock(ctx, ifb)
}

func ifTemplateBlock(ctx *ctx, ifb file.IfBlock) {
	ctx.debugItem(ifb, ifb.Name.Ident)

	if templateBlockFilled(ctx, ifb.Name.Ident) {
		ctx.debug("if block", "filled")
		scope(ctx, ifb.Then, false)
		return
	}

	for _, elseIf := range ifb.ElseIfs {
		ctx.debugItem(ifb, elseIf.Name.Ident)

		if templateBlockFilled(ctx, elseIf.Name.Ident) {
			ctx.debug("if block", "filled")

			scope(ctx, elseIf.Then, false)
			return
		}
	}

	if ifb.Else == nil {
		return
	}

	ctx.debugItem(ifb.Else, "filled")
	scope(ctx, ifb.Else.Then, false)
}

func ifMixinBlock(ctx *ctx, ifb file.IfBlock) {
	ctx.flushGenerate()
	ctx.flushClasses()
	ctx.callUnclosedIfUnclosed()

	allClosed := true

	ctx.startScope(false)
	ctx.writeln("if " + ctx.ident("mixinBlock_"+ifb.Name.Ident) + " != nil {")
	scope(ctx, ifb.Then, false)
	ctx.flushGenerate()
	ctx.flushClasses()
	ctx.callClosedIfClosed()

	if ctx.endScope().startClosed != closed {
		allClosed = false
	}

	for _, elseIf := range ifb.ElseIfs {
		ctx.startScope(false)

		ctx.writeln("} else if " + ctx.ident("mixinBlock_"+elseIf.Name.Ident) + " != nil {")
		scope(ctx, elseIf.Then, false)
		ctx.flushGenerate()
		ctx.flushClasses()
		ctx.callClosedIfClosed()

		if ctx.endScope().startClosed != closed {
			allClosed = false
		}
	}

	if ifb.Else != nil {
		ctx.startScope(false)

		ctx.writeln("} else {")
		scope(ctx, ifb.Else.Then, false)
		ctx.flushGenerate()
		ctx.flushClasses()
		ctx.callUnclosedIfUnclosed()

		if ctx.endScope().startClosed != closed {
			allClosed = false
		}
	}

	if ctx.scope().startClosed != closed {
		if allClosed && ifb.Else != nil {
			ctx.scope().startClosed = closed
		} else {
			ctx.scope().startClosed = maybeClosed
		}
	}

	ctx.writeln("}")
}

// ============================================================================
// Switch
// ======================================================================================

func _switch(ctx *ctx, sw file.Switch) {
	ctx.flushGenerate()
	ctx.flushClasses()
	ctx.callUnclosedIfUnclosed()

	allClosed := true

	ctx.write("switch ")
	if sw.Comparator != nil {
		ctx.write(inlineExpression(ctx, *sw.Comparator))
		ctx.write(" ")
	}
	ctx.writeln("{")

	for _, c := range sw.Cases {
		ctx.startScope(false)

		ctx.write("case ")
		ctx.write(inlineCondition(ctx, *c.Expression))
		ctx.writeln(":")
		scope(ctx, c.Then, false)
		ctx.flushGenerate()
		ctx.flushClasses()
		ctx.callClosedIfClosed()

		if ctx.endScope().startClosed != closed {
			allClosed = false
		}
	}

	if sw.Default != nil {
		ctx.startScope(false)

		ctx.writeln("default:")
		scope(ctx, sw.Default.Then, false)
		ctx.flushGenerate()
		ctx.flushClasses()
		ctx.callClosedIfClosed()

		if ctx.endScope().startClosed != closed {
			allClosed = false
		}
	}

	if ctx.scope().startClosed != closed {
		if allClosed && sw.Default != nil {
			ctx.scope().startClosed = closed
		} else {
			ctx.scope().startClosed = maybeClosed
		}
	}

	ctx.writeln("}")
}

// ============================================================================
// For
// ======================================================================================

func _for(ctx *ctx, f file.For) {
	_, attrLoop := fileutil.IsFirstNonControlAttr(f.Body)
	if !attrLoop {
		ctx.closeStartTag()

		nest := ctx.scope()
		ctx.startScope(false)
		defer func() {
			pop := ctx.endScope()
			if nest.startClosed != closed {
				if pop.startClosed == closed {
					ctx.scope().startClosed = maybeClosed
				}
			}
		}()
	}

	ctx.flushGenerate()
	ctx.flushClasses()

	if f.Expression == nil && len(f.Expression.Expressions) == 0 {
		ctx.writeln("for {")
		scope(ctx, f.Body, false)
		ctx.flushGenerate()
		ctx.flushClasses()
		ctx.writeln("}")
		return
	}

	if len(f.Expression.Expressions) == 1 {
		rangeExpr, ok := f.Expression.Expressions[0].(file.RangeExpression)
		if ok {
			forRange(ctx, f, rangeExpr)
			return
		}
	}

	ctx.writeln("for " + inlineExpression(ctx, *f.Expression) + " {")
	scope(ctx, f.Body, false)
	ctx.flushGenerate()
	ctx.flushClasses()
	ctx.writeln("}")
}

func forRange(ctx *ctx, f file.For, rangeExpr file.RangeExpression) {
	ctx.debugItem(rangeExpr, "(see below)")

	chainExpr, ok := rangeExpr.RangeExpression.Expressions[0].(file.ChainExpression)
	if ok {
		forChainExpression(ctx, chainExpr, func(rangerExpr string) {
			_forRange(ctx, f, rangeExpr, rangerExpr)
		})
		return
	}

	_forRange(ctx, f, rangeExpr, inlineExpression(ctx, rangeExpr.RangeExpression))
}

func _forRange(ctx *ctx, f file.For, rangeExpr file.RangeExpression, rangerExpr string) {
	if rangeExpr.Ordered {
		ctx.writeln("for _, " + ctx.ident("orderVal") + " := range " + ctx.woofFunc("OrderedMap", rangerExpr) + " {")
		if rangeExpr.Var1 != nil {
			if rangeExpr.Var1.Ident != "_" {
				if rangeExpr.Declares {
					ctx.writeln(rangeExpr.Var1.Ident + " := " + ctx.ident("orderVal") + ".K")
				} else {
					ctx.writeln(rangeExpr.Var1.Ident + " = " + ctx.ident("orderVal") + ".K")
				}
			}

			if rangeExpr.Var2 != nil && rangeExpr.Var2.Ident != "_" {
				if rangeExpr.Declares {
					ctx.writeln(rangeExpr.Var2.Ident + " := " + ctx.ident("orderVal") + ".V")
				} else {
					ctx.writeln(rangeExpr.Var2.Ident + " = " + ctx.ident("orderVal") + ".V")
				}
			}
		}
	} else {
		ctx.write("for ")
		if rangeExpr.Var1 != nil {
			ctx.write(rangeExpr.Var1.Ident)

			if rangeExpr.Var2 != nil {
				ctx.write(", " + rangeExpr.Var2.Ident)
			}

			ctx.write(" := ")
		}
		ctx.writeln("range " + rangerExpr + " {")
	}

	scope(ctx, f.Body, false)
	ctx.flushGenerate()
	ctx.flushClasses()
	ctx.writeln("}")
}
