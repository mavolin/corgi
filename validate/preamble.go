package validate

import (
	"path"
	"strconv"

	"github.com/mavolin/corgi/corgierr"
	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/fileutil"
	"github.com/mavolin/corgi/internal/anno"
)

func duplicateImports(f *file.File) errList {
	var errs errList

	cmps := make(map[string] /* namespace */ file.ImportSpec)

	for _, imp := range f.Imports {
		for _, a := range imp.Imports {
			aPath := fileutil.Unquote(a.Path)
			namespace := path.Base(aPath)
			if a.Alias != nil {
				namespace = a.Alias.Ident
			}

			b, ok := cmps[namespace]
			if !ok {
				cmps[namespace] = a
				continue
			}

			bPath := fileutil.Unquote(b.Path)
			if aPath == bPath {
				aStart := a.Path.Col
				if a.Alias != nil {
					aStart = a.Alias.Col
				}

				var aLine string
				aPathQuot := fileutil.Quote(a.Path)
				if a.Alias != nil {
					aLine = a.Alias.Ident + " " + aPathQuot
				} else {
					aLine = aPathQuot
				}

				bStart := b.Path.Col
				if b.Alias != nil {
					bStart = b.Alias.Col
				}

				var bLine string
				bPathQuot := fileutil.Quote(b.Path)
				if b.Alias != nil {
					bLine = b.Alias.Ident + " " + bPathQuot
				} else {
					bLine = bPathQuot
				}

				errs.PushBack(&corgierr.Error{
					Message: "duplicate import",
					ErrorAnnotation: corgierr.Annotation{
						File:         f,
						ContextStart: a.Path.Line,
						ContextEnd:   a.Path.Line + 1,
						Line:         a.Path.Line,
						Start:        aStart,
						End:          a.Col + len(aPathQuot),
						Annotation:   "duplicate",
						Lines:        []string{aLine},
					},
					HintAnnotations: []corgierr.Annotation{
						{
							File:         f,
							ContextStart: b.Path.Line,
							ContextEnd:   b.Path.Line + 1,
							Line:         b.Path.Line,
							Start:        bStart,
							End:          b.Col + len(bPathQuot),
							Annotation:   "first import with this path",
							Lines:        []string{bLine},
						},
					},
					Suggestions: []corgierr.Suggestion{{Suggestion: "remove one of these"}},
				})
			}
		}
	}

	return errs
}

type importNamespace struct {
	imp  file.ImportSpec
	file *file.File
}

func importNamespaces(cmps map[string]importNamespace, f *file.File) errList {
	var errs errList

	for _, imp := range f.Imports {
		for _, a := range imp.Imports {
			aPath := fileutil.Unquote(a.Path)
			namespace := path.Base(aPath)
			if a.Alias != nil {
				namespace = a.Alias.Ident
			}

			cmp, ok := cmps[namespace]
			if !ok {
				cmps[namespace] = importNamespace{a, f}
				continue
			}

			b := cmp.imp
			bPath := fileutil.Unquote(b.Path)
			if aPath == bPath {
				continue
			}

			var suggestions []corgierr.Suggestion
			switch {
			case a.Alias == nil && b.Alias == nil:
				suggestions = append(suggestions, corgierr.Suggestion{
					Suggestion: "use an import alias",
					Example:    "`" + namespace + "1 " + strconv.Quote(aPath) + "` or `" + namespace + "1 " + strconv.Quote(bPath) + "`",
				})
			case a.Alias == nil:
				suggestions = append(suggestions, corgierr.Suggestion{
					Suggestion: "use an import alias",
					Example:    "`" + namespace + "1 " + strconv.Quote(aPath) + "`",
				})
			case b.Alias == nil:
				suggestions = append(suggestions, corgierr.Suggestion{
					Suggestion: "use an import alias",
					Example:    "`" + namespace + "1 " + strconv.Quote(bPath) + "`",
				})
			}
			switch {
			case a.Alias != nil && b.Alias != nil:
				suggestions = append(suggestions, corgierr.Suggestion{
					Suggestion: "use a different import alias",
					Example:    "`" + namespace + "1 " + strconv.Quote(aPath) + "` or `" + namespace + "1 " + strconv.Quote(bPath) + "`",
				})
			case a.Alias != nil:
				suggestions = append(suggestions, corgierr.Suggestion{
					Suggestion: "use a different import alias",
					Example:    "`" + namespace + "1 " + strconv.Quote(aPath) + "`",
				})
			case b.Alias != nil:
				suggestions = append(suggestions, corgierr.Suggestion{
					Suggestion: "use a different import alias",
					Example:    "`" + namespace + "1 " + strconv.Quote(bPath) + "`",
				})
			}

			aStart := a.Path.Col
			if a.Alias != nil {
				aStart = a.Alias.Col
			}

			var aLine string
			aPathQuot := fileutil.Quote(a.Path)
			if a.Alias != nil {
				aLine = a.Alias.Ident + " " + aPathQuot
			} else {
				aLine = aPathQuot
			}

			bStart := b.Path.Col
			if b.Alias != nil {
				bStart = b.Alias.Col
			}

			var bLine string
			bPathQuot := fileutil.Quote(b.Path)
			if b.Alias != nil {
				bLine = b.Alias.Ident + " " + bPathQuot
			} else {
				bLine = bPathQuot
			}

			errs.PushBack(&corgierr.Error{
				Message: "duplicate import namespace",
				ErrorAnnotation: corgierr.Annotation{
					File:         f,
					ContextStart: a.Path.Line,
					ContextEnd:   a.Path.Line + 1,
					Line:         a.Path.Line,
					Start:        aStart,
					End:          a.Col + len(aPathQuot),
					Annotation:   "second import",
					Lines:        []string{aLine},
				},
				HintAnnotations: []corgierr.Annotation{
					{
						File:         f,
						ContextStart: b.Path.Line,
						ContextEnd:   b.Path.Line + 1,
						Line:         b.Path.Line,
						Start:        bStart,
						End:          b.Col + len(bPathQuot),
						Annotation:   "first import",
						Lines:        []string{bLine},
					},
				},
				Suggestions: suggestions,
			})
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
			for _, specFile := range spec.Library.Files {
				if mc.Mixin.File.Module+mc.Mixin.File.ModulePath == specFile.Module+specFile.ModulePath {
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
					Code:       "`_ " + strconv.Quote(fileutil.Unquote(spec.Path)) + "`",
				},
			},
		})
	}

	return errs
}
