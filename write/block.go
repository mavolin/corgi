package write

import "github.com/mavolin/corgi/file"

// ============================================================================
// Block
// ======================================================================================

func block(ctx *ctx, b file.Block) {
	ctx.debugItem(b, b.Name.Ident)

	if ctx.mixin != nil {
		ctx.flushGenerate()
		ctx.flushClasses()
		ctx.callUnclosedIfUnclosed()
		ctx.writeln("if " + ctx.ident("mixinBlock_"+b.Name.Ident) + " != nil {")
		ctx.writeln("  " + ctx.ident("mixinBlock_"+b.Name.Ident) + "()")
		if len(b.Body) > 0 {
			ctx.writeln("} else {")
			scope(ctx, b.Body)
			ctx.flushGenerate()
			ctx.flushClasses()
			ctx.callClosedIfClosed()
		}
		ctx.writeln("}")
		ctx.closed.Swap(maybeClosed)

		return
	}

	fill, stackPos := resolveTemplateBlock(ctx, b)

	if fill.Type == file.BlockTypeAppend {
		scope(ctx, b.Body)
	}

	oldStart := ctx.stackStart
	ctx.stackStart = stackPos
	scope(ctx, fill.Body)
	ctx.stackStart = oldStart

	if fill.Type == file.BlockTypePrepend {
		scope(ctx, b.Body)
	}
}

func resolveTemplateBlock(ctx *ctx, call file.Block) (b file.Block, stackPos int) {
	stack := ctx.stack()[1:]
	for i := len(stack) - 1; i >= 0; i-- {
		f := stack[i]
		for _, itm := range f.Scope {
			fill, ok := itm.(file.Block)
			if !ok {
				continue
			}

			if call.Name.Ident == fill.Name.Ident {
				return fill, i
			}
		}
	}

	return call, ctx.stackStart
}

func templateBlockFilled(ctx *ctx, name string) bool {
	stack := ctx.stack()[1:]
	for i := len(stack) - 1; i >= 0; i-- {
		f := stack[i]
		for _, itm := range f.Scope {
			fill, ok := itm.(file.Block)
			if !ok {
				continue
			}

			if name == fill.Name.Ident {
				return true
			}
		}
	}

	return false
}

// ============================================================================
// BlockExpansion
// ======================================================================================

func blockExpansion(ctx *ctx, bexp file.BlockExpansion) {
	scope(ctx, file.Scope{bexp.Item})
}
