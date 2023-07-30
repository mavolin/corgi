package write

import "github.com/mavolin/corgi/file"

// ============================================================================
// If
// ======================================================================================

func _if(ctx *ctx, _if file.If) {
	ctx.flushGenerate()
	ctx.flushClasses()
	ctx.callUnclosedIfUnclosed()

	cl := ctx.closed.Peek()
	ctx.closed.Push(cl)

	var allClosed bool

	ctx.write("if ")
	ctx.write(inlineCondition(ctx, _if.Condition))
	ctx.writeln(" {")
	scope(ctx, _if.Then)
	ctx.flushGenerate()
	ctx.flushClasses()
	ctx.callClosedIfClosed()

	if ctx.closed.Pop() != closed {
		allClosed = false
	}

	for _, elseIf := range _if.ElseIfs {
		ctx.closed.Push(cl)

		ctx.write("} else if ")
		ctx.write(inlineCondition(ctx, elseIf.Condition))
		ctx.writeln(" {")
		scope(ctx, elseIf.Then)
		ctx.flushGenerate()
		ctx.flushClasses()
		ctx.callClosedIfClosed()

		if ctx.closed.Pop() != closed {
			allClosed = false
		}
	}

	if _if.Else != nil {
		ctx.closed.Push(cl)

		ctx.writeln("} else {")
		scope(ctx, _if.Else.Then)
		ctx.flushGenerate()
		ctx.flushClasses()
		ctx.callClosedIfClosed()

		if ctx.closed.Pop() != closed {
			allClosed = false
		}
	}

	if cl != closed {
		if allClosed && _if.Else != nil {
			ctx.closed.Swap(closed)
		} else {
			ctx.closed.Swap(maybeClosed)
		}
	}

	ctx.writeln("}")
}

// ============================================================================
// If Block
// ======================================================================================

func ifBlock(ctx *ctx, ifb file.IfBlock) {
	// todo
}

// ============================================================================
// Switch
// ======================================================================================

func _switch(ctx *ctx, sw file.Switch) {
	ctx.flushGenerate()
	ctx.flushClasses()
	ctx.callUnclosedIfUnclosed()

	cl := ctx.closed.Peek()
	ctx.closed.Push(cl)

	var allClosed bool

	ctx.write("switch ")
	if sw.Comparator != nil {
		ctx.write(inlineExpression(ctx, *sw.Comparator))
		ctx.write(" ")
	}
	ctx.writeln("{")

	for _, c := range sw.Cases {
		ctx.closed.Push(cl)

		ctx.write("case ")
		ctx.write(inlineCondition(ctx, *c.Expression))
		ctx.writeln(":")
		scope(ctx, c.Then)
		ctx.flushGenerate()
		ctx.flushClasses()
		ctx.callClosedIfClosed()

		if ctx.closed.Pop() != closed {
			allClosed = false
		}
	}

	if sw.Default != nil {
		ctx.closed.Push(cl)

		ctx.writeln("default:")
		scope(ctx, sw.Default.Then)
		ctx.flushGenerate()
		ctx.flushClasses()
		ctx.callClosedIfClosed()

		if ctx.closed.Pop() != closed {
			allClosed = false
		}
	}

	if cl != closed {
		if allClosed && sw.Default != nil {
			ctx.closed.Swap(closed)
		} else {
			ctx.closed.Swap(maybeClosed)
		}
	}

	ctx.writeln("}")
}

// ============================================================================
// For
// ======================================================================================

func _for(ctx *ctx, f file.For) {
	ctx.flushGenerate()
	ctx.flushClasses()
	ctx.callUnclosedIfUnclosed()

	cl := ctx.closed.Peek()
	ctx.closed.Push(cl)

	defer func() {
		if cl != closed {
			if ctx.closed.Pop() == closed {
				ctx.closed.Swap(maybeClosed)
			}
		}
	}()

	if f.Expression == nil && len(f.Expression.Expressions) == 0 {
		ctx.writeln("for {")
		scope(ctx, f.Body)
		ctx.flushGenerate()
		ctx.flushClasses()
		ctx.callClosedIfClosed()
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
	scope(ctx, f.Body)
	ctx.flushGenerate()
	ctx.flushClasses()
	ctx.writeln("}")
}

func forRange(ctx *ctx, f file.For, rangeExpr file.RangeExpression) {
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
			ctx.writeln(rangeExpr.Var1.Ident + " := " + ctx.ident("orderVal") + ".K")

			if rangeExpr.Var2 != nil {
				ctx.writeln(rangeExpr.Var1.Ident + " := " + ctx.ident("orderVal") + ".V")
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

	scope(ctx, f.Body)
	ctx.flushGenerate()
	ctx.writeln("}")
}
