package validate

import (
	"github.com/mavolin/corgi/corgierr"
	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/fileutil"
	"github.com/mavolin/corgi/internal/anno"
	"github.com/mavolin/corgi/internal/list"
	"github.com/mavolin/corgi/internal/stack"
)

// no mixins may be declared in other mixin declarations.
func mixinsInMixins(f *file.File) errList {
	var errs errList

	fileutil.Walk(f.Scope, func(parents []fileutil.WalkContext, ctx fileutil.WalkContext) (dive bool, err error) {
		m, ok := (*ctx.Item).(file.Mixin)
		if !ok {
			return true, nil
		}

		for _, parent := range parents {
			otherMixin, ok := (*parent.Item).(file.Mixin)
			if !ok {
				continue
			}

			errs.PushBack(&corgierr.Error{
				Message: "mixin declared inside other mixin",
				ErrorAnnotation: anno.Anno(f, anno.Annotation{
					Start:      otherMixin.Position,
					Len:        (otherMixin.Name.Col - otherMixin.Col) + len(otherMixin.Name.Ident),
					Annotation: "cannot declare mixin here",
				}),
				HintAnnotations: []corgierr.Annotation{
					anno.Anno(f, anno.Annotation{
						Start:      m.Position,
						Len:        (m.Name.Col - m.Col) + len(m.Name.Ident),
						Annotation: "other mixin",
					}),
				},
				Suggestions: []corgierr.Suggestion{
					{Suggestion: "declare `" + otherMixin.Name.Ident + "` outside of `" + m.Name.Ident + "`"},
				},
			})
		}

		return true, nil
	})

	return errs
}

// duplicate names inside same scope.
func duplicateMixinNames(f *file.File) errList {
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
			return true, nil
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

		return true, nil
	})

	return errs
}
