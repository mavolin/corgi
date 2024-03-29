package write

import (
	"fmt"
	"path"
	"strings"

	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/fileutil"
)

// ============================================================================
// Mixin
// ======================================================================================

func writeLibMixins(ctx *ctx) {
	ums := fileutil.ListUsedMixins(ctx.mainFile())

	writeMixinVars(ctx, ums)

	ctx.mixinCounter = 0

	for _, src := range ums.External {
		writeLibrary(ctx, src)
	}
}

func writeMixinVars(ctx *ctx, ums fileutil.UsedMixins) {
	for _, ulib := range ums.External {
		moduleMixins := make(map[string]string, len(ulib.Mixins))

		for _, um := range ulib.Mixins {
			varName := ctx.nextMixinIdent()
			moduleMixins[um.Mixin.Name.Ident] = varName

			ctx.write("var " + varName + " ")
			writeMixinSignature(ctx, um.Mixin)
			if ctx.debugEnabled {
				ctx.writeln(" // " + path.Join(ulib.Library.Module,
					ulib.Library.PathInModule) + "." + um.Mixin.Name.Ident)
			} else {
				ctx.writeln("")
			}
		}

		ctx.mixinFuncNames.m[ulib.Library.Module+"/"+ulib.Library.PathInModule] = moduleMixins
	}
}

func writeLibrary(ctx *ctx, ulib fileutil.UsedLibrary) {
	ctx.writeln("{")
	defer ctx.writeln("}")
	ctx.debug("library", ulib.Library.Module+"/"+ulib.Library.PathInModule)
	if ulib.Library.Precompiled {
		ctx.debug("library", "precompiled")
	}

	for _, c := range ulib.Library.GlobalCode {
		allowed := -1

		for _, mcom := range c.MachineComments {
			mcom := fileutil.ParseMachineCommentLine(mcom)
			if mcom.Namespace == "corgi" && mcom.Directive == "formixin" {
				if allowed < 0 {
					allowed = 0
				}

				for _, a := range strings.Split(mcom.Args, " ") {
					for _, b := range ulib.Mixins {
						if a == b.Mixin.Name.Ident {
							allowed = 1
						}
					}
				}
			}
		}

		if allowed == 0 {
			continue
		}

		for _, ln := range c.Lines {
			ctx.writeln(ln)
		}
	}
	for _, f := range ulib.Library.Files {
		fileutil.Walk(f.Scope,
			func(parents []fileutil.WalkContext, wctx fileutil.WalkContext) (dive bool, err error) {
				c, ok := (*wctx.Item).(file.Code)
				if !ok {
					return false, nil
				}

				allowed := -1

				for _, mcom := range wctx.Comments {
					mcom := fileutil.ParseMachineComment(mcom)
					if mcom == nil {
						continue
					}
					if mcom.Namespace == "corgi" && mcom.Directive == "formixin" {
						if allowed < 0 {
							allowed = 0
						}

						for _, a := range strings.Split(mcom.Args, " ") {
							for _, b := range ulib.Mixins {
								if a == b.Mixin.Name.Ident {
									allowed = 1
								}
							}
						}
					}
				}

				if allowed == 0 {
					return false, nil
				}

				ctx.debugItem(c, "(see below)")
				for _, ln := range c.Lines {
					ctx.writeln(ln.Code)
				}

				return false, nil
			})
	}

	if !ulib.Library.Precompiled {
		for _, m := range ulib.Mixins {
			ctx.write(ctx.nextMixinIdent() + " = ")
			writeMixinFunc(ctx, m.Mixin)
		}

		return
	}

	for _, a := range ulib.Mixins {
		for _, b := range ulib.Library.Mixins {
			if a.Mixin.Name.Ident == b.Mixin.Name.Ident {
				ctx.write("var " + b.Var + " ")
				writeMixinSignature(ctx, a.Mixin)
				if ctx.debugEnabled {
					ctx.writeln(" // " + path.Join(ulib.Library.Module, ulib.Library.PathInModule) +
						"." + a.Mixin.Name.Ident)
				} else {
					ctx.writeln("")
				}

				break
			}
		}
	}

	for _, libDep := range ulib.Library.Dependencies {
	mixins:
		for _, mDep := range libDep.Mixins {
			for _, requiredBy := range mDep.RequiredBy {
				for _, um := range ulib.Mixins {
					if um.Mixin.Name.Ident == requiredBy {
						ctx.write(mDep.Var + " := func(")
						for i, param := range mDep.Mixin.Params {
							ctx.write(fmt.Sprint("p", i, " "))
							if param.Default != nil {
								ctx.write("*")
							}

							if param.Type != nil {
								ctx.write(param.Type.Type)
							} else {
								ctx.write(param.InferredType)
							}
							ctx.write(", ")
						}
						for i := range mDep.Mixin.Blocks {
							ctx.write(fmt.Sprint("b", i, " func(), "))
						}
						if mDep.Mixin.HasAndPlaceholders {
							ctx.write("ab func()")
						}
						ctx.write(") { ")
						ctx.write(ctx.mixinFuncNames.mixinByName(ctx, libDep.Module, libDep.PathInModule, mDep.Name))
						ctx.write("(")
						for i := range mDep.Mixin.Params {
							ctx.write(fmt.Sprint("p", i, ", "))
						}
						for i := range mDep.Mixin.Blocks {
							ctx.write(fmt.Sprint("b", i, ", "))
						}
						if mDep.Mixin.HasAndPlaceholders {
							ctx.write("ab")
						}
						ctx.writeln(") }")
						continue mixins
					}
				}
			}
		}
	}

	for _, a := range ulib.Mixins {
		for _, b := range ulib.Library.Mixins {
			if a.Mixin.Name.Ident == b.Mixin.Name.Ident {
				ctx.write(b.Var + " = ")
				ctx.writeBytes(a.Mixin.Precompiled)
				ctx.writeln("")
				break
			}
		}
	}

	for _, a := range ulib.Mixins {
		for _, b := range ulib.Library.Mixins {
			if a.Mixin.Name.Ident == b.Mixin.Name.Ident {
				varName := ctx.nextMixinIdent()
				ctx.writeln(varName + " = " + b.Var)
				ctx.writeln("_ = " + varName + " // in case this is only a dependency of another mixin in this lib")
				break
			}
		}
	}
}

func writeMixinSignature(ctx *ctx, m *file.Mixin) {
	ctx.write("func(")
	for _, param := range m.Params {
		if param.Default != nil {
			ctx.write("*")
		}

		if param.Type != nil {
			ctx.write(param.Type.Type)
		} else {
			ctx.write(param.InferredType)
		}
		ctx.write(", ")
	}
	for range m.Blocks {
		ctx.write("func(), ")
	}
	if m.HasAndPlaceholders {
		ctx.write("func()")
	}
	ctx.write(")")
}

func scopeMixin(ctx *ctx, m file.Mixin) {
	ctx.flushGenerate()
	ctx.flushClasses()

	name := ctx.mixinFuncNames.addScope(ctx, &m)
	ctx.write(name + " := ")
	writeMixinFunc(ctx, &m)
}

func writeMixinFunc(ctx *ctx, m *file.Mixin) {
	ctx.flushGenerate()
	ctx.flushClasses()

	if len(m.Precompiled) > 0 {
		ctx.writeln(string(m.Precompiled))
		return
	}

	textOnlyMixin := true
	fileutil.Walk(m.Body, func(parents []fileutil.WalkContext, ctx fileutil.WalkContext) (dive bool, err error) {
		switch (*ctx.Item).(type) {
		case file.Element:
			textOnlyMixin = false
			return false, fileutil.StopWalk
		case file.DivShorthand:
			textOnlyMixin = false
			return false, fileutil.StopWalk
		case file.And:
			textOnlyMixin = false
			return false, fileutil.StopWalk
		case file.Mixin:
			return false, nil
		}

		return true, nil
	})

	ctx.write("func(")
	for _, param := range m.Params {
		if param.Default != nil {
			ctx.write(ctx.ident("mixinParam_" + param.Name.Ident))
		} else {
			ctx.write(param.Name.Ident)
		}

		ctx.write(" ")

		if param.Default != nil {
			ctx.write("*")
		}

		if param.Type != nil {
			ctx.write(param.Type.Type)
		} else {
			ctx.write(param.InferredType)
		}
		ctx.write(", ")
	}
	for _, b := range m.Blocks {
		ctx.write(ctx.ident("mixinBlock_"+b.Name) + " func(), ")
	}
	if m.HasAndPlaceholders {
		ctx.write(ctx.ident(andPlaceholderFunc) + " func()")
	}
	ctx.writeln(") {")

	for _, param := range m.Params {
		if param.Default == nil {
			continue
		}

		ctx.debugItem(param, param.Name.Ident)
		val := ctx.ident("mixinParam_" + param.Name.Ident)

		var defaultVal string

		var typ string
		if param.Type != nil {
			typ = param.Type.Type
		} else {
			typ = param.InferredType
		}
		switch woofType(ctx, typ) {
		case "URL":
			defaultVal = escapedInlineContextExpression(ctx, *param.Default, urlAttrExprEscaper)
		case "Srcset":
			defaultVal = escapedInlineContextExpression(ctx, *param.Default, srcsetAttrExprEscaper)
		default:
			defaultVal = inlineExpression(ctx, *param.Default)
		}

		ctx.writeln(param.Name.Ident + " := " + ctx.woofFunc("ResolveDefault", val, defaultVal))
	}

	ctx.startScope(false).startClosed = maybeClosed
	// mixin could be used as attr value, so also escape attr values
	if textOnlyMixin {
		ctx.scope().txtEscaper = htmlTextEscaper
		ctx.scope().exprEscaper = htmlExprEscaper
	}
	defer ctx.endScope()

	ctx.mixin = m
	scope(ctx, m.Body, false)
	ctx.mixin = nil

	ctx.flushGenerate()
	ctx.flushClasses()
	ctx.callClosedIfClosed()
	ctx.writeln("}")
}

// ============================================================================
// Mixin Call
// ======================================================================================

func mixinCall(ctx *ctx, mc file.MixinCall) {
	if mc.Namespace != nil {
		ctx.debugItem(mc, mc.Namespace.Ident+"."+mc.Name.Ident)
	} else {
		ctx.debugItem(mc, mc.Name.Ident)
	}

	funcName := ctx.mixinFuncNames.mixin(ctx, mc)

	ctx.flushGenerate()
	ctx.flushClasses()
	ctx.callUnclosedIfUnclosed()

	ctx.write(funcName + "(")

	ctx.startScope(false)

params:
	for _, param := range mc.Mixin.Mixin.Params {
		for _, arg := range mc.Args {
			if arg.Name.Ident == param.Name.Ident {
				ctx.write(mixinArgExpression(ctx, param, arg))
				ctx.write(", ")
				continue params
			}
		}

		ctx.write("nil, ")
	}

blocks:
	for _, placeholder := range mc.Mixin.Mixin.Blocks {
		if placeholder.Name == "_" {
			if len(mc.Body) == 1 {
				switch itm := mc.Body[0].(type) {
				case file.InlineText:
					ctx.writeln("func() {")

					ctx.startScope(false).startClosed = maybeClosed
					inlineText(ctx, itm)
					ctx.flushGenerate()
					ctx.flushClasses()
					ctx.callClosedIfClosed()
					ctx.endScope()
					ctx.write("}, ")
					continue
				case file.BlockExpansion:
					ctx.writeln("func() {")

					ctx.startScope(false).startClosed = maybeClosed
					blockExpansion(ctx, itm)
					ctx.flushGenerate()
					ctx.flushClasses()
					ctx.callClosedIfClosed()
					ctx.write("}, ")
					ctx.endScope()
					continue
				case file.MixinMainBlockShorthand:
					ctx.writeln("func() {")

					ctx.startScope(false).startClosed = maybeClosed
					scope(ctx, itm.Body, false)
					ctx.flushGenerate()
					ctx.flushClasses()
					ctx.callClosedIfClosed()
					ctx.write("}, ")
					ctx.endScope()
					continue
				}
			}
		}

		for _, itm := range mc.Body {
			b, ok := itm.(file.Block)
			if !ok {
				continue
			}

			if b.Name.Ident == placeholder.Name {
				ctx.writeln("func() {")

				ctx.startScope(false).startClosed = maybeClosed
				scope(ctx, b.Body, false)

				ctx.flushGenerate()
				ctx.flushClasses()
				ctx.callClosedIfClosed()
				ctx.write("}, ")
				ctx.endScope()
				continue blocks
			}
		}

		ctx.write("nil, ")
	}

	if mc.Mixin.Mixin.HasAndPlaceholders {
		ctx.writeln("func() {")

		fileutil.Walk(mc.Body,
			func(parents []fileutil.WalkContext, wctx fileutil.WalkContext) (dive bool, err error) {
				switch itm := (*wctx.Item).(type) {
				case file.Block:
					return false, nil
				case file.InlineText:
					return false, nil
				case file.BlockExpansion:
					return false, nil
				case file.MixinMainBlockShorthand:
					return false, nil
				default:
					scopeItem(ctx, itm)
					return false, nil
				}
			})

		ctx.flushGenerate()
		ctx.flushClasses()

		ctx.write("}")
	}

	ctx.writeln(")")

	ctx.endScope()
	ctx.scope().startClosed = maybeClosed
}

func interpolationValueMixinCall(ctx *ctx, mc file.MixinCall, val file.InterpolationValue) {
	if mc.Namespace != nil {
		ctx.debugItem(mc, mc.Namespace.Ident+"."+mc.Name.Ident)
	} else {
		ctx.debugItem(mc, mc.Name.Ident)
	}

	funcName := ctx.mixinFuncNames.mixin(ctx, mc)

	ctx.flushGenerate()
	ctx.flushClasses()
	ctx.callUnclosedIfUnclosed()

	ctx.write(funcName + "(")

params:
	for _, param := range mc.Mixin.Mixin.Params {
		for _, arg := range mc.Args {
			if arg.Name.Ident == param.Name.Ident {
				ctx.write(mixinArgExpression(ctx, param, arg))
				ctx.write(", ")
				continue params
			}
		}

		ctx.write("nil, ")
	}

	for _, placeholder := range mc.Mixin.Mixin.Blocks {
		if placeholder.Name != "_" || val == nil {
			ctx.write("nil, ")
		}

		ctx.writeln("func() {")

		ctx.startScope(false).startClosed = maybeClosed
		ctx.closeStartTag()
		interpolationValue(ctx, val, false)
		ctx.flushGenerate()
		ctx.flushClasses()
		ctx.callClosedIfClosed()
		ctx.write("}, ")
		ctx.endScope()
	}

	if mc.Mixin.Mixin.HasAndPlaceholders {
		ctx.writeln("func() {")

		fileutil.Walk(mc.Body,
			func(parents []fileutil.WalkContext, wctx fileutil.WalkContext) (dive bool, err error) {
				switch itm := (*wctx.Item).(type) {
				case file.Block:
					return false, nil
				case file.InlineText:
					return false, nil
				case file.BlockExpansion:
					return false, nil
				default:
					scopeItem(ctx, itm)
					return false, nil
				}
			})

		ctx.flushGenerate()
		ctx.flushClasses()

		ctx.write("}")
	}

	ctx.writeln(")")

	ctx.scope().startClosed = maybeClosed
}

// ============================================================================
// Return
// ======================================================================================

func _return(ctx *ctx, ret file.Return) {
	ctx.flushGenerate()
	ctx.flushClasses()
	ctx.callClosedIfClosed()

	if ctx.mixin != nil {
		if ret.Err != nil {
			ctx.writeln(ctx.contextFunc("Panic", inlineExpression(ctx, *ret.Err)))
			return
		}

		ctx.writeln("return")
		return
	}

	if ret.Err != nil {
		ctx.writeln("return " + inlineExpression(ctx, *ret.Err))
		return
	}

	ctx.writeln("return nil")
}
