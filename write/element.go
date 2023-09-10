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
	ctx.closeStartTag()

	ctx.generate("<!doctype html>", nil)
}

// ============================================================================
// Comment
// ======================================================================================

var htmlCommentEscaper = strings.NewReplacer("-->", "-- >")

func htmlComment(ctx *ctx, c file.HTMLComment) {
	ctx.closeStartTag()

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

	var success bool
	switch el.Name {
	case "style":
		success = minifyStyleElement(ctx, el)
	case "script":
		success = minifyScriptElement(ctx, el)
	}

	if !success {
		scope(ctx, el.Body)
	}

	ctx.closeStartTag()

	ctx.debugItem(el, "/"+el.Name)
	ctx.closeElem()
}

func minifyStyleElement(ctx *ctx, el file.Element) bool {
	var n int

	for _, itm := range el.Body {
		switch itm := itm.(type) {
		case file.InlineText:
			txtLen := onlyTextInTextLines(itm.Text)
			if txtLen < 0 {
				return false
			}
			n += txtLen
		case file.ArrowBlock:
			txtLen := onlyTextInTextLines(itm.Lines...)
			if txtLen < 0 {
				return false
			}
			n += txtLen
		default:
			return false
		}
	}

	var sb strings.Builder
	sb.Grow(n)

	for _, itm := range el.Body {
		switch itm := itm.(type) {
		case file.InlineText:
			writeTextLines(&sb, nil, itm.Text)
		case file.ArrowBlock:
			writeTextLines(&sb, nil, itm.Lines...)
		default:
			return false
		}
	}

	s := styleBodyTextEscaper.f(sb.String())
	css, err := mini.String("text/css", s)
	if err != nil {
		panic(fmt.Errorf("%s:%d:%d: style contains invalid CSS: %w", ctx.currentFile().Name, el.Line, el.Col, err))
	}

	ctx.closeStartTag()
	ctx.generate(css, nil)
	return true
}

func onlyTextInTextLines(lns ...file.TextLine) int {
	var n int

	var prevLnNo int

	for _, ln := range lns {
		if prevLnNo > 0 {
			n += ln.Pos().Line - prevLnNo
		}

		for _, itm := range ln {
			switch itm := itm.(type) {
			case file.Text:
				n += len(itm.Text)
			case file.SimpleInterpolation:
				txt, ok := itm.Value.(file.TextInterpolationValue)
				if ok {
					n += len(txt.Text)
				} else {
					return -1
				}
			default:
				return -1
			}
		}
	}

	return n
}

const jsExprPlaceholder = "__corgi_expr"

func minifyScriptElement(ctx *ctx, el file.Element) bool {
	var n int

	for _, itm := range el.Body {
		switch itm := itm.(type) {
		case file.InlineText:
			txtLen := onlyExprsInTextLines(itm.Text)
			if txtLen < 0 {
				return false
			}
			n += txtLen
		case file.ArrowBlock:
			txtLen := onlyExprsInTextLines(itm.Lines...)
			if txtLen < 0 {
				return false
			}
			n += txtLen
		default:
			return false
		}
	}

	var sb strings.Builder
	sb.Grow(n)

	placeholderExprs := make([]file.Expression, 0, n)

	for _, itm := range el.Body {
		switch itm := itm.(type) {
		case file.InlineText:
			writeTextLines(&sb, &placeholderExprs, itm.Text)
		case file.ArrowBlock:
			writeTextLines(&sb, &placeholderExprs, itm.Lines...)
		default:
			return false
		}
	}

	s := scriptBodyTextEscaper.f(sb.String())
	js, err := mini.String("application/javascript", s)
	if err != nil {
		return false
	}

	ctx.closeStartTag()

	for i, expr := range placeholderExprs {
		ph := jsExprPlaceholder + strconv.Itoa(i)
		placeholderIndex := strings.Index(js, ph)
		if placeholderIndex < 0 {
			return false
		}

		ctx.generate(js[:placeholderIndex], nil)
		generateExpression(ctx, expr, nil, &scriptBodyExprEscaper, nil)
		js = js[placeholderIndex+len(ph):]
	}

	ctx.generate(js, nil)

	return true
}

func onlyExprsInTextLines(lns ...file.TextLine) int {
	var n int

	var prevLnNo int

	for _, ln := range lns {
		if prevLnNo > 0 {
			n += ln.Pos().Line - prevLnNo
		}

		for _, itm := range ln {
			switch itm := itm.(type) {
			case file.Text:
				n += len(itm.Text)
			case file.SimpleInterpolation:
				txt, ok := itm.Value.(file.TextInterpolationValue)
				if ok {
					n += len(txt.Text)
				} else {
					n += len(jsExprPlaceholder) + len("12")
				}
			default:
				return -1
			}
		}
	}

	return n
}

func writeTextLines(sb *strings.Builder, placeholderExprs *[]file.Expression, lns ...file.TextLine) {
	var prevLnNo int
	for _, ln := range lns {
		if prevLnNo > 0 {
			sb.WriteString(strings.Repeat("\n", ln.Pos().Line-prevLnNo))
		}

		for _, itm := range ln {
			switch itm := itm.(type) {
			case file.Text:
				sb.WriteString(hashUnescaper.Replace(itm.Text))
			case file.SimpleInterpolation:
				txt, ok := itm.Value.(file.TextInterpolationValue)
				if ok {
					sb.WriteString(txt.Text)
					continue
				}

				if placeholderExprs == nil {
					continue
				}

				expr, ok := itm.Value.(file.ExpressionInterpolationValue)
				if ok {
					sb.WriteString(jsExprPlaceholder)
					sb.WriteString(strconv.Itoa(len(*placeholderExprs)))
					*placeholderExprs = append(*placeholderExprs, expr.Expression)
				}
			}
		}
		prevLnNo = ln.Pos().Line
	}
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
		ctx.youShouldntSeeThisError(fmt.Errorf("unrecognized attribute collection %T", acoll))
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
		ctx.youShouldntSeeThisError(fmt.Errorf("unrecognized attribute %T", attr))
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
					s := unquoteStringExpressionText(sexpr, txt)
					s = strings.ReplaceAll(s, "##", "#")
					ctx.generateStringAttr(sattr.Name, s)
					return
				}
			}

			if attrType == woof.ContentTypePlain {
				generateExpression(ctx, *sattr.Value, &attrTextEscaper, &plainAttrExprEscaper, func(f func()) {
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
					ctx.flushGenerate()
					ctx.writeln(ctx.woofFunc("WriteAttr", ctx.ident(ctxVar), strconv.Quote(sattr.Name), expr,
						ctx.woofQual("EscapeHTMLAttrVal")))
				})
				return
			}
		}
	}

	switch attrType {
	case woof.ContentTypePlain:
		expr := inlineExpression(ctx, *sattr.Value)
		ctx.flushGenerate()
		ctx.writeln(ctx.woofFunc("WriteAttr", ctx.ident(ctxVar), strconv.Quote(sattr.Name), expr,
			ctx.woofQual("EscapeHTMLAttrVal")))
	case woof.ContentTypeCSS:
		generateExpression(ctx, *sattr.Value, &attrTextEscaper, &cssAttrExprEscaper, func(f func()) {
			ctx.generate(` `+sattr.Name+`="`, nil)
			f()
			ctx.generate(`"`, nil)
		})
	case woof.ContentTypeJS:
		generateExpression(ctx, *sattr.Value, &attrTextEscaper, &jsAttrExprEscaper, func(f func()) {
			ctx.generate(` `+sattr.Name+`="`, nil)
			f()
			ctx.generate(`"`, nil)
		})
	case woof.ContentTypeHTML:
		generateExpression(ctx, *sattr.Value, &attrTextEscaper, &htmlAttrExprEscaper, func(f func()) {
			ctx.generate(` `+sattr.Name+`="`, nil)
			f()
			ctx.generate(`"`, nil)
		})
	case woof.ContentTypeURL:
		generateContextExpression(ctx, *sattr.Value, urlAttrExprEscaper, func(f func()) {
			ctx.generate(` `+sattr.Name+`="`, nil)
			f()
			ctx.generate(`"`, nil)
		})
	case woof.ContentTypeSrcset:
		generateContextExpression(ctx, *sattr.Value, srcsetAttrExprEscaper, func(f func()) {
			ctx.generate(` `+sattr.Name+`="`, nil)
			f()
			ctx.generate(`"`, nil)
		})
	default:
		ctx.youShouldntSeeThisError(fmt.Errorf("unrecognized content type %v", attrType))
	}
}

func classAttribute(ctx *ctx, attr file.SimpleAttribute) {
	ctx.debugItem(attr, "(identified as class attr)")

	if attr.Value == nil || len(attr.Value.Expressions) == 0 {
		return
	}

	if len(attr.Value.Expressions) == 1 {
		switch exprItm := attr.Value.Expressions[0].(type) {
		case file.StringExpression:
			if len(exprItm.Contents) == 0 {
				return
			} else if len(exprItm.Contents) == 1 {
				txt, ok := exprItm.Contents[0].(file.StringExpressionText)
				if ok {
					s := unquoteStringExpressionText(exprItm, txt)
					s = strings.ReplaceAll(s, "##", "#")
					ctx.bufClass(s)
					return
				}
			}
		case file.ChainExpression:
			ctx.flushClasses()
			valueChainExpression(ctx, exprItm, func(expr string) {
				ctx.writeln(ctx.contextFunc("BufferClass", expr))
			})
			return
		}
	}

	ctx.flushClasses()
	ctx.writeln(ctx.contextFunc("BufferClass", inlineExpression(ctx, *attr.Value)))
}

// =================================== AndPlaceholder ===================================

const andPlaceholderFunc = "andPlaceholder"

func andPlaceholder(ctx *ctx, aph file.AndPlaceholder) {
	ctx.debugItem(aph, "(see below)")
	ctx.flushGenerate()
	ctx.flushClasses()
	ctx.callUnclosedIfUnclosed()

	ctx.writeln("if " + ctx.ident(andPlaceholderFunc) + " != nil {")
	ctx.writeln("  " + ctx.ident(andPlaceholderFunc) + "()")
	ctx.writeln("}")

	// force call to CloseStartTag to flush class buffer
	ctx.scope().haveBufClasses = true
}

// ================================= MixinCallAttribute =================================

func mixinCallAttribute(ctx *ctx, mcAttr file.MixinCallAttribute) {
	ctx.debugItem(mcAttr, "(see below)")

	ctx.generate(mcAttr.Name+`="`, nil)
	defer ctx.generate(`"`, nil)

	nest := ctx.startScope(true)
	defer ctx.endScope()
	nest.txtEscaper = attrTextEscaper
	nest.exprEscaper = plainAttrExprEscaper

	ctx.writeln(ctx.contextFunc("StartAttribute"))
	interpolationValueMixinCall(ctx, mcAttr.MixinCall, mcAttr.Value)
	ctx.writeln(ctx.contextFunc("EndAttribute"))
}
