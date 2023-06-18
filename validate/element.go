package validate

import (
	"fmt"

	"github.com/mavolin/corgi/corgierr"
	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/fileutil"
	"github.com/mavolin/corgi/internal/anno"
)

func topLevelAttribute(f *file.File) errList {
	if f.Extend != nil || (f.Type != file.TypeMain && f.Type != file.TypeExtend && f.Type != file.TypeUse) {
		return errList{}
	}

	return _topLevelAttributes(f, f.Scope)
}

func _topLevelAttributes(f *file.File, s file.Scope) errList {
	var errs errList

	fileutil.Walk(s, func(parents []fileutil.WalkContext, ctx fileutil.WalkContext) (dive bool, err error) {
		switch itm := (*ctx.Item).(type) {
		case file.And:
			errs.PushBack(&corgierr.Error{
				Message: "top-level attribute",
				ErrorAnnotation: anno.Anno(f, anno.Annotation{
					Start:      itm.Position,
					Annotation: "attributes cannot be placed outside of elements",
				}),
				Suggestions: []corgierr.Suggestion{
					{Suggestion: "place this `&` inside an element or remove it"},
				},
			})
			return false, nil
		case file.If:
			return true, nil
		case file.Switch:
			return true, nil
		case file.For:
			return true, nil
		case file.MixinCall:
			if fileutil.IsAttrMixin(*itm.Mixin) {
				errs.PushBack(&corgierr.Error{
					Message: "top-level attribute",
					ErrorAnnotation: anno.Anno(f, anno.Annotation{
						Start:      itm.Position,
						Annotation: "attributes cannot be placed outside of elements",
					}),
					Suggestions: []corgierr.Suggestion{
						{Suggestion: "place this `html.Attribute` call inside an element or remove it"},
					},
				})
				return false, nil
			}

			annoLen := len("+")
			if itm.Namespace != nil {
				annoLen += len(itm.Namespace.Ident) + len(".")
			}
			annoLen += len(itm.Name.Ident)

			if itm.Mixin.WritesTopLevelAttributes {
				errs.PushBack(&corgierr.Error{
					Message: "top-level `&`",
					ErrorAnnotation: anno.Anno(f, anno.Annotation{
						Start: itm.Position,
						Len:   annoLen,
						Annotation: "the mixin being called has top-level attributes,\n" +
							"and must therefore be placed inside an element",
					}),
					Suggestions: []corgierr.Suggestion{
						{Suggestion: "place this mixin call inside an element or remove it"},
					},
				})
				return false, nil // only report one err per mixin
			}

			andPos := mixinCallAttrPos(itm)

			if andPos != file.InvalidPosition && itm.Mixin.HasAndPlaceholders {
				errs.PushBack(&corgierr.Error{
					Message: "top-level `&`",
					ErrorAnnotation: anno.Anno(f, anno.Annotation{
						Start:      itm.Position,
						Len:        annoLen,
						Annotation: "the mixin being called has a top-level &-placeholder (`&(&&)`)",
					}),
					HintAnnotations: []corgierr.Annotation{
						anno.Anno(f, anno.Annotation{
							Start:      andPos,
							Annotation: "and you are providing attributes for it here",
						}),
					},
					Suggestions: []corgierr.Suggestion{
						{Suggestion: "place this mixin call inside an element or remove it"},
					},
				})
				return false, nil // only report one err per mixin
			}

			unfilledBlocks := make([]file.LinkedMixinBlock, 0, len(itm.Mixin.Blocks))
			for _, block := range itm.Mixin.Blocks {
				if block.TopLevel && block.CanAttributes {
					unfilledBlocks = append(unfilledBlocks, block)
				}
			}

			if len(itm.Body) == 1 {
				if sh, ok := itm.Body[0].(file.MixinMainBlockShorthand); ok {
					for i, ublock := range unfilledBlocks {
						if ublock.Name == "_" {
							blockErrs := _topLevelAttributes(f, sh.Body)
							if blockErrs.Len() > 0 {
								errs.PushBackList(&blockErrs)
								return false, nil // only report one err per mixin
							}

							copy(unfilledBlocks[i:], unfilledBlocks[i+1:])
							unfilledBlocks = unfilledBlocks[:len(unfilledBlocks)-1]
							goto handleUnfilledBlocks
						}
					}

					goto handleUnfilledBlocks
				}
			}

		body:
			for _, itm := range itm.Body {
				block, ok := itm.(file.Block)
				if !ok {
					continue
				}

				for i, ublock := range unfilledBlocks {
					if block.Name.Ident == ublock.Name {
						blockErrs := _topLevelAttributes(f, block.Body)
						if blockErrs.Len() > 0 {
							errs.PushBackList(&blockErrs)
							return false, nil // only report one err per mixin
						}

						copy(unfilledBlocks[i:], unfilledBlocks[i+1:])
						unfilledBlocks = unfilledBlocks[:len(unfilledBlocks)-1]
						continue body
					}
				}
			}

		handleUnfilledBlocks:
			for _, ublock := range unfilledBlocks {
				if andPos != file.InvalidPosition && ublock.DefaultTopLevelAndPlaceholder {
					errs.PushBack(&corgierr.Error{
						Message: "top-level `&`",
						ErrorAnnotation: anno.Anno(f, anno.Annotation{
							Start: itm.Position,
							Len:   annoLen,
							Annotation: "The mixin you are calling has a top-level block named `" + ublock.Name + "`\n" +
								"which you didn't fill and whose default has a top-level and placeholder (`&(&&)`).",
						}),
						HintAnnotations: []corgierr.Annotation{
							anno.Anno(f, anno.Annotation{
								Start:      andPos,
								Annotation: "and you are providing attributes for it here",
							}),
						},
						Suggestions: []corgierr.Suggestion{
							{
								Suggestion: "you are not allowed to place attributes outside of elements, therefore,\n" +
									"place this mixin call inside an element,\n" +
									"manually set `" + ublock.Name + "` without using top-level ands, or remove this mixin call",
							},
						},
					})
					return false, nil // only report one err per mixin
				} else if ublock.DefaultWritesTopLevelAttributes {
					errs.PushBack(&corgierr.Error{
						Message: "top-level attributes in top-level block default",
						ErrorAnnotation: anno.Anno(f, anno.Annotation{
							Start: itm.Position,
							Len:   annoLen,
							Annotation: "The mixin you are calling has a block named `" + ublock.Name + "`\n" +
								"which you didn't fill and whose default has one or more top-level attributes.\n" +
								"Since this mixin call is not placed inside an element, you cannot use top-level attributes.",
						}),
						Suggestions: []corgierr.Suggestion{
							{
								Suggestion: "place this mixin call inside an element,\n" +
									"manually set `" + ublock.Name + "` without using top-level attributes, or remove this mixin call",
							},
						},
					})
					// continue, as there could be more blocks like this
				}
			}

			return false, nil
		default:
			return false, nil
		}
	})
	return errs
}

func topLevelTemplateBlockAnds(f *file.File) errList {
	if f.Extend == nil {
		return errList{}
	}

	var errs errList

	fileutil.Walk(f.Scope, func(parents []fileutil.WalkContext, ctx fileutil.WalkContext) (dive bool, err error) {
		block, ok := (*ctx.Item).(file.Block)
		if !ok {
			return false, nil
		}

		blockErrs := _topLevelTemplateBlockAnds(f, block.Body)
		errs.PushBackList(&blockErrs)
		return false, nil
	})

	return errs
}

func _topLevelTemplateBlockAnds(f *file.File, s file.Scope) errList {
	var errs errList

	fileutil.Walk(s, func(parents []fileutil.WalkContext, ctx fileutil.WalkContext) (dive bool, err error) {
		switch itm := (*ctx.Item).(type) {
		case file.And:
			errs.PushBack(&corgierr.Error{
				Message: "top-level `&` in block",
				ErrorAnnotation: anno.Anno(f, anno.Annotation{
					Start:      itm.Position,
					Annotation: "attributes may not be placed at the top level of a template block",
				}),
			})
			return false, nil
		case file.If:
			return true, nil
		case file.Switch:
			return true, nil
		case file.For:
			return true, nil
		case file.MixinCall:
			if itm.Mixin.File.Module == "" && itm.Mixin.File.ModulePath == "html" && itm.Name.Ident == "Attr" {
				errs.PushBack(&corgierr.Error{
					Message: "top-level attribute",
					ErrorAnnotation: anno.Anno(f, anno.Annotation{
						Start:      itm.Position,
						Annotation: "attributes cannot be placed outside of elements",
					}),
					Suggestions: []corgierr.Suggestion{
						{Suggestion: "place this `html.Attribute` call inside an element or remove it"},
					},
				})
				return false, nil
			}

			annoLen := len("+")
			if itm.Namespace != nil {
				annoLen += len(itm.Namespace.Ident) + len(".")
			}
			annoLen += len(itm.Name.Ident)

			if itm.Mixin.WritesTopLevelAttributes {
				errs.PushBack(&corgierr.Error{
					Message: "top-level `&`",
					ErrorAnnotation: anno.Anno(f, anno.Annotation{
						Start: itm.Position,
						Len:   annoLen,
						Annotation: "the mixin being called has top-level attributes,\n" +
							"and must therefore be placed inside an element",
					}),
					Suggestions: []corgierr.Suggestion{
						{Suggestion: "place this mixin call inside an element or remove it"},
					},
				})
				return false, nil // only report one err per mixin
			}

			andPos := mixinCallAttrPos(itm)

			if andPos != file.InvalidPosition && itm.Mixin.HasAndPlaceholders {
				errs.PushBack(&corgierr.Error{
					Message: "top-level `&`",
					ErrorAnnotation: anno.Anno(f, anno.Annotation{
						Start:      itm.Position,
						Len:        annoLen,
						Annotation: "the mixin being called has a top-level &-placeholder (`&(&&)`)",
					}),
					HintAnnotations: []corgierr.Annotation{
						anno.Anno(f, anno.Annotation{
							Start:      andPos,
							Annotation: "and you are providing attributes for it here",
						}),
					},
					Suggestions: []corgierr.Suggestion{
						{Suggestion: "place this mixin call inside an element or remove it"},
					},
				})
				return false, nil // only report one err per mixin
			}

			unfilledBlocks := make([]file.LinkedMixinBlock, 0, len(itm.Mixin.Blocks))
			for _, block := range itm.Mixin.Blocks {
				if block.TopLevel && block.CanAttributes {
					unfilledBlocks = append(unfilledBlocks, block)
				}
			}

			if len(itm.Body) == 1 {
				if sh, ok := itm.Body[0].(file.MixinMainBlockShorthand); ok {
					for i, ublock := range unfilledBlocks {
						if ublock.Name == "_" {
							blockErrs := _topLevelAttributes(f, sh.Body)
							if blockErrs.Len() > 0 {
								errs.PushBackList(&blockErrs)
								return false, nil // only report one err per mixin
							}

							copy(unfilledBlocks[i:], unfilledBlocks[i+1:])
							unfilledBlocks = unfilledBlocks[:len(unfilledBlocks)-1]
							goto handleUnfilledBlocks
						}
					}

					goto handleUnfilledBlocks
				}
			}

		body:
			for _, itm := range itm.Body {
				block, ok := itm.(file.Block)
				if !ok {
					continue
				}

				for i, ublock := range unfilledBlocks {
					if block.Name.Ident == ublock.Name {
						blockErrs := _topLevelAttributes(f, block.Body)
						if blockErrs.Len() > 0 {
							errs.PushBackList(&blockErrs)
							return false, nil // only report one err per mixin
						}

						copy(unfilledBlocks[i:], unfilledBlocks[i+1:])
						unfilledBlocks = unfilledBlocks[:len(unfilledBlocks)-1]
						continue body
					}
				}
			}

		handleUnfilledBlocks:
			for _, ublock := range unfilledBlocks {
				if andPos != file.InvalidPosition && ublock.DefaultTopLevelAndPlaceholder {
					errs.PushBack(&corgierr.Error{
						Message: "top-level `&`",
						ErrorAnnotation: anno.Anno(f, anno.Annotation{
							Start: itm.Position,
							Len:   annoLen,
							Annotation: "The mixin you are calling has a top-level block named `" + ublock.Name + "`\n" +
								"which you didn't fill and whose default has a top-level and placeholder (`&(&&)`).",
						}),
						HintAnnotations: []corgierr.Annotation{
							anno.Anno(f, anno.Annotation{
								Start:      andPos,
								Annotation: "and you are providing attributes for it here",
							}),
						},
						Suggestions: []corgierr.Suggestion{
							{
								Suggestion: "you are not allowed to place attributes outside of elements, therefore,\n" +
									"place this mixin call inside an element,\n" +
									"manually set `" + ublock.Name + "` without using top-level ands, or remove this mixin call",
							},
						},
					})
					return false, nil // only report one err per mixin
				} else if ublock.DefaultWritesTopLevelAttributes {
					errs.PushBack(&corgierr.Error{
						Message: "top-level `&` in top-level block default",
						ErrorAnnotation: anno.Anno(f, anno.Annotation{
							Start: itm.Position,
							Len:   annoLen,
							Annotation: "The mixin you are calling has a block named `" + ublock.Name + "`\n" +
								"which you didn't fill and whose default has one or more top-level attributes.\n" +
								"Since this mixin call is not placed inside an element, you cannot use top-level attributes.",
						}),
						Suggestions: []corgierr.Suggestion{
							{
								Suggestion: "place this mixin call inside an element,\n" +
									"manually set `" + ublock.Name + "` without using top-level ands, or remove this mixin call",
							},
						},
					})
					// continue, as there could be more blocks like this
				}
			}

			return false, nil
		default:
			return false, nil
		}
	})
	return errs
}

// attributePlacement checks that & directives and `html.Attr` calls are placed
// according to the following rules:
//
//   - an attr must be placed before writing to the body of the element, i.e.
//     before writing any text or other elements.
//   - an attr may be placed inside a conditional
//   - to be able to use further attrs after a conditional, all branches of
//     that
//   - conditional must also fulfill these rules
//   - attrs must not be placed after blocks -- although resolvable at
//     compile-time, it is intransparent, as blocks are filled elsewhere from
//     their placeholder
//   - attrs may be placed after mixin calls, if neither the mixin nor its
//     blocks write to the element's body
func attributePlacement(f *file.File) errList {
	var errs errList

	fileutil.Walk(f.Scope, func(parents []fileutil.WalkContext, ctx fileutil.WalkContext) (dive bool, err error) {
		switch itm := (*ctx.Item).(type) {
		case file.Element:
			elAnno := anno.Anno(f, anno.Annotation{
				Start:      itm.Position,
				Len:        len(itm.Name),
				Annotation: "in this element",
			})
			_, elErrs := _attributePlacement(f, elAnno, nil, itm.Body)
			elErrs.PushBackList(&elErrs)
		case file.DivShorthand:
			elAnno := anno.Anno(f, anno.Annotation{
				Start:      itm.Position,
				Annotation: "in this div shorthand",
			})
			_, divErrs := _attributePlacement(f, elAnno, nil, itm.Body)
			errs.PushBackList(&divErrs)
		case file.MixinCall:
			if itm.Mixin.File.Module == "" && itm.Mixin.File.ModulePath == "html" && itm.Name.Ident == "Element" {
				var end file.Position
				if itm.RParenPos != nil {
					end = *itm.RParenPos
				}
				elAnno := anno.Anno(f, anno.Annotation{
					Start:      itm.Position,
					End:        end,
					Annotation: "in this element",
				})
				_, mcErrs := _attributePlacement(f, elAnno, nil, itm.Body)
				errs.PushBackList(&mcErrs)
			}
		}
		return true, nil
	})

	return errs
}

func _attributePlacement(f *file.File, elAnno corgierr.Annotation, firstText *corgierr.Annotation, scope file.Scope) (*corgierr.Annotation, errList) {
	var errs errList

	fileutil.Walk(scope, func(parents []fileutil.WalkContext, ctx fileutil.WalkContext) (dive bool, err error) {
		switch itm := (*ctx.Item).(type) {
		case file.And:
			if firstText == nil {
				return false, nil
			}

			errs.PushBack(&corgierr.Error{
				Message: "use of attribute after writing to element's body",
				ErrorAnnotation: anno.Anno(f, anno.Annotation{
					Start:      itm.Position,
					Annotation: "so you cannot place an `&` here",
				}),
				HintAnnotations: []corgierr.Annotation{elAnno, *firstText},
				Suggestions: []corgierr.Suggestion{
					{
						Suggestion: "you can only use the `&` operator before you write to the body of an element.",
					},
				},
			})
			return false, nil
		case file.If:
			firstTextAfterIf, ifErrs := _attributePlacement(f, elAnno, firstText, itm.Then)
			errs.PushBackList(&ifErrs)

			for _, elseIf := range itm.ElseIfs {
				firstElseIfText, elseIfErrs := _attributePlacement(f, elAnno, firstText, elseIf.Then)
				errs.PushBackList(&elseIfErrs)
				if firstTextAfterIf == nil {
					firstTextAfterIf = firstElseIfText
				}
			}

			if itm.Else != nil {
				firstElseText, elseErrs := _attributePlacement(f, elAnno, firstText, itm.Else.Then)
				errs.PushBackList(&elseErrs)
				if firstTextAfterIf == nil {
					firstTextAfterIf = firstElseText
				}
			}

			if firstText == nil {
				firstText = firstTextAfterIf
			}
		case file.IfBlock:
			firstTextAfterIf, ifErrs := _attributePlacement(f, elAnno, firstText, itm.Then)
			errs.PushBackList(&ifErrs)

			for _, elseIf := range itm.ElseIfs {
				firstElseIfText, elseIfErrs := _attributePlacement(f, elAnno, firstText, elseIf.Then)
				errs.PushBackList(&elseIfErrs)
				if firstTextAfterIf == nil {
					firstTextAfterIf = firstElseIfText
				}
			}

			if itm.Else != nil {
				firstElseTextItm, elseErrs := _attributePlacement(f, elAnno, firstText, itm.Else.Then)
				errs.PushBackList(&elseErrs)
				if firstTextAfterIf == nil {
					firstTextAfterIf = firstElseTextItm
				}
			}

			if firstText == nil {
				firstText = firstTextAfterIf
			}
		case file.Switch:
			var firstTextAfterSwitch *corgierr.Annotation

			for _, c := range itm.Cases {
				firstCaseText, caseErrs := _attributePlacement(f, elAnno, firstText, c.Then)
				errs.PushBackList(&caseErrs)
				if firstTextAfterSwitch == nil {
					firstTextAfterSwitch = firstCaseText
				}
			}

			if itm.Default != nil {
				firstDefaultText, defaultErrs := _attributePlacement(f, elAnno, firstText, itm.Default.Then)
				errs.PushBackList(&defaultErrs)
				if firstTextAfterSwitch == nil {
					firstTextAfterSwitch = firstDefaultText
				}
			}

			if firstText == nil {
				firstText = firstTextAfterSwitch
			}
		case file.For:
			firstTextAfterFor, forErrs := _attributePlacement(f, elAnno, firstText, itm.Body)
			errs.PushBackList(&forErrs)

			if firstText == nil && firstTextAfterFor != nil {
				if nonCtrl, ok := fileutil.IsFirstNonControlAttr(itm.Body); ok {
					if and, ok := nonCtrl.(file.And); ok {
						errs.PushBack(&corgierr.Error{
							Message: "use of `&` in for-loop that also writes to element's body",
							ErrorAnnotation: anno.Anno(f, anno.Annotation{
								Start:      and.Position,
								Annotation: "you placed an `&` here",
							}),
							HintAnnotations: []corgierr.Annotation{elAnno, *firstTextAfterFor},
							Suggestions: []corgierr.Suggestion{
								{
									Suggestion: "you can only use the `&` in loops, if you don't also write to the body of the element;\n" +
										"consider writing two loops, one for the `&` and one for the rest",
								},
							},
						})
					} else if mixin, ok := nonCtrl.(file.MixinCall); ok {
						errs.PushBack(&corgierr.Error{
							Message: "use of attribute in for-loop that also writes to element's body",
							ErrorAnnotation: anno.Anno(f, anno.Annotation{
								Start:      mixin.Position,
								Annotation: "you placed a mixin writing an attribute (and possibly also text) here",
							}),
							HintAnnotations: []corgierr.Annotation{elAnno, *firstTextAfterFor},
						})
					}
				}

				firstText = firstTextAfterFor
			}
		case file.Element:
			if firstText == nil {
				a := anno.Anno(f, anno.Annotation{
					Start:      itm.Position,
					Len:        len(itm.Name),
					Annotation: "you wrote an element here",
				})
				firstText = &a
			}
		case file.HTMLComment:
			if firstText == nil {
				a := anno.Anno(f, anno.Annotation{
					Start:      itm.Position,
					ToEOL:      true,
					Annotation: "you wrote a html comment here",
				})
				firstText = &a
			}
		case file.DivShorthand:
			if firstText == nil {
				a := anno.Anno(f, anno.Annotation{
					Start:      itm.Position,
					Annotation: "you wrote a div shorthand here",
				})
				firstText = &a
			}
		case file.ArrowBlock:
			if firstText == nil {
				a := anno.Anno(f, anno.Annotation{
					Start:      itm.Position,
					ToEOL:      true,
					Annotation: "you wrote an arrow block here",
				})
				firstText = &a
			}
		case file.Assign:
			if firstText == nil {
				a := anno.Anno(f, anno.Annotation{
					Start:      itm.Position,
					ToEOL:      true,
					Annotation: "you wrote an assign here",
				})
				firstText = &a
			}
		case file.InlineText:
			if firstText == nil {
				a := anno.Anno(f, anno.Annotation{
					Start:      itm.Position,
					ToEOL:      true,
					Annotation: "you wrote inline text here",
				})
				firstText = &a
			}
		case file.Block:
			if firstText == nil {
				a := anno.Anno(f, anno.Annotation{
					Start:      itm.Position,
					Len:        (itm.Name.Col - itm.Position.Col) + len(itm.Name.Ident),
					Annotation: "you used a template block placeholder here, after which you cannot place `&`s",
				})
				firstText = &a
			}
		case file.Include:
			if firstText == nil {
				a := anno.Anno(f, anno.Annotation{
					Start:      itm.Position,
					ToEOL:      true,
					Annotation: "you included another file here, after which you cannot place attributes",
				})
				firstText = &a
			}
		case file.MixinCall:
			if itm.Mixin.File.Module == "" && itm.Mixin.File.ModulePath == "html" {
				var end file.Position
				if itm.RParenPos != nil {
					end = *itm.RParenPos
				}

				switch itm.Name.Ident {
				case "Element":
					if firstText == nil {
						a := anno.Anno(f, anno.Annotation{
							Start:      itm.Position,
							End:        end,
							Annotation: "you wrote an element here",
						})
						firstText = &a
					}
					return false, nil
				case "Attr":
					if firstText == nil {
						return false, nil
					}

					errs.PushBack(&corgierr.Error{
						Message: "use of attribute after writing to element's body",
						ErrorAnnotation: anno.Anno(f, anno.Annotation{
							Start:      itm.Position,
							End:        end,
							Annotation: "so you cannot place an attribute here",
						}),
						HintAnnotations: []corgierr.Annotation{elAnno, *firstText},
						Suggestions: []corgierr.Suggestion{
							{
								Suggestion: "you can only use the `&` operator before you write to the body of an element.",
							},
						},
					})
					return false, nil
				}
			}

			annoLen := len("+")
			if itm.Namespace != nil {
				annoLen += len(itm.Namespace.Ident) + len(".")
			}
			annoLen += len(itm.Name.Ident)

			mixinErrs := _mixinCallAndPlacement(f, itm, elAnno, firstText)
			errs.PushBackList(&mixinErrs)

			mcFirstText := _mixinCallFirstTextAnno(f, itm)
			if mcFirstText != nil {
				firstText = mcFirstText
			}
		}
		return false, nil
	})

	return firstText, errs
}

// _mixinCallAndPlacement checks whether the mixin call can be placed, or if it
// attempts to write attributes to the element containing it, even though the
// elements body has already been written.
func _mixinCallAndPlacement(f *file.File, mc file.MixinCall, elAnno corgierr.Annotation, firstText *corgierr.Annotation) errList {
	var errs errList

	if firstText == nil {
		return errList{}
	}

	annoLen := len("+")
	if mc.Namespace != nil {
		annoLen += len(mc.Namespace.Ident) + len(".")
	}
	annoLen += len(mc.Name.Ident)

	if mc.Mixin.WritesTopLevelAttributes {
		errs.PushBack(&corgierr.Error{
			Message: "use of `&` after writing to element's body",
			ErrorAnnotation: anno.Anno(f, anno.Annotation{
				Start:      mc.Position,
				Len:        annoLen,
				Annotation: "this mixin writes attributes to the element it is called in",
			}),
			HintAnnotations: []corgierr.Annotation{elAnno, *firstText},
			Suggestions: []corgierr.Suggestion{
				{
					Suggestion: "you can only use the `&` operator before you write to the body of an element.",
				},
			},
		})
		return errs // only one attr err per mixin call
	}

	andPos := mixinCallAttrPos(mc)
	if andPos != file.InvalidPosition && mc.Mixin.TopLevelAndPlaceholder {
		errs.PushBack(&corgierr.Error{
			Message: "use of `&` after writing to element's body",
			ErrorAnnotation: anno.Anno(f, anno.Annotation{
				Start:      mc.Position,
				Len:        annoLen,
				Annotation: "the mixin being called has a top-level &-placeholder (`&(&&)`)",
			}),
			HintAnnotations: []corgierr.Annotation{
				elAnno,
				anno.Anno(f, anno.Annotation{
					Start:      andPos,
					Annotation: "and you are providing attributes for it here",
				}),
			},
			Suggestions: []corgierr.Suggestion{
				{
					Suggestion: "you can only use the `&` operator before you write to the body of an element",
				},
			},
		})
		return errs // only one attr err per mixin call
	}

	unfilledBlocks := make([]file.LinkedMixinBlock, 0, len(mc.Mixin.Blocks))
	for _, block := range mc.Mixin.Blocks {
		if block.TopLevel && block.CanAttributes {
			unfilledBlocks = append(unfilledBlocks, block)
		}
	}

	if len(mc.Body) == 1 {
		if sh, ok := mc.Body[0].(file.MixinMainBlockShorthand); ok {
			for i, ublock := range unfilledBlocks {
				if ublock.Name == "_" {
					blockErrs := _topLevelAttributes(f, sh.Body)
					if blockErrs.Len() > 0 {
						errs.PushBackList(&blockErrs)
						return errs // only report one err per mixin
					}

					copy(unfilledBlocks[i:], unfilledBlocks[i+1:])
					unfilledBlocks = unfilledBlocks[:len(unfilledBlocks)-1]
					goto handleUnfilledBlocks
				}
			}

			goto handleUnfilledBlocks
		}
	}

body:
	for _, itm := range mc.Body {
		block, ok := itm.(file.Block)
		if !ok {
			continue
		}

		for i, ublock := range unfilledBlocks {
			if block.Name.Ident == ublock.Name {
				blockErrs := _topLevelAttributes(f, block.Body)
				if blockErrs.Len() > 0 {
					errs.PushBackList(&blockErrs)
					return errs // only report one err per mixin
				}

				copy(unfilledBlocks[i:], unfilledBlocks[i+1:])
				unfilledBlocks = unfilledBlocks[:len(unfilledBlocks)-1]
				continue body
			}
		}
	}

handleUnfilledBlocks:
	for _, ublock := range unfilledBlocks {
		if andPos != file.InvalidPosition && ublock.DefaultTopLevelAndPlaceholder {
			errs.PushBack(&corgierr.Error{
				Message: "use of `&` after writing to element's body",
				ErrorAnnotation: anno.Anno(f, anno.Annotation{
					Start: mc.Position,
					Len:   annoLen,
					Annotation: "The mixin you are calling has a top-level block named `" + ublock.Name + "`\n" +
						"which you didn't fill and whose default has a top-level and placeholder (`&(&&)`).",
				}),
				HintAnnotations: []corgierr.Annotation{
					elAnno,
					anno.Anno(f, anno.Annotation{
						Start: andPos,
						Annotation: "and you are providing attributes for it here,\n" +
							"even though you already wrote to the body of the element",
					}),
				},
				Suggestions: []corgierr.Suggestion{
					{
						Suggestion: fmt.Sprintf("you can only use the `&` operator before you write to the body of an element,\n"+
							"therefore, place this mixin call before line %d, remove all attributes from the mixin call,\n"+
							"or overwrite the block default", firstText.Line),
					},
				},
			})
			return errs // only one attr err per mixin call
		} else if ublock.DefaultWritesTopLevelAttributes {
			errs.PushBack(&corgierr.Error{
				Message: "top-level `&` in top-level block default",
				ErrorAnnotation: anno.Anno(f, anno.Annotation{
					Start: mc.Position,
					Len:   annoLen,
					Annotation: "The mixin you are calling has a block named `" + ublock.Name + "`\n" +
						"which you didn't fill and whose default has one or more top-level `&`s.\n" +
						"Since this mixin call is not placed inside an element, you cannot use top-level `&`.",
				}),
				HintAnnotations: []corgierr.Annotation{elAnno},
				Suggestions: []corgierr.Suggestion{
					{
						Suggestion: "place this mixin call inside an element,\n" +
							"manually set `" + ublock.Name + "` without using top-level ands, or remove this mixin call",
					},
				},
			})
			// continue, as there could be more blocks like this
		}
	}

	return errs
}

// _mixinCallFirstTextAnno returns the first text annotation for the passed
// mixin call.
func _mixinCallFirstTextAnno(f *file.File, mc file.MixinCall) *corgierr.Annotation {
	if mc.Mixin.WritesBody {
		annoLen := len("+")
		if mc.Namespace != nil {
			annoLen += len(mc.Namespace.Ident) + len(".")
		}
		annoLen += len(mc.Name.Ident)

		a := anno.Anno(f, anno.Annotation{
			Start:      mc.Position,
			Len:        annoLen,
			Annotation: "you used a mixin, that writes to the element here",
		})
		return &a
	}

	unfilledBlocks := make([]file.LinkedMixinBlock, 0, len(mc.Mixin.Blocks))
	for _, block := range mc.Mixin.Blocks {
		if block.TopLevel {
			unfilledBlocks = append(unfilledBlocks, block)
		}
	}

	for _, itm := range mc.Body {
		block, ok := itm.(file.Block)
		if !ok {
			continue
		}

		for i, ublock := range unfilledBlocks {
			if ublock.Name == block.Name.Ident {
				if firstText := _firstTextAnno(f, block.Body); firstText != nil {
					return firstText
				}
			}
			copy(unfilledBlocks[i:], unfilledBlocks[i+1:])
			unfilledBlocks = unfilledBlocks[i:]
		}
	}

	for _, ublock := range unfilledBlocks {
		annoLen := len("+")
		if mc.Namespace != nil {
			annoLen += len(mc.Namespace.Ident) + len(".")
		}
		annoLen += len(mc.Name.Ident)

		if ublock.DefaultWritesBody {
			a := anno.Anno(f, anno.Annotation{
				Start:      mc.Position,
				Len:        annoLen,
				Annotation: "you used a mixin, whose default block `" + ublock.Name + "` writes to the element's body",
			})
			return &a
		}
	}

	return nil
}

func _firstTextAnno(f *file.File, s file.Scope) *corgierr.Annotation {
	var firstText *corgierr.Annotation
	fileutil.Walk(s, func(parents []fileutil.WalkContext, ctx fileutil.WalkContext) (dive bool, err error) {
		switch itm := (*ctx.Item).(type) {
		case file.Element:
			a := anno.Anno(f, anno.Annotation{
				Start:      itm.Position,
				Len:        len(itm.Name),
				Annotation: "you wrote an element here",
			})
			firstText = &a
			return false, fileutil.StopWalk
		case file.HTMLComment:
			a := anno.Anno(f, anno.Annotation{
				Start:      itm.Position,
				ToEOL:      true,
				Annotation: "you wrote a html comment here",
			})
			firstText = &a
			return false, fileutil.StopWalk
		case file.DivShorthand:
			a := anno.Anno(f, anno.Annotation{
				Start:      itm.Position,
				Annotation: "you wrote a div shorthand here",
			})
			firstText = &a
			return false, fileutil.StopWalk
		case file.ArrowBlock:
			a := anno.Anno(f, anno.Annotation{
				Start:      itm.Position,
				ToEOL:      true,
				Annotation: "you wrote an arrow block here",
			})
			firstText = &a
			return false, fileutil.StopWalk
		case file.Assign:
			a := anno.Anno(f, anno.Annotation{
				Start:      itm.Position,
				ToEOL:      true,
				Annotation: "you wrote an assign here",
			})
			firstText = &a
			return false, fileutil.StopWalk
		case file.InlineText:
			a := anno.Anno(f, anno.Annotation{
				Start:      itm.Position,
				ToEOL:      true,
				Annotation: "you wrote inline text here",
			})
			firstText = &a
			return false, fileutil.StopWalk
		case file.Block:
			a := anno.Anno(f, anno.Annotation{
				Start:      itm.Position,
				Len:        (itm.Name.Col - itm.Position.Col) + len(itm.Name.Ident),
				Annotation: "you used a template block placeholder here, after which you cannot place attributes",
			})
			firstText = &a
			return false, fileutil.StopWalk
		case file.Include:
			a := anno.Anno(f, anno.Annotation{
				Start:      itm.Position,
				ToEOL:      true,
				Annotation: "you included another file here, after which you cannot place attributes",
			})
			firstText = &a
			return false, fileutil.StopWalk
		case file.MixinCall:
			if itm.Mixin.File.Module == "" && itm.Mixin.File.ModulePath == "html" && itm.Name.Ident == "Element" {
				var end file.Position
				if itm.RParenPos != nil {
					end = *itm.RParenPos
				}

				a := anno.Anno(f, anno.Annotation{
					Start:      itm.Position,
					End:        end,
					Annotation: "you wrote an element here",
				})
				firstText = &a
				return false, fileutil.StopWalk
			}

			mcFirstText := _mixinCallFirstTextAnno(f, itm)
			if mcFirstText != nil {
				firstText = mcFirstText
				return false, fileutil.StopWalk
			}

			return false, nil
		default:
			return true, nil
		}
	})

	return firstText
}
