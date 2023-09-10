package write

import (
	"fmt"

	"github.com/mavolin/corgi/file"
)

func scope(ctx *ctx, s file.Scope, needsCodeScoping bool) {
	if needsCodeScoping {
		for _, itm := range s {
			_, ok := itm.(file.Code)
			if ok {
				ctx.writeln("{")
				//goland:noinspection GoDeferInLoop
				defer ctx.writeln("}")
				break
			}
		}
	}

	ctx.startScope(true)
	for _, itm := range s {
		scopeItem(ctx, itm)
	}
	ctx.endScope()
}

func scopeItem(ctx *ctx, itm file.ScopeItem) {
	switch itm := itm.(type) {
	case file.CorgiComment:
	// ignore

	case file.Block:
		block(ctx, itm)
	case file.BlockExpansion:
		blockExpansion(ctx, itm)
	case file.Code:
		code(ctx, itm)
	case file.If:
		_if(ctx, itm)
	case file.IfBlock:
		ifBlock(ctx, itm)
	case file.Switch:
		_switch(ctx, itm)
	case file.For:
		_for(ctx, itm)
	case file.Doctype:
		doctype(ctx, itm)
	case file.HTMLComment:
		htmlComment(ctx, itm)
	case file.Element:
		element(ctx, itm)
	case file.DivShorthand:
		divShorthand(ctx, itm)
	case file.CommandFilter:
		commandFilter(ctx, itm)
	case file.RawFilter:
		rawFilter(ctx, itm)
	case file.Include:
		include(ctx, itm)
	case file.Mixin:
		scopeMixin(ctx, itm)
	case file.MixinCall:
		mixinCall(ctx, itm)
	case file.Return:
		_return(ctx, itm)
	case file.And:
		and(ctx, itm)
	case file.InlineText:
		inlineText(ctx, itm)
	case file.ArrowBlock:
		arrowBlock(ctx, itm)
	default:
		ctx.youShouldntSeeThisError(fmt.Errorf("%s:%d:%d: unknown scope item %T", ctx.currentFile().Name,
			itm.Pos().Line, itm.Pos().Col, itm))
	}
}
