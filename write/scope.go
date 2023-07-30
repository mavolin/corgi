package write

import (
	"fmt"

	"github.com/mavolin/corgi/file"
)

func scope(ctx *ctx, s file.Scope) {
	for _, itm := range s {
		switch itm := itm.(type) {
		case file.CorgiComment:
			// ignore

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
}
