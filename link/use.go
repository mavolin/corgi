package link

import (
	"errors"
	"path"

	"github.com/mavolin/corgi/corgierr"
	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/fileutil"
	"github.com/mavolin/corgi/internal/anno"
	"github.com/mavolin/corgi/internal/list"
)

func (l *Linker) linkDependencies(ctx *context, lib *file.Library) {
	ctx.n += len(lib.Dependencies)
	for i := range lib.Dependencies {
		dep := &lib.Dependencies[i]
		go func() {
			var usingFile *file.File

			reqName := dep.Mixins[0].RequiredBy[0]
			for _, m := range lib.Mixins {
				if m.Mixin.Name.Ident == reqName {
					usingFile = m.File
					break
				}
			}

			var err error
			dep.Library, err = l.loader.LoadLibrary(usingFile, path.Join(dep.Module, dep.PathInModule))
			if err != nil {
				ctx.errs <- list.List1(&corgierr.Error{
					Message: "failed to load dependency of precompiled library",
					ErrorAnnotation: corgierr.Annotation{
						File:         usingFile,
						ContextStart: 1,
						Line:         1,
						ContextEnd:   2,
						Start:        1,
						End:          2,
						Annotation: "no position;\n" +
							path.Join(lib.Module, lib.PathInModule) + " requires " + path.Join(dep.Module, dep.PathInModule),
						Lines: []string{""},
					},
					Suggestions: []corgierr.Suggestion{
						{
							Suggestion: "this is probably a module that has become unavailable;\n" +
								"re-recompiling the library will give you a more exact error message",
						},
					},
					Cause: err,
				})
				return
			}

			if len(dep.Library.Files) == 0 {
				ctx.errs <- list.List1(&corgierr.Error{
					Message: "dependency of precompiled library contains no library files",
					ErrorAnnotation: corgierr.Annotation{
						File:         usingFile,
						ContextStart: 1,
						Line:         1,
						ContextEnd:   2,
						Start:        1,
						End:          2,
						Annotation: "no position;\n" +
							path.Join(lib.Module, lib.PathInModule) + " requires " + path.Join(dep.Module, dep.PathInModule),
						Lines: []string{""},
					},
					Suggestions: []corgierr.Suggestion{
						{
							Suggestion: "the dependency has probably been changed since the library was precompiled;\n" +
								"you should re-precompile",
						},
					},
					Cause: err,
				})
				return
			}

			ctx.errs <- l.linkMixinDependencies(lib, dep)
		}()
	}
}

func (l *Linker) linkMixinDependencies(lib *file.Library, libDep *file.LibDependency) *errList {
	var errs errList

mixins:
	for i, a := range libDep.Mixins {
		if libDep.Library.Precompiled {
			for _, b := range libDep.Library.Mixins {
				b := b
				if a.Name == b.Mixin.Name.Ident {
					libDep.Mixins[i].Mixin = &b.Mixin
					continue mixins
				}
			}

			errs.PushBack(&corgierr.Error{
				Message: "failed to link mixin dependency of precompiled library",
				ErrorAnnotation: corgierr.Annotation{
					File:         lib.Files[0],
					ContextStart: 1,
					Line:         1,
					ContextEnd:   2,
					Start:        1,
					End:          2,
					Annotation: "no position;\n" +
						path.Join(lib.Module, lib.PathInModule) + " requires " + path.Join(libDep.Library.Module, libDep.Library.PathInModule) + "." + a.Name + " cannot be found",
					Lines: []string{""},
				},
				Suggestions: []corgierr.Suggestion{
					{Suggestion: "this is likely because of changes in a module and can be resolved by re-precompiling"},
				},
			})
			continue
		}

		for _, f := range libDep.Library.Files {
			for j, itm := range f.Scope {
				b, ok := itm.(file.Mixin)
				if ok && a.Name == b.Name.Ident {
					libDep.Mixins[i].Mixin = ptrOfSliceElem[file.ScopeItem, file.Mixin](f.Scope, j)
					continue mixins
				}
			}
		}

		errs.PushBack(&corgierr.Error{
			Message: "failed to link mixin dependency of precompiled library",
			ErrorAnnotation: corgierr.Annotation{
				File:         lib.Files[0],
				ContextStart: 1,
				Line:         1,
				ContextEnd:   2,
				Start:        1,
				End:          2,
				Annotation: "no position;\n" +
					path.Join(lib.Module, lib.PathInModule) + " requires " + path.Join(libDep.Library.Module, libDep.Library.PathInModule) + "." + a.Name + " cannot be found",
				Lines: []string{""},
			},
			Suggestions: []corgierr.Suggestion{
				{Suggestion: "this is likely because of changes in a module and can be resolved by re-precompiling"},
			},
		})
	}

	return &errs
}

func (l *Linker) linkUses(ctx *context, f *file.File) {
	for _, use := range f.Uses {
		use := use
		ctx.n += len(use.Uses)
		for specI := range use.Uses {
			spec := &use.Uses[specI]
			go func() {
				ctx.errs <- l.linkUseSpec(f, spec)
			}()
		}
	}
}

func (l *Linker) linkUseSpec(f *file.File, spec *file.UseSpec) *errList {
	lib, err := l.loader.LoadLibrary(f, fileutil.Unquote(spec.Path))
	if err != nil {
		var cerr *corgierr.Error
		if errors.As(err, &cerr) {
			return list.List1(cerr)
		}
		var clerr corgierr.List
		if errors.As(err, &clerr) {
			return list.FromSlice(clerr)
		}

		return list.List1(&corgierr.Error{
			Message: "failed to load library",
			ErrorAnnotation: anno.Anno(f, anno.Annotation{
				Start:      spec.Position,
				ToEOL:      true,
				Annotation: err.Error(),
			}),
			Cause: err,
		})
	}

	if lib == nil {
		return list.List1(&corgierr.Error{
			Message: "library not found",
			ErrorAnnotation: anno.Anno(f, anno.Annotation{
				Start:      spec.Position,
				ToEOL:      true,
				Annotation: "found no library with this use path",
			}),
		})
	}

	spec.Library = lib
	return &errList{}
}
