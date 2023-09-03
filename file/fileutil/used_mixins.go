package fileutil

import "github.com/mavolin/corgi/file"

type (
	// UsedMixins lists all mixins used by a given file, if it were printed.
	//
	// Depending on whether [ListUsedMixins] [LibraryDependencies] was called,
	// some fields might not be set.
	UsedMixins struct {
		// External lists all used mixins from library files.
		//
		// If LibraryDependencies was called, only direct dependencies will be
		// included here.
		External []UsedLibrary

		// Self lists all mixins from the calling library that are used by it.
		//
		// It is only present for calls to [LibraryDependencies].
		Self []UsedMixin
	}

	UsedLibrary struct {
		// Library is the library that is being referenced.
		Library *file.Library

		Mixins []UsedMixin
	}

	UsedMixin struct {
		Mixin *file.Mixin

		// RequiredBy lists the mixins requiring this mixin in the library
		// being analyzed.
		//
		// Only present/valid for [LibraryDependencies].
		RequiredBy []string
	}
)

type usedMixinsLister struct {
	UsedMixins

	self *file.Library

	_stack     []*file.File
	stackStart int

	inMixin bool
}

func (l *usedMixinsLister) stack() []*file.File {
	return l._stack[l.stackStart:]
}

func (l *usedMixinsLister) insertMixinCall(mc file.MixinCall, requiredBy string, direct bool) {
	if mc.Mixin.File.Library == nil {
		return
	}

	if l.self != nil && EqualLibrary(l.self, mc.Mixin.File.Library) {
		l.insertSelfMixinCall(mc, requiredBy, direct)
		return
	}

	for i, lib := range l.External {
		if !EqualLibrary(mc.Mixin.File.Library, lib.Library) {
			continue
		}

		for i, m := range lib.Mixins {
			if m.Mixin.Name.Ident == mc.Name.Ident {
				m.RequiredBy = append(m.RequiredBy, requiredBy)
				lib.Mixins[i] = m
				return
			}
		}

		var requiredBySlice []string
		if requiredBy != "" {
			requiredBySlice = []string{requiredBy}
		}
		l.External[i].Mixins = append(l.External[i].Mixins, UsedMixin{
			Mixin:      mc.Mixin.Mixin,
			RequiredBy: requiredBySlice,
		})
		if !direct {
			l.listDeps(mc.Mixin.File.Library, mc.Mixin.Mixin)
		}
		return
	}

	l.External = append(l.External, UsedLibrary{
		Library: mc.Mixin.File.Library,
		Mixins: []UsedMixin{
			{
				Mixin:      mc.Mixin.Mixin,
				RequiredBy: []string{requiredBy},
			},
		},
	})
	if !direct {
		l.listDeps(mc.Mixin.File.Library, mc.Mixin.Mixin)
	}
}

func (l *usedMixinsLister) insertSelfMixinCall(b file.MixinCall, requiredBy string, direct bool) {
	for i, a := range l.Self {
		if a.Mixin.Name.Ident != b.Name.Ident {
			continue
		}

		if requiredBy != "" {
			for _, other := range a.RequiredBy {
				if requiredBy == other {
					return
				}
			}

			a.RequiredBy = append(a.RequiredBy, requiredBy)
		}
		l.Self[i] = a
		return
	}

	l.Self = append(l.Self, UsedMixin{
		Mixin:      b.Mixin.Mixin,
		RequiredBy: []string{requiredBy},
	})

	if !direct {
		l.listDeps(b.Mixin.File.Library, b.Mixin.Mixin)
	}
}

func (l *usedMixinsLister) insertPrecomp(lib *file.Library, m *file.Mixin) {
	for i, ulib := range l.External {
		if !EqualLibrary(lib, ulib.Library) {
			continue
		}

		for _, um := range ulib.Mixins {
			if um.Mixin.Name.Ident == m.Name.Ident {
				return
			}
		}

		l.External[i].Mixins = append(l.External[i].Mixins, UsedMixin{
			Mixin: m,
		})
		l.listDeps(lib, m)
		return
	}

	l.External = append(l.External, UsedLibrary{
		Library: lib,
		Mixins:  []UsedMixin{{Mixin: m}},
	})
	l.listDeps(lib, m)
}

func (l *usedMixinsLister) listDeps(lib *file.Library, m *file.Mixin) {
	if !lib.Precompiled {
		l.listScope(m.Body, "", false)
		return
	}

	for _, libDep := range lib.Dependencies {
		for _, mDep := range libDep.Mixins {
			for _, requiredBy := range mDep.RequiredBy {
				if requiredBy == m.Name.Ident {
					l.insertPrecomp(libDep.Library, mDep.Mixin)
					break
				}
			}
		}
	}

	for _, b := range lib.Mixins {
		b := b

		for _, requiredBy := range b.RequiredBy {
			if requiredBy == m.Name.Ident {
				l.insertPrecomp(lib, &b.Mixin)
				break
			}
		}
	}
}

func (l *usedMixinsLister) listScope(s file.Scope, requiredBy string, direct bool) {
	Walk(s, func(parents []WalkContext, ctx WalkContext) (dive bool, err error) { //nolint:errcheck
		switch itm := (*ctx.Item).(type) {
		case file.MixinCall:
			oldInMixin := l.inMixin
			l.inMixin = true
			l.insertMixinCall(itm, requiredBy, direct)
			l.inMixin = oldInMixin
		case file.Mixin:
			oldInMixin := l.inMixin
			l.inMixin = true
			l.listScope(itm.Body, requiredBy, direct)
			l.inMixin = oldInMixin
			return false, nil
		case file.IfBlock:
			if l.inMixin {
				return true, nil
			}

			fill, _ := resolveTemplateBlock(l, itm.Name.Ident)
			if fill != nil {
				l.listScope(itm.Then, requiredBy, direct)
				return false, nil
			}

			for _, elseIf := range itm.ElseIfs {
				fill, _ := resolveTemplateBlock(l, elseIf.Name.Ident)
				if fill != nil {
					l.listScope(elseIf.Then, requiredBy, direct)
					return false, nil
				}
			}

			if itm.Else != nil {
				l.listScope(itm.Else.Then, requiredBy, direct)
				return false, nil
			}

			return false, nil
		case file.Block:
			if l.inMixin {
				return true, nil
			}

			fill, stackPos := resolveTemplateBlock(l, itm.Name.Ident)
			if fill == nil {
				fill, stackPos = &itm, l.stackStart
			}

			oldStart := l.stackStart
			l.stackStart = stackPos
			l.listScope(fill.Body, requiredBy, direct)
			l.stackStart = oldStart
		case file.And:
			l.listAttributeCollections(itm.Attributes, requiredBy, direct)
		case file.Element:
			l.listAttributeCollections(itm.Attributes, requiredBy, direct)
		case file.DivShorthand:
			l.listAttributeCollections(itm.Attributes, requiredBy, direct)
		case file.InlineText:
			l.listTextLines([]file.TextLine{itm.Text}, requiredBy, direct)
		case file.ArrowBlock:
			l.listTextLines(itm.Lines, requiredBy, direct)
		}

		return true, nil
	})
}

func resolveTemplateBlock(ctx *usedMixinsLister, name string) (b *file.Block, stackPos int) {
	stack := ctx.stack()[1:]
	for i := len(stack) - 1; i >= 0; i-- {
		f := stack[i]
		for _, itm := range f.Scope {
			fill, ok := itm.(file.Block)
			if !ok {
				continue
			}

			if fill.Name.Ident == name {
				return &fill, i
			}
		}
	}

	return nil, -1
}

func (l *usedMixinsLister) listAttributeCollections(acolls []file.AttributeCollection, requiredBy string, direct bool) {
	for _, acoll := range acolls {
		alist, ok := acoll.(file.AttributeList)
		if !ok {
			continue
		}

		for _, attr := range alist.Attributes {
			mcAttr, ok := attr.(file.MixinCallAttribute)
			if ok {
				l.insertMixinCall(mcAttr.MixinCall, requiredBy, direct)
			}
		}
	}
}

func (l *usedMixinsLister) listTextLines(lns []file.TextLine, requiredBy string, direct bool) {
	for _, ln := range lns {
		for _, itm := range ln {
			switch itm := itm.(type) {
			case file.MixinCallInterpolation:
				l.insertMixinCall(itm.MixinCall, requiredBy, direct)
			case file.ElementInterpolation:
				l.listAttributeCollections(itm.Element.Attributes, requiredBy, direct)
			}
		}
	}
}

// ============================================================================
// ListUsedMixins
// ======================================================================================

func ListUsedMixins(f *file.File) UsedMixins {
	var l usedMixinsLister

	var n int
	for f := f; ; {
		n++
		if f.Extend == nil {
			break
		}
		f = f.Extend.File
	}
	l._stack = make([]*file.File, n)
	for i := n - 1; i >= 0; i-- {
		l._stack[i] = f
		if f.Extend != nil {
			f = f.Extend.File
		}
	}

	l.listScope(f.Scope, "", false)
	return l.UsedMixins
}

// ============================================================================
// LibraryDependencies
// ======================================================================================

func LibraryDependencies(lib *file.Library) UsedMixins {
	var l usedMixinsLister
	l.self = lib

	for _, f := range lib.Files {
		l._stack = []*file.File{f}

		for _, itm := range f.Scope {
			m, ok := itm.(file.Mixin)
			if ok {
				l.listScope(m.Body, m.Name.Ident, true)
			}
		}
	}

	return l.UsedMixins
}
