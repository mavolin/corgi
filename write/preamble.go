package write

import (
	"path"

	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/fileutil"
)

func writePackage(ctx *ctx) {
	ctx.writeln("package " + ctx.destPackage)
}

func writeImports(ctx *ctx) {
	found := make(map[string]struct{})
	namespaceUsed := make(map[string]struct{})

	files := make([]*file.File, len(ctx._stack))
	copy(files, ctx._stack)

	ctx.writeln("import (")
	defer ctx.writeln(")")

	writeBaseImports(ctx)

	for _, f := range files {
		if _, ok := found[f.Module+f.PathInModule]; ok {
			continue
		}

		found[f.Module+f.PathInModule] = struct{}{}

		for _, use := range f.Uses {
			for _, useSpec := range use.Uses {
				files = append(files, useSpec.Library.Files...)
			}
		}

		for _, imp := range f.Imports {
			for _, impSpec := range imp.Imports {
				var namespace string
				if impSpec.Alias != nil {
					namespace = impSpec.Alias.Ident
				} else {
					namespace = path.Base(fileutil.Unquote(impSpec.Path))
				}

				if _, ok := namespaceUsed[namespace]; ok {
					continue
				}

				namespaceUsed[namespace] = struct{}{}

				if impSpec.Alias != nil {
					ctx.writeln(impSpec.Alias.Ident + " " + fileutil.Quote(impSpec.Path))
				} else {
					ctx.writeln(fileutil.Quote(impSpec.Path))
				}
			}
		}
	}
}

func writeBaseImports(ctx *ctx) {
	ctx.writeln(ctx.ident("io") + ` "io"`)
	ctx.writeln(ctx.ident("woof") + ` "github.com/mavolin/corgi/woof"`)
}

func writeGlobalCode(ctx *ctx) {
	for _, code := range ctx.mainFile().GlobalCode {
		for _, ln := range code.Lines {
			ctx.writeln(ln.Code)
		}
	}
}

func writeFunc(ctx *ctx) {
	ctx.write("func " + ctx.mainFile().Func.Name.Ident + "(" + ctx.ident("w") + " " + ctx.ioQual("Writer"))

	for _, param := range ctx.mainFile().Func.Params {
		ctx.write(", ")

		for nameI, name := range param.Names {
			if nameI > 0 {
				ctx.writeln(", ")
			}
			ctx.write(name.Ident)
		}

		ctx.write(" " + param.Type.Type)
	}

	ctx.writeln(") error {")
	defer ctx.writeln("}")

	ctx.writeln(ctx.ident(ctxVar) + " := " + ctx.woofFunc("NewContext", ctx.ident("w")))
	ctx.writeln("defer " + ctx.contextFunc("Recover"))

	writeMixins(ctx)

	scope(ctx, ctx.baseFile().Scope)
	ctx.flushGenerate()
	ctx.writeln("return " + ctx.contextFunc("Err"))
}
