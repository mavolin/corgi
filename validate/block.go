package validate

import (
	"github.com/mavolin/corgi/corgierr"
	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/fileutil"
	"github.com/mavolin/corgi/internal/anno"
	"github.com/mavolin/corgi/internal/list"
)

func duplicateTemplateBlocks(f *file.File) errList {
	if f.Extend == nil {
		return errList{}
	}

	var errs errList

	var cmpBlocks list.List[file.Block]

	for _, itm := range f.Scope {
		block, ok := itm.(file.Block)
		if !ok {
			continue
		}

		if cmpBlocks.Len() == 0 {
			continue
		}

		for cmpBlockE := cmpBlocks.Front(); cmpBlockE != nil; cmpBlockE = cmpBlockE.Next() {
			if block.Name.Ident == cmpBlockE.V().Name.Ident {
				errs.PushBack(&corgierr.Error{
					Message: "template block filled twice",
					ErrorAnnotation: anno.Anno(f, anno.Annotation{
						ContextLen: 2,
						Start:      block.Name.Position,
						ToEOL:      true,
						Annotation: "then here",
					}),
					HintAnnotations: []corgierr.Annotation{
						anno.Anno(f, anno.Annotation{
							ContextLen: 2,
							Start:      cmpBlockE.V().Name.Position,
							ToEOL:      true,
							Annotation: "first filled here",
						}),
					},
					Suggestions: []corgierr.Suggestion{
						{Suggestion: "merge these, or remove one of these blocks"},
					},
				})
			}
		}

		cmpBlocks.PushBack(block)
	}

	return errs
}

func nonExistentTemplateBlocks(f *file.File) errList {
	if f.Extend == nil {
		return errList{}
	}

	var errs errList

scope:
	for _, itm := range f.Scope {
		block, ok := itm.(file.Block)
		if !ok {
			continue
		}

		extend := f.Extend
		for extend != nil {
			var found bool
			fileutil.Walk(extend.File.Scope, func(parents []fileutil.WalkContext, ctx fileutil.WalkContext) (dive bool, err error) {
				switch itm := (*ctx.Item).(type) {
				case file.Mixin:
					return false, nil
				case file.Block:
					if itm.Name != block.Name {
						return true, nil
					}

					if len(parents) == 0 {
						found = true
						return false, fileutil.StopWalk
					}

					_, ok := (*parents[len(parents)-1].Item).(file.MixinCall)
					if !ok {
						found = true
						return false, fileutil.StopWalk
					}

					return true, nil
				default:
					return true, nil
				}
			})
			if found {
				continue scope
			}

			extend = extend.File.Extend
		}

		errs.PushBack(&corgierr.Error{
			Message: "unknown template block",
			ErrorAnnotation: anno.Anno(f, anno.Annotation{
				Start:      block.Position,
				End:        file.Position{Line: block.Line, Col: block.Name.Col + len(block.Name.Ident)},
				Annotation: "this template block is not used by the template you are extending",
			}),
		})
	}

	return errs
}
