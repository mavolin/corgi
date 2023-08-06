package write

import (
	"fmt"
	"strconv"

	"github.com/mavolin/corgi/file"
)

// ============================================================================
// InlineText
// ======================================================================================

func inlineText(ctx *ctx, txt file.InlineText) {
	ctx.debugItem(txt, "(see below)")

	ctx.closeTag()
	textLines(ctx, txt.Text)
}

// ============================================================================
// ArrowBlock
// ======================================================================================

func arrowBlock(ctx *ctx, ab file.ArrowBlock) {
	ctx.debugItem(ab, "(see below)")

	ctx.closeTag()

	if len(ab.Lines) == 0 { // special case
		ctx.generate("\n", nil)
		return
	}

	textLines(ctx, ab.Lines...)
}

// ============================================================================
// TextItem
// ======================================================================================

func textLines(ctx *ctx, lns ...file.TextLine) {
	for i, ln := range lns {
		if i > 0 {
			ctx.generate("\n", nil)
		}

		for _, txtItm := range ln {
			switch txtItm := txtItm.(type) {
			case file.Text:
				ctx.generate(txtItm.Text, ctx.txtEscaper.Peek())
			case file.Interpolation:
				interpolation(ctx, txtItm)
			}
		}
	}
}

// ============================================================================
// Interpolation
// ======================================================================================

func interpolation(ctx *ctx, interp file.Interpolation) {
	switch interp := interp.(type) {
	case file.SimpleInterpolation:
		simpleInterpolation(ctx, interp)
	case file.ElementInterpolation:
		elementInterpolation(ctx, interp)
	case file.MixinCallInterpolation:
		mixinCallInterpolation(ctx, interp)
	default:
		panic(fmt.Errorf("unrecognized interpolation %T (you shouldn't see this error, please open an issue)", interp))
	}
}

// ================================ SimpleInterpolation =================================

func simpleInterpolation(ctx *ctx, interp file.SimpleInterpolation) {
	ctx.debugItem(interp, "(see interpolation value)")
	interpolationValue(ctx, interp.Value, interp.NoEscape)
}

// ================================ ElementInterpolation ================================

func elementInterpolation(ctx *ctx, interp file.ElementInterpolation) {
	ctx.startElem(interp.Element.Name, interp.Element.Void)

	for _, acoll := range interp.Element.Attributes {
		attributeCollection(ctx, acoll)
	}
	ctx.closeTag()

	interpolationValue(ctx, interp.Value, false)

	ctx.debugItem(interp, "/"+interp.Element.Name)
	ctx.closeElem()
}

// =============================== MixinCallInterpolation ===============================

func mixinCallInterpolation(ctx *ctx, interp file.MixinCallInterpolation) {
	interpolationValueMixinCall(ctx, interp.MixinCall, interp.Value)
}

// ============================================================================
// InterpolationValue
// ======================================================================================

func interpolationValue(ctx *ctx, interp file.InterpolationValue, noEscape bool) {
	switch interp := interp.(type) {
	case file.TextInterpolationValue:
		textInterpolationValue(ctx, interp, noEscape)
	case file.ExpressionInterpolationValue:
		expressionInterpolationValue(ctx, interp, noEscape)
	default:
		panic(fmt.Errorf("unrecognized interpolation value %T (you shouldn't see this error, please open an issue)", interp))
	}
}

func textInterpolationValue(ctx *ctx, tinterp file.TextInterpolationValue, noEscape bool) {
	ctx.debugItem(tinterp, tinterp.Text)
	if noEscape {
		ctx.generate(tinterp.Text, nil)
		return
	}

	ctx.generate(tinterp.Text, ctx.txtEscaper.Peek())
}

func expressionInterpolationValue(ctx *ctx, exprInterp file.ExpressionInterpolationValue, noEscape bool) {
	ctx.debugItem(exprInterp, "(see below)")

	var esc *escaper
	if !noEscape {
		esc = ctx.exprEscaper.Peek()
	}

	if exprInterp.FormatDirective == "" {
		ctx.debugItem(exprInterp, "(see sub expressions)")
		ctx.generateExpression(exprInterp.Expression, nil, esc, nil)
		return
	}

	ctx.debugItem(exprInterp, "[%"+exprInterp.FormatDirective+"] (see sub expressions)")
	fmtString := strconv.Quote("%" + exprInterp.FormatDirective)
	ctx.generateExpr(ctx.fmtFunc("Sprintf", fmtString, inlineExpression(ctx, exprInterp.Expression)), esc)
}
