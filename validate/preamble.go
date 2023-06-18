package validate

import (
	"path"

	"github.com/mavolin/corgi/corgierr"
	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/fileutil"
	"github.com/mavolin/corgi/internal/anno"
)

func importNamespaces(f *file.File) errList {
	var errs errList

	for impI, imp := range f.Imports {
		for _, spec := range imp.Imports {
			impPath := fileutil.Unquote(spec.Path)
			namespace := path.Base(impPath)
			if spec.Alias != nil {
				namespace = spec.Alias.Ident
			}

			for _, cmpImp := range f.Imports[:impI] {
				for _, cmpSpec := range cmpImp.Imports {
					cmpImpPath := fileutil.Unquote(cmpSpec.Path)
					cmpNamespace := path.Base(cmpImpPath)
					if cmpSpec.Alias != nil {
						cmpNamespace = cmpSpec.Alias.Ident
					}

					if namespace != cmpNamespace {
						continue
					}

					switch {
					case impPath == cmpImpPath:
						errs.PushBack(&corgierr.Error{
							Message: "duplicate import",
							ErrorAnnotation: anno.Anno(f, anno.Annotation{
								Start:      spec.Path.Position,
								ToEOL:      true,
								Annotation: "duplicate",
							}),
							HintAnnotations: []corgierr.Annotation{
								anno.Anno(f, anno.Annotation{
									Start:      cmpSpec.Path.Position,
									ToEOL:      true,
									Annotation: "first import with this path",
								}),
							},
							Suggestions: []corgierr.Suggestion{{Suggestion: "remove one of these"}},
						})
					case spec.Alias != nil && cmpSpec.Alias != nil && spec.Alias.Ident == cmpSpec.Alias.Ident:
						errs.PushBack(&corgierr.Error{
							Message: "duplicate import alias",
							ErrorAnnotation: anno.Anno(f, anno.Annotation{
								Start:      spec.Alias.Position,
								Len:        len(spec.Alias.Ident),
								Annotation: "duplicate",
							}),
							HintAnnotations: []corgierr.Annotation{
								anno.Anno(f, anno.Annotation{
									Start:      cmpSpec.Alias.Position,
									Len:        len(cmpSpec.Alias.Ident),
									Annotation: "first import with this alias",
								}),
							},
							Suggestions: []corgierr.Suggestion{{Suggestion: "use a different alias for one of these"}},
						})
					default:
						errs.PushBack(&corgierr.Error{
							Message: "import namespace collision",
							ErrorAnnotation: anno.Anno(f, anno.Annotation{
								Start:      spec.Alias.Position,
								ToEOL:      true,
								Annotation: "duplicate",
							}),
							HintAnnotations: []corgierr.Annotation{
								anno.Anno(f, anno.Annotation{
									Start:      cmpSpec.Alias.Position,
									ToEOL:      true,
									Annotation: "first import with this namespace",
								}),
							},
							Suggestions: []corgierr.Suggestion{{Suggestion: "use an import alias"}},
						})
					}
				}
			}
		}
	}

	return errs
}

func useNamespaces(f *file.File) errList {
	var errs errList

	for useI, use := range f.Uses {
		for _, spec := range use.Uses {
			usePath := fileutil.Unquote(spec.Path)
			namespace := path.Base(usePath)
			if spec.Alias != nil {
				namespace = spec.Alias.Ident
			}

			for _, cmpUse := range f.Uses[:useI] {
				for _, cmpSpec := range cmpUse.Uses {
					cmpUsePath := fileutil.Unquote(cmpSpec.Path)
					cmpNamespace := path.Base(cmpUsePath)
					if cmpSpec.Alias != nil {
						cmpNamespace = cmpSpec.Alias.Ident
					}

					if namespace != cmpNamespace {
						continue
					}

					switch {
					case usePath == cmpUsePath:
						errs.PushBack(&corgierr.Error{
							Message: "duplicate use",
							ErrorAnnotation: anno.Anno(f, anno.Annotation{
								Start:      spec.Path.Position,
								ToEOL:      true,
								Annotation: "duplicate",
							}),
							HintAnnotations: []corgierr.Annotation{
								anno.Anno(f, anno.Annotation{
									Start:      cmpSpec.Path.Position,
									ToEOL:      true,
									Annotation: "first use with this path",
								}),
							},
							Suggestions: []corgierr.Suggestion{{Suggestion: "remove one of these"}},
						})
					case spec.Alias != nil && cmpSpec.Alias != nil && spec.Alias.Ident == cmpSpec.Alias.Ident:
						errs.PushBack(&corgierr.Error{
							Message: "duplicate use alias",
							ErrorAnnotation: anno.Anno(f, anno.Annotation{
								Start:      spec.Alias.Position,
								Len:        len(spec.Alias.Ident),
								Annotation: "duplicate",
							}),
							HintAnnotations: []corgierr.Annotation{
								anno.Anno(f, anno.Annotation{
									Start:      cmpSpec.Alias.Position,
									Len:        len(cmpSpec.Alias.Ident),
									Annotation: "first use with this alias",
								}),
							},
							Suggestions: []corgierr.Suggestion{{Suggestion: "use a different alias for one of these"}},
						})
					default:
						errs.PushBack(&corgierr.Error{
							Message: "use namespace collision",
							ErrorAnnotation: anno.Anno(f, anno.Annotation{
								Start:      spec.Alias.Position,
								ToEOL:      true,
								Annotation: "duplicate",
							}),
							HintAnnotations: []corgierr.Annotation{
								anno.Anno(f, anno.Annotation{
									Start:      cmpSpec.Alias.Position,
									ToEOL:      true,
									Annotation: "first use with this namespace",
								}),
							},
							Suggestions: []corgierr.Suggestion{{Suggestion: "use an alias"}},
						})
					}
				}
			}
		}
	}

	return errs
}

func unusedUses(f *file.File) errList {
	var n int
	for _, use := range f.Uses {
		n += len(use.Uses)
	}

	unusedSpecs := make([]file.UseSpec, 0, n)
	for _, use := range f.Uses {
		for _, spec := range use.Uses {
			// import for side effects
			if spec.Alias != nil && spec.Alias.Ident == "_" {
				continue
			}

			unusedSpecs = append(unusedSpecs, spec)
		}
	}

	fileutil.Walk(f.Scope, func(parents []fileutil.WalkContext, ctx fileutil.WalkContext) (dive bool, err error) {
		mc, ok := (*ctx.Item).(file.MixinCall)
		if !ok {
			return true, nil
		}

	unusedSpecs:
		for i, spec := range unusedSpecs {
			for _, specFile := range spec.Files {
				if mc.Mixin.File.AbsolutePath == specFile.AbsolutePath {
					copy(unusedSpecs[i:], unusedSpecs[i+1:])
					unusedSpecs = unusedSpecs[:len(unusedSpecs)-1]
					break unusedSpecs
				}
			}
		}

		if len(unusedSpecs) == 0 {
			return false, fileutil.StopWalk
		}

		return true, nil
	})

	if len(unusedSpecs) == 0 {
		return errList{}
	}

	var errs errList
	for _, spec := range unusedSpecs {
		errs.PushBack(&corgierr.Error{
			Message: "unused `use`",
			ErrorAnnotation: anno.Anno(f, anno.Annotation{
				Start:      spec.Position,
				ToEOL:      true,
				Annotation: "no mixin requires this package",
			}),
			Suggestions: []corgierr.Suggestion{
				{Suggestion: "remove this `use`"},
				{
					Suggestion: "if you are using this package for side effects, add the `_` use alias",
					Code:       "`_ " + string(spec.Path.Quote) + spec.Path.Contents + string(spec.Path.Quote) + "`",
				},
			},
		})
	}

	return errs
}
