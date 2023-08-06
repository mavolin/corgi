package write

import (
	"fmt"

	"github.com/mavolin/corgi/file"
)

func scope(ctx *ctx, s file.Scope) {
	ctx.mixinFuncNames.startScope(ctx)
	defer ctx.mixinFuncNames.endScope(ctx)

	for _, itm := range s {
		scopeItem(ctx, itm)
	}
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
		panic(fmt.Errorf("%d:%d: unknown scope item %T", itm.Pos().Line, itm.Pos().Col, itm))
	}
}
