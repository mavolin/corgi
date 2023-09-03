package validate

import (
	"github.com/mavolin/corgi/corgierr"
	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/fileutil"
	"github.com/mavolin/corgi/internal/anno"
	"github.com/mavolin/corgi/internal/list"
)

// check that only template files contain template block placeholders.
func onlyTemplateFilesContainBlockPlaceholders(f *file.File) *errList {
	if f.Type == file.TypeTemplate {
		return &errList{}
	}

	var errs errList

	fileutil.Walk(f.Scope, func(parents []fileutil.WalkContext, ctx fileutil.WalkContext) (dive bool, err error) {
		switch itm := (*ctx.Item).(type) {
		case file.Include:
			return false, nil
		case file.Mixin:
			return false, nil
		case file.Block:
			if len(parents) == 0 {
				if f.Extend != nil {
					return false, nil
				}

				errs.PushBack(&corgierr.Error{
					Message: "use of template block without extending a template",
					ErrorAnnotation: anno.Anno(f, anno.Annotation{
						Start:      itm.Position,
						Len:        (itm.Name.Col - itm.Col) + len(itm.Name.Ident),
						Annotation: "you can't fill a template block if you aren't extending a template",
					}),
					Suggestions: []corgierr.Suggestion{
						{
							Suggestion: "Template blocks are used to fill placeholders in a template file.\n" +
								"To fill such a placeholder, you must place an `extend` directive at the start of the file.",
						},
					},
				})
			}

			// don't accidentally report (ill-placed) nested mixin call blocks
			reportErr := true
			var parentBlock bool
			for i := len(parents) - 1; i >= 0; i-- {
				switch (*parents[i].Item).(type) {
				case file.Block:
					parentBlock = true
				case file.MixinCall:
					if !parentBlock {
						reportErr = false
					}
				}
			}

			if !reportErr {
				return true, nil
			}

			errs.PushBack(&corgierr.Error{
				Message: "use of template block outside of template file",
				ErrorAnnotation: anno.Anno(f, anno.Annotation{
					Start:      itm.Position,
					Len:        (itm.Name.Col - itm.Col) + len(itm.Name.Ident),
					Annotation: "cannot place a template block in a main or include file",
				}),
				Suggestions: []corgierr.Suggestion{
					{
						Suggestion: "This block belongs neither to a mixin nor to a mixin call.\n" +
							"Since this isn't a template file, you cannot place this block here.",
					},
				},
			})
			return false, nil
		default:
			return true, nil
		}
	})

	return &errs
}

func duplicateTemplateBlocks(f *file.File) *errList {
	if f.Extend == nil {
		return &errList{}
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

	return &errs
}

func nonExistentTemplateBlocks(f *file.File) *errList {
	if f.Extend == nil {
		return &errList{}
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
			fileutil.Walk(extend.File.Scope,
				func(parents []fileutil.WalkContext, ctx fileutil.WalkContext) (dive bool, err error) {
					switch itm := (*ctx.Item).(type) {
					case file.Include:
						return false, nil
					case file.Mixin:
						return false, nil
					case file.Block:
						if itm.Name.Ident != block.Name.Ident {
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
				Annotation: "this template block never appears in the template you are extending",
			}),
		})
	}

	return &errs
}
