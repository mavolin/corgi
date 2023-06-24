package link

import (
	"errors"

	"github.com/mavolin/corgi/corgierr"
	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/fileutil"
	"github.com/mavolin/corgi/internal/anno"
	"github.com/mavolin/corgi/internal/list"
)

func (l *Linker) linkIncludes(lctx *context, f *file.File) {
	fileutil.Walk(f.Scope, func(parents []fileutil.WalkContext, ctx fileutil.WalkContext) (dive bool, err error) {
		incl, ok := (*ctx.Item).(file.Include)
		if !ok {
			return true, nil
		}

		lctx.n++
		go func() {
			errs := l.linkInclude(f, &incl)
			*ctx.Item = incl
			lctx.errs <- errs
		}()
		return false, err
	})
}

func (l *Linker) linkInclude(f *file.File, incl *file.Include) errList {
	inclFile, err := l.loader.LoadInclude(f, fileutil.Unquote(incl.Path))
	if err != nil {
		var cerr *corgierr.Error
		if errors.As(err, &cerr) {
			return list.List1(cerr)
		}

		return list.List1(&corgierr.Error{
			Message: "failed to load included file",
			ErrorAnnotation: anno.Anno(f, anno.Annotation{
				Start:      incl.Position,
				ToEOL:      true,
				Annotation: err.Error(),
			}),
			Cause: err,
		})
	}

	if inclFile == nil {
		return list.List1(&corgierr.Error{
			Message: "included file not found",
			ErrorAnnotation: anno.Anno(f, anno.Annotation{
				Start:      incl.Position,
				ToEOL:      true,
				Annotation: "there is no file with this path",
			}),
		})
	}

	incl.Include = inclFile
	return errList{}
}
