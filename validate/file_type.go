package validate

import (
	"fmt"

	"github.com/mavolin/corgi/corgierr"
	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/fileutil"
	"github.com/mavolin/corgi/internal/anno"
	"github.com/mavolin/corgi/internal/list"
)

func mainFile(f *file.File) errList {
	if f.Type != file.TypeMain {
		return errList{}
	}

	if f.Func == nil {
		expectPos := file.Position{Line: 1, Col: 1}
		if len(f.Uses) > 0 {
			expectPos.Line = f.Uses[len(f.Uses)-1].Uses[len(f.Uses[len(f.Uses)-1].Uses)-1].Line + 1
		} else if len(f.Imports) > 0 {
			expectPos.Line = f.Imports[len(f.Imports)-1].Imports[len(f.Imports[len(f.Uses)-1].Imports)-1].Line + 1
		}

		return list.List1(&corgierr.Error{
			Message: "missing func header",
			ErrorAnnotation: anno.Anno(f, anno.Annotation{
				Start:      expectPos,
				Annotation: "expected the func header here",
			}),
			Example: "`func RenderFoo(num int)`",
		})
	}

	return errList{}
}

func templateFile(f *file.File) errList {
	if f.Type != file.TypeTemplate {
		return errList{}
	}

	var errs errList

	if f.Func != nil {
		errs.PushBack(&corgierr.Error{
			Message: "template file with `func` header",
			ErrorAnnotation: anno.Anno(f, anno.Annotation{
				Start:      f.Func.Position,
				Len:        len("func"),
				Annotation: "a template file shouldn't have a `func header",
			}),
		})
	}

	return errs
}

func extendingFile(f *file.File) errList {
	if f.Extend == nil {
		return errList{}
	}

	var errs errList

	fileutil.Walk(f.Scope, func(parents []fileutil.WalkContext, ctx fileutil.WalkContext) (dive bool, err error) {
		switch itm := (*ctx.Item).(type) {
		case file.Code:
		case file.Block:
		case file.Mixin:
		case file.CorgiComment:
		default:
			errs.PushBack(&corgierr.Error{
				Message: fmt.Sprintf("unexpected top-level item %T", itm),
				ErrorAnnotation: anno.Anno(f, anno.Annotation{
					Start: itm.Pos(),
					ToEOL: true,
					Annotation: "files extending other files may only have `block`, `append`, or `prepend` directives,\n" +
						"comments, code, or mixins as top-level items",
				}),
			})
		}
		return false, nil
	})
	return errs
}

func libraryFile(f *file.File) errList {
	if f.Type != file.TypeLibraryFile {
		return errList{}
	}

	var errs errList

	if f.Func != nil {
		errs.PushBack(&corgierr.Error{
			Message: "func header in use file",
			ErrorAnnotation: anno.Anno(f, anno.Annotation{
				Start:      f.Func.Position,
				Len:        len("func"),
				Annotation: "`func` headers cannot be used in use files",
			}),
		})
	}

	fileutil.Walk(f.Scope, func(parents []fileutil.WalkContext, ctx fileutil.WalkContext) (dive bool, err error) {
		switch itm := (*ctx.Item).(type) {
		case file.Code:
		case file.Mixin:
		case file.CorgiComment:
		default:
			errs.PushBack(&corgierr.Error{
				Message: fmt.Sprintf("unexpected top-level item %T", itm),
				ErrorAnnotation: anno.Anno(f, anno.Annotation{
					Start:      itm.Pos(),
					ToEOL:      true,
					Annotation: "use files may only have comments, code, or mixins as top-level items",
				}),
			})
		}
		return false, nil
	})
	return errs
}
