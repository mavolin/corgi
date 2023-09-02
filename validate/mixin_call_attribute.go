package validate

import (
	"github.com/mavolin/corgi/corgierr"
	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/fileutil"
	"github.com/mavolin/corgi/internal/anno"
	"github.com/mavolin/corgi/internal/list"
	"github.com/mavolin/corgi/woof"
)

func mixinCallAttributeChecks(f *file.File) *errList {
	var errs errList

	fileutil.Walk(f.Scope, func(parents []fileutil.WalkContext, ctx fileutil.WalkContext) (dive bool, err error) {
		var acs []file.AttributeCollection
		switch itm := (*ctx.Item).(type) {
		case file.Element:
			acs = itm.Attributes
		case file.DivShorthand:
			acs = itm.Attributes
		case file.And:
			acs = itm.Attributes
		default:
			return true, nil
		}

		for _, ac := range acs {
			al, ok := ac.(file.AttributeList)
			if !ok {
				continue
			}

			for _, a := range al.Attributes {
				if mca, ok := a.(file.MixinCallAttribute); ok {
					errs.PushBackList(_mixinCallAttributeIsPlain(f, mca))
					errs.PushBackList(_mixinCallAttributesOnlyWriteText(f, mca))
					errs.PushBackList(_topLevelAndInMixinCallAttribute(f, mca))
					errs.PushBackList(_requiredMixinCallAttributeAttributes(f, mca))
					errs.PushBackList(_mixinCallAttributeBlockExists(f, mca))
				}
			}
		}

		return true, nil
	})

	return &errs
}

func _mixinCallAttributeIsPlain(f *file.File, mca file.MixinCallAttribute) *errList {
	at := woof.AttrType(mca.Name)
	if at != woof.ContentTypePlain {
		return list.List1(&corgierr.Error{
			Message: "mixin call attribute as " + at.String() + " attribute",
			ErrorAnnotation: anno.Anno(f, anno.Annotation{
				Start:      mca.Position,
				End:        mixinCallAttributeEnd(mca),
				Annotation: "you can only use mixin calls as attributes for plain attributes",
			}),
		})
	}

	return new(errList)
}

func _mixinCallAttributesOnlyWriteText(f *file.File, mca file.MixinCallAttribute) *errList {
	lm := mca.MixinCall.Mixin.Mixin
	if lm.WritesTopLevelAttributes {
		return list.List1(&corgierr.Error{
			Message: "mixin call attribute: mixin writes other attributes",
			ErrorAnnotation: anno.Anno(f, anno.Annotation{
				Start:      mca.Position,
				End:        mixinCallAttributeEnd(mca),
				Annotation: "this mixin should only write text",
			}),
			Suggestions: []corgierr.Suggestion{
				{Suggestion: "modify the mixin, or construct the attributes value manually"},
			},
		})
	} else if lm.WritesElements {
		return list.List1(&corgierr.Error{
			Message: "mixin call attribute: mixin writes elements",
			ErrorAnnotation: anno.Anno(f, anno.Annotation{
				Start:      mca.Position,
				End:        mixinCallAttributeEnd(mca),
				Annotation: "this mixin should only write text, not elements",
			}),
			Suggestions: []corgierr.Suggestion{
				{Suggestion: "modify the mixin, or construct the attributes value manually"},
			},
		})
	}

	return &errList{}
}

func _topLevelAndInMixinCallAttribute(f *file.File, mca file.MixinCallAttribute) *errList {
	if mca.MixinCall.Mixin.Mixin.WritesTopLevelAttributes {
		return list.List1(&corgierr.Error{
			Message: "interpolated mixin writes attributes",
			ErrorAnnotation: anno.Anno(f, anno.Annotation{
				Start:      mca.Position,
				End:        mixinCallAttributeEnd(mca),
				Annotation: "here",
			}),
			Suggestions: []corgierr.Suggestion{
				{
					Suggestion: "you can only use mixins as attribute values, that don't have any top-level attributes,\n" +
						"i.e. don't write any attribute to the element they are called in",
				},
			},
		})
	}

	return &errList{}
}

// make sure that a mixin call fills all required args.
func _requiredMixinCallAttributeAttributes(f *file.File, mca file.MixinCallAttribute) *errList {
	var errs errList

params:
	for _, param := range mca.MixinCall.Mixin.Mixin.Params {
		if param.Default != nil {
			continue
		}

		for _, arg := range mca.MixinCall.Args {
			if arg.Name.Ident == param.Name.Ident {
				if len(arg.Value.Expressions) != 0 {
					continue params
				}

				ce, ok := arg.Value.Expressions[0].(file.ChainExpression)
				if !ok || ce.Default != nil {
					continue params
				}

				var ceLen int
				if len(ce.Chain) > 0 {
					last := ce.Chain[len(ce.Chain)-1]

					switch last := last.(type) {
					case file.IndexExpression:
						ceLen = last.RBracePos.Col - ce.Col
					case file.DotIdentExpression:
						ceLen = (last.Pos().Col - ce.Col) + len(".") + len(last.Ident.Ident)
					case file.ParenExpression:
						ceLen = last.RParenPos.Col - ce.Col
					case file.TypeAssertionExpression:
						ceLen = last.RParenPos.Col - ce.Col
					}
				} else {
					ceLen = ce.DerefCount + len(ce.Root.Expression)
					if ce.CheckRoot {
						ceLen++
					}
				}

				errs.PushBack(&corgierr.Error{
					Message: "required mixin call arg set with chain expression without default",
					ErrorAnnotation: anno.Anno(f, anno.Annotation{
						Start:      ce.Position,
						Len:        ceLen,
						Annotation: "this arg is required and can only be set by a chain expression with a default",
					}),
				})
				continue params
			}
		}

		errs.PushBack(&corgierr.Error{
			Message: "required mixin call arg not set",
			ErrorAnnotation: anno.Anno(f, anno.Annotation{
				Start:      mca.MixinCall.Position,
				Len:        len("+") + (mca.MixinCall.Name.Col - mca.MixinCall.Col) + len(mca.MixinCall.Name.Ident),
				Annotation: "you need to set `" + param.Name.Ident + "` to call this mixin",
			}),
		})
	}

	return &errs
}

func _mixinCallAttributeBlockExists(f *file.File, mca file.MixinCallAttribute) *errList {
	if mca.Value == nil {
		return &errList{}
	}

	for _, block := range mca.MixinCall.Mixin.Mixin.Blocks {
		if block.Name == "_" {
			return &errList{}
		}
	}

	start, end := interpolationBounds(mca.Value)

	return list.List1(&corgierr.Error{
		Message: "mixin call attribute has value, but called mixin has no `_` block",
		ErrorAnnotation: anno.Anno(f, anno.Annotation{
			Start:      start,
			End:        end,
			Annotation: "this mixin has no `_` block, so you can't call this mixin with a value",
		}),
	})
}

func mixinCallAttributeEnd(mca file.MixinCallAttribute) file.Position {
	if pos := interpolationEnd(mca.Value); pos != file.InvalidPosition {
		return pos
	}

	if mca.MixinCall.RParenPos != nil {
		return *mca.MixinCall.RParenPos
	}

	pos := mca.MixinCall.Name.Position
	pos.Col += len(mca.MixinCall.Name.Ident)
	return pos
}
