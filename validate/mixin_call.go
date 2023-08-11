package validate

import (
	"fmt"

	"github.com/mavolin/corgi/corgierr"
	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/fileutil"
	"github.com/mavolin/corgi/internal/anno"
	"github.com/mavolin/corgi/internal/list"
)

func mixinCallChecks(f *file.File) *errList {
	var errs errList

	fileutil.Walk(f.Scope, func(parents []fileutil.WalkContext, ctx fileutil.WalkContext) (dive bool, err error) {
		mc, ok := (*ctx.Item).(file.MixinCall)
		if !ok {
			return true, nil
		}

		errs.PushBackList(_mixinCallArgsExist(f, mc))
		errs.PushBackList(_duplicateMixinCallArgs(f, mc))
		errs.PushBackList(_requiredMixinCallAttributes(f, mc))
		errs.PushBackList(_mixinCallBody(f, mc))
		errs.PushBackList(_mixinCallBlocksExist(f, mc))
		errs.PushBackList(_duplicateMixinCallBlocks(f, mc))
		errs.PushBackList(_mixinCallBlockAttrs(f, mc))

		return true, nil
	})

	return &errs
}

// mixin call args exist.
func _mixinCallArgsExist(f *file.File, mc file.MixinCall) *errList {
	var errs errList

args:
	for _, arg := range mc.Args {
		for _, param := range mc.Mixin.Mixin.Params {
			if arg.Name.Ident == param.Name.Ident {
				continue args
			}
		}

		errs.PushBack(&corgierr.Error{
			Message: "non-existent mixin call arg",
			ErrorAnnotation: anno.Anno(f, anno.Annotation{
				ContextStart: mc.Position,
				Start:        arg.Name.Position,
				Len:          len(arg.Name.Ident),
				Annotation:   "`" + mc.Mixin.Mixin.Name.Ident + "` doesn't have any param named `" + arg.Name.Ident + "`",
			}),
		})
	}

	return &errs
}

// mixin call doesn't specify any args twice.
func _duplicateMixinCallArgs(f *file.File, mc file.MixinCall) *errList {
	var errs errList

	for i, arg := range mc.Args {
		for _, cmp := range mc.Args[:i] {
			if arg.Name.Ident == cmp.Name.Ident {
				errs.PushBack(&corgierr.Error{
					Message: "duplicate mixin call arg",
					ErrorAnnotation: anno.Anno(f, anno.Annotation{
						ContextStart: mc.Position,
						Start:        arg.Name.Position,
						Len:          len(arg.Name.Ident),
						Annotation:   "and then here again",
					}),
					HintAnnotations: []corgierr.Annotation{
						anno.Anno(f, anno.Annotation{
							ContextStart: mc.Position,
							Start:        cmp.Name.Position,
							Len:          len(cmp.Name.Ident),
							Annotation:   "you first set `" + cmp.Name.Ident + "` here",
						}),
					},
					Suggestions: []corgierr.Suggestion{
						{Suggestion: "you should only set the arg once"},
					},
				})
			}
		}
	}

	return &errs
}

// make sure that a mixin call fills all required args.
func _requiredMixinCallAttributes(f *file.File, mc file.MixinCall) *errList {
	var errs errList

params:
	for _, param := range mc.Mixin.Mixin.Params {
		if param.Default != nil {
			continue
		}

		for _, arg := range mc.Args {
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
				Start:      mc.Position,
				Len:        len("+") + (mc.Name.Col - mc.Col) + len(mc.Name.Ident),
				Annotation: "you need to set `" + param.Name.Ident + "` to call this mixin",
			}),
		})
	}

	return &errs
}

// mixin call body only contains for, if, if block, switch, &s, attribute mixin
// calls, and top-level blocks.
func _mixinCallBody(f *file.File, mc file.MixinCall) *errList {
	var errs errList

	if len(mc.Body) == 1 {
		switch mc.Body[0].(type) {
		case file.BlockExpansion:
			return &errs
		case file.InlineText:
			return &errs
		case file.MixinMainBlockShorthand:
			return &errs
		}
	}

	fileutil.Walk(mc.Body, func(parents []fileutil.WalkContext, ctx fileutil.WalkContext) (dive bool, err error) {
		switch itm := (*ctx.Item).(type) {
		case file.CorgiComment:
			return true, nil
		case file.If:
			return true, nil
		case file.IfBlock:
			return true, nil
		case file.Switch:
			return true, nil
		case file.For:
			return true, nil
		case file.And:
			if mc.Mixin.Mixin.HasAndPlaceholders {
				return true, nil
			}

			for _, b := range mc.Mixin.Mixin.Blocks {
				if b.DefaultTopLevelAndPlaceholder {
					return true, nil
				}
			}

			errs.PushBack(&corgierr.Error{
				Message: "use of & in mixin call to mixin without &-placeholder",
				ErrorAnnotation: anno.Anno(f, anno.Annotation{
					Start:      itm.Position,
					Annotation: "no element to place these attributes on",
				}),
				HintAnnotations: []corgierr.Annotation{inThisMixinCall(f, mc)},
				Suggestions: []corgierr.Suggestion{
					{Suggestion: "remove these attributes or add an &-placeholder to the mixin"},
				},
			})
			return true, nil
		case file.Block:
			if len(parents) == 0 {
				return false, nil
			}

		parents:
			for _, parent := range parents {
				switch (*parent.Item).(type) {
				case file.If:
				case file.IfBlock:
				case file.Switch:
				case file.For:
					continue parents
				default:
					continue parents
				}

				errs.PushBack(&corgierr.Error{
					Message: "conditional mixin call block",
					ErrorAnnotation: anno.Anno(f, anno.Annotation{
						ContextStartDelta: -1,
						Start:             itm.Position,
						Len:               (itm.Name.Col - itm.Col) + len(itm.Name.Ident),
						Annotation:        "you cannot set a mixin call block conditionally",
					}),
					HintAnnotations: []corgierr.Annotation{inThisMixinCall(f, mc)},
					Suggestions: []corgierr.Suggestion{
						{Suggestion: "put the conditional inside the block"},
					},
				})
				return false, nil
			}

			return false, nil
		default:
			errs.PushBack(&corgierr.Error{
				Message: fmt.Sprintf("unexpected item %T in mixin call", itm),
				ErrorAnnotation: anno.Anno(f, anno.Annotation{
					Start:      itm.Pos(),
					ToEOL:      true,
					Annotation: "cannot place this inside a mixin call",
				}),
				HintAnnotations: []corgierr.Annotation{inThisMixinCall(f, mc)},
			})
			return false, nil
		}
	})

	return &errs
}

func _mixinCallBlocksExist(f *file.File, mc file.MixinCall) *errList {
	if len(mc.Body) == 1 {
		if sh, ok := mc.Body[0].(file.MixinMainBlockShorthand); ok {
			for _, block := range mc.Mixin.Mixin.Blocks {
				if block.Name == "_" {
					return &errList{}
				}
			}

			return list.List1(&corgierr.Error{
				Message: "unknown block in mixin call",
				ErrorAnnotation: anno.Anno(f, anno.Annotation{
					Start:      sh.Position,
					Annotation: "this mixin has no `_` block, so you can't use a shorthand",
				}),
				HintAnnotations: []corgierr.Annotation{inThisMixinCall(f, mc)},
			})
		}
	}

	var errs errList

body:
	for _, itm := range mc.Body {
		block, ok := itm.(file.Block)
		if !ok {
			continue
		}

		for _, mblock := range mc.Mixin.Mixin.Blocks {
			if block.Name.Ident == mblock.Name {
				continue body
			}
		}

		errs.PushBack(&corgierr.Error{
			Message: "unknown block in mixin call",
			ErrorAnnotation: anno.Anno(f, anno.Annotation{
				Start:      block.Position,
				Len:        (block.Name.Col - block.Col) + len(block.Name.Ident),
				Annotation: "the mixin you are calling has no block named `" + block.Name.Ident + "`",
			}),
			HintAnnotations: []corgierr.Annotation{inThisMixinCall(f, mc)},
		})
	}

	return &errs
}

func _duplicateMixinCallBlocks(f *file.File, mc file.MixinCall) *errList {
	var errs errList

	var foundBlocks list.List[file.Block]

	for _, itm := range mc.Body {
		block, ok := itm.(file.Block)
		if !ok {
			continue
		}

		for otherE := foundBlocks.Front(); otherE != nil; otherE = otherE.Next() {
			if block.Name.Ident == otherE.V().Name.Ident {
				errs.PushBack(&corgierr.Error{
					Message: "mixin call block filled twice",
					ErrorAnnotation: anno.Anno(f, anno.Annotation{
						Start:      block.Position,
						Len:        (block.Name.Col - block.Col) + len(block.Name.Ident),
						Annotation: "and then here again",
					}),
					HintAnnotations: []corgierr.Annotation{
						anno.Anno(f, anno.Annotation{
							Start:      otherE.V().Position,
							Len:        (otherE.V().Name.Col - otherE.V().Col) + len(otherE.V().Name.Ident),
							Annotation: "block `" + block.Name.Ident + "` first filled here",
						}),
						inThisMixinCall(f, mc),
					},
				})
			}
		}

		foundBlocks.PushBack(block)
	}

	return &errs
}

// check that only mixin blocks that can contain attrs, actually contain attrs.
func _mixinCallBlockAttrs(f *file.File, mc file.MixinCall) *errList {
	if len(mc.Body) == 1 {
		if sh, ok := mc.Body[0].(file.MixinMainBlockShorthand); ok {
			for _, block := range mc.Mixin.Mixin.Blocks {
				if block.Name != "_" {
					continue
				}

				if block.CanAttributes {
					return &errList{}
				}

				attr, ok := fileutil.IsFirstNonControlAttr(sh.Body)
				if !ok {
					return &errList{}
				}
				switch attr := attr.(type) {
				case file.And:
					return list.List1(&corgierr.Error{
						Message: "top-level attribute in mixin call block that doesn't allow top-level attributes",
						ErrorAnnotation: anno.Anno(f, anno.Annotation{
							Start: attr.Position,
							Annotation: "this mixin places the `_` block after it has written to the body of an element\n" +
								" and you can therefore not place any attributes in it",
						}),
						HintAnnotations: []corgierr.Annotation{inThisMixinCall(f, mc)},
					})
				case file.MixinCall:
					return list.List1(&corgierr.Error{
						Message: "top-level attribute in mixin call block that doesn't allow top-level attributes",
						ErrorAnnotation: anno.Anno(f, anno.Annotation{
							Start: attr.Position,
							Len:   len("+") + (mc.Name.Col - mc.Col) + len(mc.Name.Ident),
							Annotation: "the outer mixin places the `_` block after it has written to the body of an element\n" +
								" and you can therefore not place any attributes in it",
						}),
						HintAnnotations: []corgierr.Annotation{inThisMixinCall(f, mc)},
					})
				default:
					panic("shouldn't happen")
				}
			}

			return &errList{}
		}
	}

	var errs errList

body:
	for _, itm := range mc.Body {
		mcBlock, ok := itm.(file.Block)
		if !ok {
			continue
		}

		for _, block := range mc.Mixin.Mixin.Blocks {
			if block.Name != mcBlock.Name.Ident {
				continue
			}

			if block.CanAttributes {
				continue body
			}

			attr, ok := fileutil.IsFirstNonControlAttr(mcBlock.Body)
			if !ok {
				continue body
			}
			switch attr := attr.(type) {
			case file.And:
				errs.PushBack(&corgierr.Error{
					Message: "top-level attribute in mixin call block that doesn't allow top-level attributes",
					ErrorAnnotation: anno.Anno(f, anno.Annotation{
						Start: attr.Position,
						Annotation: "this mixin places the `" + mcBlock.Name.Ident + "` block after it has written\n" +
							"to the body of an element and you can therefore not place any attributes in it",
					}),
					HintAnnotations: []corgierr.Annotation{inThisMixinCall(f, mc)},
				})
			case file.MixinCall:
				errs.PushBack(&corgierr.Error{
					Message: "top-level attribute in mixin call block that doesn't allow top-level attributes",
					ErrorAnnotation: anno.Anno(f, anno.Annotation{
						Start: attr.Position,
						Len:   len("+") + (mc.Name.Col - mc.Col) + len(mc.Name.Ident),
						Annotation: "the outer mixin places the `" + mcBlock.Name.Ident + "` block after it has written\n" +
							"to the body of an element and you can therefore not place any attributes in it",
					}),
					HintAnnotations: []corgierr.Annotation{inThisMixinCall(f, mc)},
				})
			default:
				panic("shouldn't happen")
			}
			continue body
		}
	}

	return &errs
}

func andPlaceholderPlacement(f *file.File) *errList {
	var errs errList

	fileutil.Walk(f.Scope, func(parents []fileutil.WalkContext, ctx fileutil.WalkContext) (dive bool, err error) {
		var ap *file.AndPlaceholder
		switch itm := (*ctx.Item).(type) {
		case file.Element:
			ap = getAndPlaceholder(itm.Attributes)
		case file.DivShorthand:
			ap = getAndPlaceholder(itm.Attributes)
		case file.And:
			ap = getAndPlaceholder(itm.Attributes)
		default:
			return true, nil
		}

		if ap == nil {
			return true, nil
		}

		for _, parent := range parents {
			if _, ok := (*parent.Item).(file.Mixin); ok {
				return true, nil
			}
		}

		errs.PushBack(&corgierr.Error{
			Message: "&-placeholder used outside of mixin",
			ErrorAnnotation: anno.Anno(f, anno.Annotation{
				Start:      ap.Position,
				Len:        2,
				Annotation: "you can only place &-placeholders inside mixins",
			}),
		})

		return true, nil
	})

	return &errs
}
