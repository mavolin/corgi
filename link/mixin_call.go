package link

import (
	"fmt"
	"path"

	"github.com/mavolin/corgi/corgierr"
	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/fileutil"
	"github.com/mavolin/corgi/internal/anno"
	"github.com/mavolin/corgi/internal/list"
)

func (l *Linker) linkMixinCalls(f *file.File) errList {
	var errs errList

	fileutil.Walk(f.Scope, func(parents []fileutil.WalkContext, ctx fileutil.WalkContext) (dive bool, err error) {
		var mcs []*file.MixinCall
		switch itm := (*ctx.Item).(type) {
		case file.MixinCall:
			mc := ptrOfSliceElem[file.ScopeItem, file.MixinCall](ctx.Scope, ctx.Index)
			mcs = []*file.MixinCall{mc}

		case file.InlineText:
			mcs = mixinCallsInText(itm.Text)
		case file.ArrowBlock:
			mcs = mixinCallsInText(itm.Lines...)

		case file.Element:
			mcs = mixinCallsInAttributeCollections(itm.Attributes)
		case file.DivShorthand:
			mcs = mixinCallsInAttributeCollections(itm.Attributes)
		case file.And:
			mcs = mixinCallsInAttributeCollections(itm.Attributes)
		}

		for _, mc := range mcs {
			mcErrs := l.linkMixinCall(f, parents, ctx, mc)
			errs.PushBackList(&mcErrs)
		}
		return true, err
	})

	return errs
}

func mixinCallsInText(lns ...file.TextLine) []*file.MixinCall {
	var mcs []*file.MixinCall

	for _, ln := range lns {
		for txtItmI, txtItm := range ln {
			_, ok := txtItm.(file.MixinCallInterpolation)
			if ok {
				mcs = append(mcs, &ptrOfSliceElem[file.TextItem, file.MixinCallInterpolation](ln, txtItmI).MixinCall)
			}
		}
	}

	return mcs
}

func mixinCallsInAttributeCollections(acolls []file.AttributeCollection) []*file.MixinCall {
	var mcs []*file.MixinCall

	for _, acoll := range acolls {
		attrList, ok := acoll.(file.AttributeList)
		if !ok {
			continue
		}

		for attrI, attr := range attrList.Attributes {
			_, ok := attr.(file.MixinCallAttribute)
			if ok {
				mcs = append(mcs, &ptrOfSliceElem[file.Attribute, file.MixinCallAttribute](attrList.Attributes, attrI).MixinCall)
			}
		}
	}

	return mcs
}

func (l *Linker) linkMixinCall(f *file.File, parents []fileutil.WalkContext, ctx fileutil.WalkContext, mc *file.MixinCall) errList {
	if mc.Namespace == nil {
		l.linkScopeMixinCall(f, ctx.Scope, mc)
		if mc.Mixin != nil {
			return errList{}
		}

		l.linkParentMixinCall(f, parents, mc)
		if mc.Mixin != nil {
			return errList{}
		}

		if f.DirLibrary != nil {
			l.linkLibraryMixinCall(f.DirLibrary, mc)
			if mc.Mixin != nil {
				return errList{}
			}
		}

		if f.Library != nil {
			l.linkLibraryMixinCall(f.Library, mc)
			if mc.Mixin != nil {
				return errList{}
			}
		}

		hasUnlinkedLibs := l.linkExternalMixinCall(".", f, mc)
		if mc.Mixin != nil {
			return errList{}
		}

		// don't report this mixin as unknown, cause the more likely cause is
		// simply that some use directive needs a typo fixed
		if hasUnlinkedLibs {
			return errList{}
		}

		return list.List1(&corgierr.Error{
			Message: "call to unknown mixin",
			ErrorAnnotation: anno.Anno(f, anno.Annotation{
				Start: mc.Name.Position,
				Len:   len(mc.Name.Ident),
				Annotation: "found no mixin with this name in this or a parent scope,\n" +
					"any library files in this file's directory, or any used libs with a `.` alias",
			}),
		})
	}

	hasUnlinkedLibs := l.linkExternalMixinCall(mc.Namespace.Ident, f, mc)
	if mc.Mixin != nil {
		return errList{}
	}

	if hasUnlinkedLibs {
		return errList{}
	}

	for _, useSpecs := range f.Uses {
		for _, use := range useSpecs.Uses {
			var useNamespace string
			if use.Alias != nil {
				useNamespace = use.Alias.Ident
			} else {
				useNamespace = path.Base(use.Path.Contents)
			}

			if useNamespace != mc.Namespace.Ident {
				continue
			}

			return list.List1(&corgierr.Error{
				Message: "call to unknown mixin",
				ErrorAnnotation: anno.Anno(f, anno.Annotation{
					Start:      mc.Namespace.Position,
					End:        mc.Name.Position,
					EndOffset:  len(mc.Name.Ident),
					Annotation: "found no mixin with this name in the above library",
				}),
				HintAnnotations: []corgierr.Annotation{
					anno.Anno(f, anno.Annotation{
						Start:      use.Position,
						ToEOL:      true,
						Annotation: "this library",
					}),
				},
			})
		}
	}

	return list.List1(&corgierr.Error{
		Message: "missing use for library `" + mc.Namespace.Ident + "`",
		ErrorAnnotation: anno.Anno(f, anno.Annotation{
			Start:      mc.Namespace.Position,
			Len:        len(mc.Namespace.Ident),
			Annotation: "this file imports no library under the namespace `" + mc.Namespace.Ident + "`",
		}),
		Suggestions: []corgierr.Suggestion{
			{Suggestion: "did you forget to add a `use`?"},
			{Suggestion: "did you forget to add a `use` alias?"},
		},
	})
}

func (l *Linker) linkParentMixinCall(f *file.File, parents []fileutil.WalkContext, mc *file.MixinCall) {
	for i := len(parents) - 1; i >= 0; i-- {
		parent := parents[i]

		l.linkScopeMixinCall(f, parent.Scope, mc)
		if mc.Mixin != nil {
			return
		}
	}
}

func (l *Linker) linkLibraryMixinCall(lib *file.Library, mc *file.MixinCall) {
	if lib.Precompiled {
		l.linkPrecompiledMixinsMixinCall(lib.Mixins, mc)
		return
	}

	for _, libFile := range lib.Files {
		l.linkScopeMixinCall(libFile, libFile.Scope, mc)
		if mc.Mixin != nil {
			return
		}
	}
}

func (l *Linker) linkExternalMixinCall(namespace string, f *file.File, mc *file.MixinCall) (unlinkedLibs bool) {
	for _, useSpecs := range f.Uses {
		for _, use := range useSpecs.Uses {
			var useNamespace string
			if use.Alias != nil {
				useNamespace = use.Alias.Ident
			} else {
				useNamespace = path.Base(use.Path.Contents)
			}

			if useNamespace != namespace {
				continue
			}

			if use.Library == nil {
				unlinkedLibs = true
				continue
			}

			l.linkLibraryMixinCall(use.Library, mc)
			if mc.Mixin != nil {
				return
			}
		}
	}

	return unlinkedLibs
}

func (l *Linker) linkScopeMixinCall(f *file.File, s file.Scope, mc *file.MixinCall) {
	for i, itm := range s {
		m, ok := itm.(file.Mixin)
		if !ok {
			continue
		}

		if m.Name.Ident != mc.Name.Ident {
			continue
		}

		mptr := ptrOfSliceElem[file.ScopeItem, file.Mixin](s, i)

		mc.Mixin = &file.LinkedMixin{
			File:  f,
			Mixin: mptr,
		}

		return
	}
}

func (l *Linker) linkPrecompiledMixinsMixinCall(ms []file.PrecompiledMixin, mc *file.MixinCall) {
	for _, m := range ms {
		if m.Mixin.Name.Ident != mc.Name.Ident {
			continue
		}

		mc.Mixin = &file.LinkedMixin{
			File:  m.File,
			Mixin: &m.Mixin, //nolint:exportloopref
		}
		return
	}
}

func (l *Linker) checkRecursion(f *file.File) errList {
	var errs errList

	fileutil.Walk(f.Scope, func(parents []fileutil.WalkContext, ctx fileutil.WalkContext) (dive bool, err error) {
		_, ok := (*ctx.Item).(file.Mixin)
		if ok {
			recErrs := l._checkRecursion(f, ptrOfSliceElem[file.ScopeItem, file.Mixin](ctx.Scope, ctx.Index))
			errs.PushBackList(&recErrs)
			return false, nil
		}

		return true, nil
	})

	return errs
}

func (l *Linker) _checkRecursion(f *file.File, m *file.Mixin, mcs ...file.MixinCall) errList {
	for i, mc := range mcs {
		if mc.Mixin.Mixin == m {
			if i == 0 {
				return list.List1(&corgierr.Error{
					Message: "recursion",
					ErrorAnnotation: anno.Anno(f, anno.Annotation{
						Start:      m.Position,
						Len:        (m.Name.Col - m.Col) + len(m.Name.Ident),
						Annotation: "this mixin",
					}),
					HintAnnotations: []corgierr.Annotation{
						anno.Anno(f, anno.Annotation{
							Start:      mcs[i].Position,
							Len:        (mcs[i].Name.Col - mcs[i].Col) + len(mcs[i].Name.Ident),
							Annotation: "calls itself",
						}),
					},
					Suggestions: []corgierr.Suggestion{
						{
							Suggestion: "Corgi (in it's current version) doesn't allow recursion.\n" +
								"Therefore, rewrite the mixin so that it doesn't call itself.",
						},
					},
				})
			}

			has := make([]corgierr.Annotation, 0, 2*i+1)
			for j, mc := range mcs[:i] {
				has = append(has, anno.Anno(mc.Mixin.File, anno.Annotation{
					Start:      mc.Position,
					Len:        (mc.Name.Col - mc.Col) + len(mc.Name.Ident),
					Annotation: fmt.Sprint(j*2+2, ": calls this mixin"),
				}))
				has = append(has, anno.Anno(mc.Mixin.File, anno.Annotation{
					Start:      mc.Mixin.Mixin.Position,
					Len:        (mc.Mixin.Mixin.Name.Col - mc.Mixin.Mixin.Col) + len(mc.Mixin.Mixin.Name.Ident),
					Annotation: fmt.Sprint(j*2+3, ": that mixin"),
				}))
			}

			has = append(has, anno.Anno(f, anno.Annotation{
				Start:      mcs[i].Position,
				Len:        (mcs[i].Name.Col - mcs[i].Col) + len(mcs[i].Name.Ident),
				Annotation: fmt.Sprint(2*i+2, ": calls the first mixin"),
			}))

			return list.List1(&corgierr.Error{
				Message: "recursion",
				ErrorAnnotation: anno.Anno(f, anno.Annotation{
					Start:      m.Position,
					Len:        (m.Name.Col - m.Col) + len(m.Name.Ident),
					Annotation: "1: this mixin",
				}),
				HintAnnotations: has,
				Suggestions: []corgierr.Suggestion{
					{
						Suggestion: "Corgi doesn't allow recursion. Therefore, rewrite the mixin so that it doesn't call itself.\n" +
							"A future version might support recursion.",
					},
				},
			})
		}
	}

	s := m.Body
	if len(mcs) > 0 {
		s = mcs[len(mcs)-1].Mixin.Mixin.Body
		if mcs[len(mcs)-1].Mixin.Mixin.Precompiled != nil {
			return errList{}
		}
	}

	var errs errList
	fileutil.Walk(s, func(parents []fileutil.WalkContext, ctx fileutil.WalkContext) (dive bool, err error) {
		mc, ok := (*ctx.Item).(file.MixinCall)
		if ok {
			errs = l._checkRecursion(f, m, append(mcs, mc)...)
			if errs.Len() > 0 {
				return false, fileutil.StopWalk
			}
		}

		return true, nil
	})

	return errs
}
