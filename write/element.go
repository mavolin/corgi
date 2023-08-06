package write

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/woof"
)

// ============================================================================
// Doctype
// ======================================================================================

func doctype(ctx *ctx, _ file.Doctype) {
	ctx.closeTag()

	ctx.generate("<!doctype html>", nil)
}

// ============================================================================
// Comment
// ======================================================================================

var htmlCommentEscaper = strings.NewReplacer("-->", "-- >")

func htmlComment(ctx *ctx, c file.HTMLComment) {
	ctx.closeTag()

	ctx.generate("<!--", nil)

	for i, ln := range c.Lines {
		if i > 0 {
			ctx.generate("\n", nil)
		}

		ctx.generate(htmlCommentEscaper.Replace(ln.Comment), nil)
	}

	ctx.generate("-->", nil)
}

// ============================================================================
// Element
// ======================================================================================

func element(ctx *ctx, el file.Element) {
	ctx.debugItem(el, el.Name)
	ctx.startElem(el.Name, el.Void)

	for _, acoll := range el.Attributes {
		attributeCollection(ctx, acoll)
	}

	scope(ctx, el.Body)

	ctx.debugItem(el, "/"+el.Name)
	ctx.closeElem()
}

func divShorthand(ctx *ctx, dsh file.DivShorthand) {
	ctx.debugItem(dsh, "")

	ctx.startElem("div", false)

	for _, acoll := range dsh.Attributes {
		attributeCollection(ctx, acoll)
	}

	scope(ctx, dsh.Body)

	ctx.debugItem(dsh, "/div")
	ctx.closeElem()
}

// ============================================================================
// And
// ======================================================================================

func and(ctx *ctx, and file.And) {
	ctx.debugItem(and, "(see attributes below)")

	for _, acoll := range and.Attributes {
		attributeCollection(ctx, acoll)
	}
}

// ============================================================================
// AttributeCollection
// ======================================================================================

func attributeCollection(ctx *ctx, acoll file.AttributeCollection) {
	switch acoll := acoll.(type) {
	case file.IDShorthand:
		idShorthand(ctx, acoll)
	case file.ClassShorthand:
		classShorthand(ctx, acoll)
	case file.AttributeList:
		attributeList(ctx, acoll)
	default:
		panic(fmt.Errorf("unrecognized attribute collection %T (you shouldn't see this error, please open an issue)", acoll))
	}
}

// ==================================== IDShorthand =====================================

func idShorthand(ctx *ctx, idSh file.IDShorthand) {
	ctx.debugItem(idSh, idSh.ID)
	ctx.generateStringAttr("id", idSh.ID)
}

// =================================== ClassShorthand ===================================

func classShorthand(ctx *ctx, csh file.ClassShorthand) {
	ctx.debugItem(csh, csh.Name)
	ctx.bufClass(csh.Name)
}

// =================================== AttributeList ====================================

func attributeList(ctx *ctx, alist file.AttributeList) {
	ctx.debugItem(alist, "(see individual attributes)")

	for _, attr := range alist.Attributes {
		attribute(ctx, attr)
	}
}

// ============================================================================
// Attribute
// ======================================================================================

func attribute(ctx *ctx, attr file.Attribute) {
	switch attr := attr.(type) {
	case file.SimpleAttribute:
		simpleAttribute(ctx, attr)
	case file.AndPlaceholder:
		andPlaceholder(ctx, attr)
	case file.MixinCallAttribute:
		mixinCallAttribute(ctx, attr)
	default:
		panic(fmt.Errorf("unrecognized attribute %T (you shouldn't see this error, please open an issue)", attr))
	}
}

// ================================== SimpleAttribute ===================================

func simpleAttribute(ctx *ctx, sattr file.SimpleAttribute) {
	ctx.debugItem(sattr, sattr.Name)

	if sattr.Name == "class" {
		classAttribute(ctx, sattr)
		return
	}

	if sattr.Value == nil {
		ctx.generate(" "+sattr.Name, nil)
		return
	}

	attrType := woof.AttrType(sattr.Name)

	if len(sattr.Value.Expressions) == 1 {
		sexpr, ok := sattr.Value.Expressions[0].(file.StringExpression)
		if ok {
			if len(sexpr.Contents) == 0 {
				ctx.generate(" "+sattr.Name, nil)
				return
			} else if len(sexpr.Contents) == 1 {
				txt, ok := sexpr.Contents[0].(file.StringExpressionText)
				if ok {
					ctx.generateStringAttr(sattr.Name, unquoteStringExpressionText(sexpr, txt))
					return
				}
			}

			if attrType == woof.ContentTypePlain {
				ctx.generateExpression(*sattr.Value, nil, attrEscaper, func(f func()) {
					ctx.generate(` `+sattr.Name+`="`, nil)
					f()
					ctx.generate(`"`, nil)
				})
				return
			}
		}

		if attrType == woof.ContentTypePlain {
			cexpr, ok := sattr.Value.Expressions[0].(file.ChainExpression)
			if ok {
				valueChainExpression(ctx, cexpr, func(expr string) {
					ctx.writeln(ctx.woofFunc("WriteAttr", ctx.ident(ctxVar), strconv.Quote(sattr.Name), expr, ctx.woofQual("EscapeHTMLAttr")))
				})
				return
			}
		}
	}

	if attrType == woof.ContentTypePlain {
		expr := inlineExpression(ctx, *sattr.Value)
		ctx.writeln(ctx.woofFunc("WriteAttr", ctx.ident(ctxVar), strconv.Quote(sattr.Name), expr, ctx.woofQual("EscapeHTMLAttr")))
	}

	switch attrType { //nolint:exhaustive
	case woof.ContentTypeCSS:
		ctx.generateExpression(*sattr.Value, attrEscaper, cssEscaper, func(f func()) {
			ctx.generate(` `+sattr.Name+`="`, nil)
			f()
			ctx.generate(`"`, nil)
		})
	case woof.ContentTypeHTML:
		ctx.generateExpression(*sattr.Value, attrEscaper, nil, func(f func()) {
			ctx.generate(` `+sattr.Name+`="`, nil)
			f()
			ctx.generate(`"`, nil)
		})
	case woof.ContentTypeURL:
		generateContextExpression(ctx, *sattr.Value, "FilterURL", "URL", func(s string) string {
			u := woof.NormalizeURL(woof.URL(s))
			return string(u)
		}, func(f func()) {
			ctx.generate(` `+sattr.Name+`="`, nil)
			f()
			ctx.generate(`"`, nil)
		})
	case woof.ContentTypeSrcset:
		generateContextExpression(ctx, *sattr.Value, "FilterSrcset", "Srcset", func(s string) string {
			u := woof.NormalizeURL(woof.URL(s))
			return string(u)
		}, func(f func()) {
			ctx.generate(` `+sattr.Name+`="`, nil)
			f()
			ctx.generate(`"`, nil)
		})
	default:
		panic(fmt.Errorf("unrecognized content type %v (you shouldn't see this error, please open an issue)", attrType))
	}

}

func classAttribute(ctx *ctx, attr file.SimpleAttribute) {
	if attr.Value == nil || len(attr.Value.Expressions) == 0 {
		return
	}

	switch exprItm := attr.Value.Expressions[0].(type) {
	case file.StringExpression:
		if len(exprItm.Contents) == 0 {
			return
		} else if len(exprItm.Contents) == 1 {
			txt, ok := exprItm.Contents[0].(file.StringExpressionText)
			if ok {
				ctx.bufClass(unquoteStringExpressionText(exprItm, txt))
				return
			}
		}
	case file.ChainExpression:
		valueChainExpression(ctx, exprItm, func(expr string) {
			ctx.writeln(ctx.contextFunc("BufferClass", expr))
		})
	default:
		ctx.writeln(ctx.contextFunc("BufferClass", inlineExpression(ctx, *attr.Value)))
	}
}

// =================================== AndPlaceholder ===================================

const andPlaceholderFunc = "andPlaceholder"

func andPlaceholder(ctx *ctx, aph file.AndPlaceholder) {
	ctx.debugItem(aph, "(see below)")
	ctx.writeln("if " + ctx.ident(andPlaceholderFunc) + " != nil {")
	ctx.writeln("  " + ctx.ident(andPlaceholderFunc) + "()")
	ctx.writeln("}")
}

// ================================= MixinCallAttribute =================================

func mixinCallAttribute(ctx *ctx, mcAttr file.MixinCallAttribute) {
	ctx.generate(mcAttr.Name+`="`, nil)
	defer ctx.generate(`"`, nil)

	ctx.txtEscaper.Push(attrEscaper)
	defer ctx.txtEscaper.Pop()

	interpolationValueMixinCall(ctx, mcAttr.MixinCall, mcAttr.Value)
}
