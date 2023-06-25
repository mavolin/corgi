package link

import (
	"errors"

	"github.com/mavolin/corgi/corgierr"
	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/fileutil"
	"github.com/mavolin/corgi/internal/anno"
	"github.com/mavolin/corgi/internal/list"
)

func (l *Linker) linkMixinCalls(f *file.File) errList {
	var errs errList

	fileutil.Walk(f.Scope, func(parents []fileutil.WalkContext, ctx fileutil.WalkContext) (dive bool, err error) {
		incl, ok := (*ctx.Item).(file.MixinCall)
		if !ok {
			return true, nil
		}

		mcErrs := l.linkMixinCall(f, &incl)
		errs.PushBackList(&mcErrs)
		return true, err
	})

	return errs
}

func (l *Linker) linkMixinCall(f *file.File, parents []fileutil.WalkContext, mc *file.MixinCall) errList {

	inclFile, err := l.loader.LoadMixinCall(f, fileutil.Unquote(mc.Path))
	if err != nil {
		var cerr *corgierr.Error
		if errors.As(err, &cerr) {
			return list.List1(cerr)
		}

		return list.List1(&corgierr.Error{
			Message: "failed to load mixinCalld file",
			ErrorAnnotation: anno.Anno(f, anno.Annotation{
				Start:      mc.Position,
				ToEOL:      true,
				Annotation: err.Error(),
			}),
			Cause: err,
		})
	}

	if inclFile == nil {
		return list.List1(&corgierr.Error{
			Message: "mixinCalld file not found",
			ErrorAnnotation: anno.Anno(f, anno.Annotation{
				Start:      mc.Position,
				ToEOL:      true,
				Annotation: "there is no file with this path",
			}),
		})
	}

	mc.MixinCall = inclFile
	return errList{}
}
