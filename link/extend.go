package link

import (
	"errors"

	"github.com/mavolin/corgi/corgierr"
	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/fileutil"
	"github.com/mavolin/corgi/internal/anno"
	"github.com/mavolin/corgi/internal/list"
)

func (l *Linker) linkExtend(ctx *context, f *file.File) {
	if f.Extend == nil {
		return
	}

	ctx.n++

	go func() {
		template, err := l.loader.LoadTemplate(fileutil.Unquote(f.Extend.Path))
		if err != nil {
			var cerr *corgierr.Error
			if errors.As(err, &cerr) {
				ctx.errs <- list.List1(cerr)
				return
			}

			ctx.errs <- list.List1(&corgierr.Error{
				Message: "failed to load template",
				ErrorAnnotation: anno.Anno(f, anno.Annotation{
					Start:      f.Extend.Position,
					ToEOL:      true,
					Annotation: err.Error(),
				}),
				Cause: err,
			})
			return
		}

		if template == nil {
			ctx.errs <- list.List1(&corgierr.Error{
				Message: "template not found",
				ErrorAnnotation: anno.Anno(f, anno.Annotation{
					Start:      f.Extend.Position,
					ToEOL:      true,
					Annotation: "there is no template with this path",
				}),
			})
		}

		f.Extend.File = template
		ctx.errs <- errList{}
	}()
}
