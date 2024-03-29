// Package writer provides a writer that allows converting a file.File to Go
// code.
package write

import (
	"bytes"
	"io"
	"strconv"
	"strings"

	"github.com/mavolin/corgi/corgierr"
	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/fileutil"
	"github.com/mavolin/corgi/file/precomp"
	"github.com/mavolin/corgi/internal/list"
)

type Options struct {
	// IdentPrefix is the prefix corgi puts in front of an identifier.
	//
	// It defaults to "__corgi_".
	IdentPrefix string

	AllowedFilters  []string
	AllowAllFilters bool

	// CLI indicates that this writer is run by the corgi CLI and may reference
	// CLI options in error messages, and print to stderr on its own.
	CLI bool
	// CorgierrPretty are the [corgierr.PrettyOptions] used to print pretty
	// errors.
	//
	// Only used if CLI is true.
	CorgierrPretty corgierr.PrettyOptions

	// Debug, if set to true, attaches file and position information of scope
	// items to the generated file.
	Debug bool
}

type Writer struct {
	o Options
}

func New(o Options) *Writer {
	if o.IdentPrefix == "" {
		o.IdentPrefix = "__corgi_"
	}

	return &Writer{o}
}

func (w *Writer) GenerateFile(out io.Writer, destPackage string, f *file.File) (err error) {
	// judge me all you want, no function deserves to have an error check every
	// two lines, so write and co. panic
	defer func() {
		if rec := recover(); rec != nil {
			var ok bool
			//goland:noinspection GoTypeAssertionOnErrors
			err, ok = rec.(error)
			if !ok || strings.HasPrefix(err.Error(), "runtime error:") {
				panic(rec)
			}
		}
	}()

	ctx := newCtx(w.o)
	ctx.out = out
	ctx.destPackage = destPackage

	var n int
	for f := f; f != nil; {
		n++
		if f.Extend == nil {
			break
		}

		f = f.Extend.File
	}
	ctx._stack = make([]*file.File, n)
	for i := n - 1; i >= 0; i-- {
		ctx._stack[i] = f
		if f.Extend != nil {
			f = f.Extend.File
		}
	}

	writePackage(ctx)
	writeImports(ctx)
	writeCodegenComment(ctx)
	writeGlobalCode(ctx)
	writeFunc(ctx)

	return err
}

func (w *Writer) PrecompileLibrary(out io.Writer, lib *file.Library) (err error) {
	// judge me all you want, no function deserves to have an error check every
	// two lines, so write and co. panic
	defer func() {
		if rec := recover(); rec != nil {
			var ok bool
			//goland:noinspection GoTypeAssertionOnErrors
			err, ok = rec.(error)
			if !ok || strings.HasPrefix(err.Error(), "runtime error:") {
				panic(rec)
			}
		}
	}()

	lib.Precompiled = true

	deps := fileutil.LibraryDependencies(lib)

	var mixinNames int

	for _, f := range lib.Files {
		fileutil.Walk(f.Scope, func(parents []fileutil.WalkContext, ctx fileutil.WalkContext) (dive bool, err error) {
			switch itm := (*ctx.Item).(type) {
			case file.Code:
				mComs := make([]string, 0, len(ctx.Comments))

				for _, com := range ctx.Comments {
					if fileutil.ParseMachineComment(com) != nil {
						mComs = append(mComs, com.Lines[0].Comment)
					}
				}

				lns := make([]string, len(itm.Lines))
				for i, ln := range itm.Lines {
					lns[i] = ln.Code
				}

				lib.GlobalCode = append(lib.GlobalCode, file.PrecompiledCode{
					MachineComments: mComs,
					Lines:           lns,
				})
			case file.Mixin:
				mComs := make([]string, 0, len(ctx.Comments))

				for _, com := range ctx.Comments {
					if fileutil.ParseMachineComment(com) != nil {
						mComs = append(mComs, com.Lines[0].Comment)
					}
				}

				var requiredBy []string
				for _, mDep := range deps.Self {
					if mDep.Mixin.Name.Ident == itm.Name.Ident {
						requiredBy = mDep.RequiredBy
						break
					}
				}

				lib.Mixins = append(lib.Mixins, file.PrecompiledMixin{
					File:            f,
					MachineComments: mComs,
					Mixin:           itm,
					Var:             w.o.IdentPrefix + "preMixin" + strconv.Itoa(mixinNames),
					RequiredBy:      requiredBy,
				})

				mixinNames++
			}

			return false, nil
		})
	}

	mixinFuncNames := mixinFuncMap{
		m:     make(map[string]map[string]string),
		scope: make(map[*file.File]*list.List[map[string]string]),
	}

	selfMixins := make(map[string]string, len(deps.Self))
	for _, m := range lib.Mixins {
		selfMixins[m.Mixin.Name.Ident] = m.Var
	}
	mixinFuncNames.m[lib.Module+"/"+lib.PathInModule] = selfMixins

	lib.Dependencies = make([]file.LibDependency, len(deps.External))
	for i, ulib := range deps.External {
		moduleMixins := make(map[string]string, len(ulib.Mixins))

		libDep := file.LibDependency{
			Module:       ulib.Library.Module,
			PathInModule: ulib.Library.PathInModule,
			Library:      ulib.Library,
			Mixins:       make([]file.MixinDependency, len(ulib.Mixins)),
		}

		for i, um := range ulib.Mixins {
			varName := w.o.IdentPrefix + "preMixin" + strconv.Itoa(mixinNames)
			moduleMixins[um.Mixin.Name.Ident] = varName
			libDep.Mixins[i] = file.MixinDependency{
				Name:       um.Mixin.Name.Ident,
				Var:        varName,
				Mixin:      um.Mixin,
				RequiredBy: um.RequiredBy,
			}
			mixinNames++
		}
		lib.Dependencies[i] = libDep

		mixinFuncNames.m[ulib.Library.Module+"/"+ulib.Library.PathInModule] = moduleMixins
	}

	for i, pm := range lib.Mixins {
		pm := pm
		var buf bytes.Buffer

		ctx := newCtx(w.o)
		ctx.out = &buf
		ctx._stack = []*file.File{pm.File}
		ctx.mixin = &pm.Mixin
		ctx.mixinFuncNames = mixinFuncNames
		ctx.hasNonce = true

		writeMixinFunc(ctx, &pm.Mixin)

		lib.Mixins[i].Mixin.Precompiled = buf.Bytes()
	}

	if err := precomp.Encode(out, lib); err != nil {
		return err
	}

	return err
}
