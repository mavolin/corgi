package validate

import (
	"github.com/mavolin/corgi/corgierr"
	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/fileutil"
	"github.com/mavolin/corgi/internal/anno"
	"github.com/mavolin/corgi/internal/list"
	"github.com/mavolin/corgi/internal/stack"
)

func mixinChecks(f *file.File) *errList {
	var errs errList

	fileutil.Walk(f.Scope, func(parents []fileutil.WalkContext, ctx fileutil.WalkContext) (dive bool, err error) {
		m, ok := (*ctx.Item).(file.Mixin)
		if !ok {
			return true, nil
		}

		errs.PushBackList(_mixinParamsHaveType(f, m))
		errs.PushBackList(_duplicateMixinParams(f, m))

		return true, nil
	})

	return &errs
}

// no mixins may be declared in other mixin declarations.
func mixinsInMixins(f *file.File) *errList {
	var errs errList

	fileutil.Walk(f.Scope, func(parents []fileutil.WalkContext, ctx fileutil.WalkContext) (dive bool, err error) {
		inner, ok := (*ctx.Item).(file.Mixin)
		if !ok {
			return true, nil
		}

		for _, parent := range parents {
			outer, ok := (*parent.Item).(file.Mixin)
			if !ok {
				continue
			}

			errs.PushBack(&corgierr.Error{
				Message: "mixin declared inside other mixin",
				ErrorAnnotation: anno.Anno(f, anno.Annotation{
					Start:      inner.Position,
					Len:        (inner.Name.Col - inner.Col) + len(inner.Name.Ident),
					Annotation: "cannot declare mixin here",
				}),
				HintAnnotations: []corgierr.Annotation{
					anno.Anno(f, anno.Annotation{
						Start:      outer.Position,
						Len:        (outer.Name.Col - outer.Col) + len(outer.Name.Ident),
						Annotation: "outer mixin",
					}),
				},
				Suggestions: []corgierr.Suggestion{
					{Suggestion: "declare `" + inner.Name.Ident + "` outside of `" + outer.Name.Ident + "`"},
				},
			})
		}

		return true, nil
	})

	return &errs
}

// duplicate names inside same scope.
func duplicateMixinNames(f *file.File) *errList {
	var errs errList

	var s stack.Stack[*list.List[file.Mixin]]

	fileutil.Walk(f.Scope, func(parents []fileutil.WalkContext, ctx fileutil.WalkContext) (dive bool, err error) {
		if len(parents)+1 > s.Len() {
			s.Push(&list.List[file.Mixin]{})
		} else if len(parents)+1 < s.Len() {
			s.Pop()
		}

		m, ok := (*ctx.Item).(file.Mixin)
		if !ok {
			return false, nil
		}

		scopeMixins := s.Peek()
		for otherE := scopeMixins.Front(); otherE != nil; otherE = otherE.Next() {
			if otherE.V().Name.Ident == m.Name.Ident {
				errs.PushBack(&corgierr.Error{
					Message: "duplicate mixin name within same scope",
					ErrorAnnotation: anno.Anno(f, anno.Annotation{
						Start:      m.Name.Position,
						Len:        len(m.Name.Ident),
						Annotation: "and then here again",
					}),
					HintAnnotations: []corgierr.Annotation{
						anno.Anno(f, anno.Annotation{
							Start:      otherE.V().Name.Position,
							Len:        len(otherE.V().Name.Ident),
							Annotation: "you first used the name here",
						}),
					},
					Suggestions: []corgierr.Suggestion{
						{Suggestion: "rename either of these mixins or delete one of them"},
					},
				})
			}
		}

		s.Peek().PushBack(m)

		return false, nil
	})

	return &errs
}

func _mixinParamsHaveType(f *file.File, m file.Mixin) *errList {
	var errs errList

	for _, param := range m.Params {
		if param.Type != nil || param.InferredType != "" {
			continue
		}

		errs.PushBack(&corgierr.Error{
			Message: "unable to infer type of mixin param",
			ErrorAnnotation: anno.Anno(f, anno.Annotation{
				Start:      param.Name.Position,
				Len:        len(param.Name.Ident),
				Annotation: "this param has no explicit type,\nand no type could be inferred from the default",
			}),
			Suggestions: []corgierr.Suggestion{
				{
					Suggestion: "give this param an explicit type",
					Example:    "`" + param.Name.Ident + " string = ...`",
				},
			},
		})
	}

	return &errs
}

func _duplicateMixinParams(f *file.File, m file.Mixin) *errList {
	if len(m.Params) <= 1 {
		return &errList{}
	}

	var errs errList

	dupls := make([]file.MixinParam, 0, len(m.Params)-1)
	skip := make([]int, 0, len(m.Params)-1)

	for ai, a := range m.Params[:len(m.Params)-1] {
	b:
		for bi, b := range m.Params[ai+1:] {
			for _, skipI := range skip {
				if bi == skipI {
					continue b
				}
			}

			if a.Name.Ident == b.Name.Ident {
				dupls = append(dupls, b)
				skip = append(skip, bi)
			}
		}

		if len(dupls) == 0 {
			continue
		}

		has := make([]corgierr.Annotation, len(dupls))
		for i, dupl := range dupls {
			has[i] = anno.Anno(f, anno.Annotation{
				Start:      dupl.Name.Position,
				Len:        len(dupl.Name.Ident),
				Annotation: "here again",
			})
		}

		errs.PushBack(&corgierr.Error{
			Message: "duplicate mixin parameter",
			ErrorAnnotation: anno.Anno(f, anno.Annotation{
				Start:      a.Name.Position,
				Len:        len(a.Name.Ident),
				Annotation: "this parameter name is used multiple times",
			}),
			HintAnnotations: has,
		})

		dupls = dupls[:0]
	}

	return &errs
}
