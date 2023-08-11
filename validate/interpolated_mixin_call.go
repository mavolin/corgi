package validate

import (
	"github.com/mavolin/corgi/corgierr"
	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/fileutil"
	"github.com/mavolin/corgi/internal/anno"
	"github.com/mavolin/corgi/internal/list"
)

func interpolatedMixinCallChecks(f *file.File) *errList {
	var errs errList

	fileutil.Walk(f.Scope, func(parents []fileutil.WalkContext, ctx fileutil.WalkContext) (dive bool, err error) {
		var lines []file.TextLine
		switch itm := (*ctx.Item).(type) {
		case file.ArrowBlock:
			lines = itm.Lines
		case file.InlineText:
			lines = []file.TextLine{itm.Text}
		default:
			return true, nil
		}

		for _, line := range lines {
			for _, itm := range line {
				if mci, ok := itm.(file.MixinCallInterpolation); ok {
					errs.PushBackList(_topLevelAndInInterpolatedMixinCall(f, mci))
					errs.PushBackList(_requiredInterpolatedMixinCallAttributes(f, mci))
					errs.PushBackList(_interpolatedMixinCallBlockExists(f, mci))
				}
			}
		}

		return true, nil
	})

	return &errs
}

func _topLevelAndInInterpolatedMixinCall(f *file.File, mci file.MixinCallInterpolation) *errList {
	if mci.MixinCall.Mixin.Mixin.WritesTopLevelAttributes {
		return list.List1(&corgierr.Error{
			Message: "interpolated mixin writes attributes",
			ErrorAnnotation: anno.Anno(f, anno.Annotation{
				Start:      mci.Position,
				End:        interpolationEnd(mci.Value),
				Annotation: "here",
			}),
			Suggestions: []corgierr.Suggestion{
				{
					Suggestion: "you can only interpolate mixins that don't have any top-level attributes,\n" +
						"i.e. don't write any attribute to the element they are called in",
				},
			},
		})
	}

	return &errList{}
}

// make sure that a mixin call fills all required args.
func _requiredInterpolatedMixinCallAttributes(f *file.File, mci file.MixinCallInterpolation) *errList {
	var errs errList

	annoLen := len("+") + len(mci.MixinCall.Name.Ident)
	if mci.MixinCall.Namespace != nil {
		annoLen += len(mci.MixinCall.Namespace.Ident) + len(".") //nolint:ineffassign
	}

params:
	for _, param := range mci.MixinCall.Mixin.Mixin.Params {
		if param.Default != nil {
			continue
		}

		for _, arg := range mci.MixinCall.Args {
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
				Start:      mci.MixinCall.Position,
				Len:        len("+") + (mci.MixinCall.Name.Col - mci.MixinCall.Col) + len(mci.MixinCall.Name.Ident),
				Annotation: "you need to set `" + param.Name.Ident + "` to call this mixin",
			}),
		})
	}

	return &errs
}

func _interpolatedMixinCallBlockExists(f *file.File, mci file.MixinCallInterpolation) *errList {
	if mci.Value == nil {
		return &errList{}
	}

	for _, block := range mci.MixinCall.Mixin.Mixin.Blocks {
		if block.Name == "_" {
			return &errList{}
		}
	}

	start, end := interpolationBounds(mci.Value)

	return list.List1(&corgierr.Error{
		Message: "interpolated mixin call has value, but called mixin has no `_` block",
		ErrorAnnotation: anno.Anno(f, anno.Annotation{
			Start:      start,
			End:        end,
			Annotation: "this mixin has no `_` block, so you can't call this mixin with a value",
		}),
	})
}
