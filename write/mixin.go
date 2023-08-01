package write

import (
	"path"
	"strconv"

	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/fileutil"
)

// ============================================================================
// Mixin
// ======================================================================================

func writeLibMixins(ctx *ctx) {
	ums := make(usedMixins)

	listLibMixinCalls(ctx, ums, ctx.baseFile().Scope)

	var n int

	mfm := make(mixinFuncMap)
	for p, src := range ums {
		moduleMixins := make(map[string]string)

		for name, um := range src.mixins {
			m := um.m
			if um.pm != nil {
				m = &um.pm.Mixin
			}

			varName := ctx.ident("mixin" + strconv.Itoa(n))
			moduleMixins[name] = varName

			ctx.write("var " + varName + "func(")
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
			if ctx.debugEnabled {
				ctx.writeln(" // " + path.Join(src.lib.Module, src.lib.PathInModule) + "." + m.Name.Ident)
			} else {
				ctx.writeln("")
			}

			n++
		}

		mfm[p] = moduleMixins
	}

	n = 0

	for _, src := range ums {
		ctx.writeln("{")
		ctx.debug("library", path.Join(src.lib.Module, src.lib.PathInModule))

		if src.lib.Precompiled {
			for _, c := range src.lib.GlobalCode {
				for _, ln := range c.Lines {
					ctx.writeln(ln)
				}
			}
		} else {
			for _, f := range src.lib.Files {
				for _, itm := range f.Scope {
					c, ok := itm.(file.Code)
					if ok {
						code(ctx, c)
					}
				}
			}
		}

		for _, um := range src.mixins {
			m := um.m
			if um.pm != nil {
				m = &um.pm.Mixin
			}

			ctx.writeln(" // " + path.Join(src.lib.Module, src.lib.PathInModule) + "." + m.Name.Ident)
			ctx.write(ctx.ident("mixin" + strconv.Itoa(n) + " = "))
			writeMixinFunc(ctx, m)
			n++
		}

		ctx.writeln("}")
	}
}

type (
	usedMixins  map[string]*mixinSource
	mixinSource struct {
		lib    *file.Library
		mixins map[string]usedMixin
	}
	usedMixin struct {
		m  *file.Mixin // either is set
		pm *file.PrecompiledMixin
	}
)

func (ums usedMixins) insert(ctx *ctx, mc file.MixinCall) {
	if mc.Mixin.File.Type != file.TypeLibraryFile {
		return
	}

	ums.insertMixin(ctx, mc.Mixin.File.Library, mc.Mixin.Mixin)
}

func (ums usedMixins) insertMixin(ctx *ctx, lib *file.Library, m *file.Mixin) {
	src, ok := ums[lib.Module+lib.PathInModule]
	if !ok {
		src = &mixinSource{
			lib:    lib,
			mixins: make(map[string]usedMixin),
		}
		ums[lib.Module+lib.PathInModule] = src
	}

	if _, ok = src.mixins[m.Name.Ident]; ok {
		return
	}

	if m.Precompiled == nil {
		src.mixins[m.Name.Ident] = usedMixin{m: m}
		listLibMixinCalls(ctx, ums, m.Body)
	} else {
		for _, pm := range lib.Mixins {
			if pm.Mixin.Name.Ident == m.Name.Ident {
				src.mixins[m.Name.Ident] = usedMixin{pm: &pm}
			}
		}

		ums.insertDeps(ctx, m, lib)
	}
}

func (ums usedMixins) insertMixinByName(ctx *ctx, lib *file.Library, name string) {
	if lib.Precompiled {
		for _, m := range lib.Mixins {
			if m.Mixin.Name.Ident == name {
				ums.insertMixin(ctx, lib, &m.Mixin)
				return
			}
		}

		return
	}

	for _, f := range lib.Files {
		for _, itm := range f.Scope {
			m, ok := itm.(file.Mixin)
			if !ok {
				continue
			}

			if m.Name.Ident == name {
				ums.insertMixin(ctx, lib, &m)
				return
			}
		}
	}
}

func (ums usedMixins) insertDeps(ctx *ctx, of *file.Mixin, lib *file.Library) {
	for _, dep := range lib.Dependencies {
		for _, m := range dep.Mixins {
			for _, reqBy := range m.RequiredBy {
				if reqBy != of.Name.Ident {
					continue
				}

				ums.insertMixinByName(ctx, dep.Library, m.Name)
			}
		}
	}
}

func listLibMixinCalls(ctx *ctx, ums usedMixins, s file.Scope) {
	fileutil.Walk(s, func(parents []fileutil.WalkContext, wctx fileutil.WalkContext) (dive bool, err error) {
		switch itm := (*wctx.Item).(type) {
		case file.Block:
			for _, parent := range parents {
				switch (*parent.Item).(type) {
				case file.Mixin:
					return true, nil
				case file.MixinCall:
					return true, nil
				}
			}

			b, stackPos := resolveTemplateBlock(ctx, itm)

			oldPos := ctx.stackStart
			ctx.stackStart = stackPos
			listLibMixinCalls(ctx, ums, b.Body)
			ctx.stackStart = oldPos
			return false, nil
		case file.MixinCall:
			ums.insert(ctx, itm)
			return true, nil
		default:
			return true, nil
		}
	})
}

func writeMixinFunc(ctx *ctx, m *file.Mixin) {
	ctx.write("func(")
	for _, param := range m.Params {
		if param.Default != nil {
			ctx.write(ctx.ident("mixinParam_" + param.Name.Ident))
			ctx.write("*")
		} else {
			ctx.writeln(param.Name.Ident)
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

		ctx.writeln("var " + param.Name.Ident + " ")
		if param.Type != nil {
			ctx.writeln(param.Type.Type)
		} else {
			ctx.writeln(param.InferredType)
		}

		ctx.writeln("if " + ctx.ident("mixinParam_"+param.Name.Ident) + " == nil {")
		ctx.writeln("  " + param.Name.Ident + " = " + inlineExpression(ctx, *param.Default))
		ctx.writeln("} else {")
		ctx.writeln("  " + param.Name.Ident + " = *" + ctx.ident("mixinParam_"+param.Name.Ident))
		ctx.writeln("}")
	}

	ctx.mixin = m
	scope(ctx, m.Body)
	ctx.mixin = nil

	ctx.writeln("}")
}

// ============================================================================
// Mixin Call
// ======================================================================================

func mixinCall(ctx *ctx, mc file.MixinCall) {
	funcName := ctx.mixinFuncNames.mixin(mc)

	ctx.write(funcName + "(")

params:
	for _, param := range mc.Mixin.Mixin.Params {
		for _, arg := range mc.Args {
			if arg.Name.Ident == param.Name.Ident {
				inlineExpression(ctx, arg.Value)
				ctx.write(", ")
				continue params
			}
		}

		ctx.write("nil, ")
	}

blocks:
	for _, placeholder := range mc.Mixin.Mixin.Blocks {
		for _, itm := range mc.Body {
			b, ok := itm.(file.Block)
			if !ok {
				continue
			}

			if b.Name.Ident == placeholder.Name {
				ctx.writeln("func() {")
				scope(ctx, b.Body)
				ctx.write("}, ")
				continue blocks
			}
		}

		ctx.write("nil, ")
	}

	// todo and placeholder

	ctx.writeln(")")
}
