package link

import (
	"errors"

	"github.com/mavolin/corgi/corgierr"
	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/fileutil"
	"github.com/mavolin/corgi/internal/anno"
	"github.com/mavolin/corgi/internal/list"
)

func (l *Linker) linkUses(ctx *context, f *file.File) {
	for _, use := range f.Uses {
		use := use
		ctx.n += len(use.Uses)
		for specI := range use.Uses {
			specI := specI
			go func() {
				ctx.errs <- l.linkUseSpec(f, &use.Uses[specI])
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
				Annotation: "there is no library available under this path",
			}),
		})
	}

	spec.Library = lib
	return &errList{}
}
