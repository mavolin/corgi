// Package link links implements a linker for corgi files.
// It resolves imports and links mixin calls.
// Furthermore, it validates that there are no namespace collisions from uses
// or from redeclared namespaces.
package link

import (
	"unsafe"

	"github.com/mavolin/corgi/corgierr"
	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/internal/list"
	"github.com/mavolin/corgi/load"
)

type Linker struct {
	loader load.Loader
}

type errList = list.List[*corgierr.Error]

// New creates a new *Linker that uses the passed load.
func New(loader load.Loader) *Linker {
	return &Linker{loader: loader}
}

// LinkFile concurrently links the passed file.
//
// It expects the passed file to have passed [validate.PreLink] and have
// a linked DirLibrary, or none at all.
//
// If it returns an error, that error will be of type [corgierr.List].
func (l *Linker) LinkFile(f *file.File) error {
	errsChan := make(chan *errList)
	ctx := context{errs: errsChan}

	l.linkExtend(&ctx, f)
	l.linkUses(&ctx, f)
	l.linkIncludes(&ctx, f)

	var errs list.List[*corgierr.Error]
	for i := 0; i < ctx.n; i++ {
		subErrs := <-errsChan
	subErrs:
		for errE := subErrs.Front(); errE != nil; errE = errE.Next() {
			for cmpE := errs.Front(); cmpE != nil; cmpE = cmpE.Next() {
				if equalErr(errE.V(), cmpE.V()) {
					continue subErrs
				}
			}
			errs.PushBack(errE.V())
		}
	}

	mcErrs := l.linkMixinCalls(f)
	errs.PushBackList(mcErrs)

	recErrs := l.checkRecursion(f)
	errs.PushBackList(recErrs)
	if recErrs.Len() > 0 {
		return corgierr.List(errs.ToSlice())
	}

	mErrs := l.analyzeMixins(f)
	errs.PushBackList(mErrs)

	if errs.Len() == 0 {
		return nil
	}
	return corgierr.List(errs.ToSlice())
}

// LinkLibrary concurrently links the passed library.
//
// It expects all files in the passed library to have passed [validate.PreLink].
//
// If it returns an error, that error will be of type [corgierr.List].
func (l *Linker) LinkLibrary(lib *file.Library) error {
	errsChan := make(chan *errList)
	ctx := context{errs: errsChan}

	l.linkDependencies(&ctx, lib)

	for _, f := range lib.Files {
		f := f
		l.linkUses(&ctx, f)
		l.linkIncludes(&ctx, f)
	}

	var errs errList
	for i := 0; i < ctx.n; i++ {
		subErrs := <-errsChan
	subErrs:
		for errE := subErrs.Front(); errE != nil; errE = errE.Next() {
			for cmpE := errs.Front(); cmpE != nil; cmpE = cmpE.Next() {
				if equalErr(errE.V(), cmpE.V()) {
					continue subErrs
				}
			}
			errs.PushBack(errE.V())
		}
	}

	var hasRecursionErrs bool
	for _, f := range lib.Files {
		f := f
		mcErrs := l.linkMixinCalls(f)
		errs.PushBackList(mcErrs)

		recErrs := l.checkRecursion(f)
		errs.PushBackList(recErrs)
		if recErrs.Len() > 0 {
			hasRecursionErrs = true
		}
	}
	if hasRecursionErrs {
		return corgierr.List(errs.ToSlice())
	}

	mErrs := l.analyzeMixins(lib.Files...)
	errs.PushBackList(mErrs)

	if errs.Len() == 0 {
		return nil
	}
	return corgierr.List(errs.ToSlice())
}

type context struct {
	n    int
	errs chan<- *errList
}

func equalErr(a, b *corgierr.Error) bool {
	aa, ba := a.ErrorAnnotation, b.ErrorAnnotation

	if aa.Start != ba.Start || aa.End != ba.End || aa.Annotation != ba.Annotation {
		return false
	}

	if len(a.HintAnnotations) != len(b.HintAnnotations) {
		return false
	}
	for i, ah := range a.HintAnnotations {
		bh := b.HintAnnotations[i]
		if ah.Start != bh.Start || ah.End != bh.End || ah.Annotation != bh.Annotation {
			return false
		}
	}

	if a.Example != b.Example || a.ShouldBe != b.ShouldBe {
		return false
	}

	if len(a.Suggestions) != len(b.Suggestions) {
		return false
	}
	for i, as := range a.Suggestions {
		bs := b.Suggestions[i]
		if as.Suggestion != bs.Suggestion || as.Example != bs.Example || as.ShouldBe != bs.ShouldBe || as.Code != bs.Code {
			return false
		}
	}

	if a.Cause != nil && b.Cause == nil {
		return false
	} else if a.Cause == nil {
		return true
	}
	return a.Cause.Error() == b.Cause.Error()
}

// This looks worse than it actually is.
// Basically we need sometimes need pointers to elements in a slice
// (more specifically pointers to mixins for linking).
// This is fine, if the element to reference is a concrete type.
//
// However, referencing the "concrete" element in a slice of interfaces is a
// whole different story.
//
// So where we could just &s[i] for a slice s []file.Mixin to obtain a ptr to
// the mixin, we can't do the same for s since interface values can't be
// addressed:
// Something like &(s[i].(file.Mixin)) isn't valid since interfaces are
// supposed to be immutable.
//
// Luckily, the mem representation of an interface contains a pointer to the
// type it represents, which we can extract with a bit of pointer magic in the
// lines below.
//
// Also: ptrToX sounded to much like a conversion, so I chose ptrOf
func ptrOfSliceElem[E any, T any](s []E, i int) *T {
	eface := (*struct { // this is how Go stores an interface in memory
		t   unsafe.Pointer
		ptr unsafe.Pointer
	})(unsafe.Pointer(&s[i]))

	return (*T)(eface.ptr)
}
