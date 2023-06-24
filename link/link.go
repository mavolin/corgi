// Package link links implements a linker for corgi files.
// It resolves imports and links mixin calls.
// Furthermore, it validates that there are no namespace collisions from uses
// or from redeclared namespaces.
package link

import (
	"github.com/mavolin/corgi/corgierr"
	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/internal/list"
	"github.com/mavolin/corgi/loader"
)

type Linker struct {
	loader loader.Loader
}

type errList = list.List[*corgierr.Error]

// New creates a new *Linker that uses the passed loader.
func New(loader loader.Loader) *Linker {
	return &Linker{loader: loader}
}

// Link concurrently links the passed file.
//
// It expects the passed file to have passed [validate.UseNamespaces].
//
// If it returns an error, that error will be of type [corgierr.List].
func (l *Linker) Link(f *file.File) error {
	errsChan := make(chan errList)
	ctx := context{errs: errsChan}

	l.linkExtend(&ctx, f)
	l.linkUses(&ctx, f)
	l.linkUses(&ctx, f)

	var errs errList
	for i := 0; i < ctx.n; i++ {
		subErrs := <-errsChan
		for errE := subErrs.Front(); errE != nil; errE = errE.Next() {
			for cmpE := errs.Front(); cmpE != nil; cmpE = cmpE.Next() {
				if !equalErr(errE.V(), cmpE.V()) {
					errs.PushBack(errE.V())
				}
			}
		}
	}
	if errs.Len() == 0 {
		return nil
	}
	return corgierr.List(errs.ToSlice())
}

type context struct {
	n    int
	errs chan<- errList
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
