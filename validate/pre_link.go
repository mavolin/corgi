package validate

import (
	"path"

	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/fileutil"
	"github.com/mavolin/corgi/fileerr"
	"github.com/mavolin/corgi/internal/anno"
)

func useNamespaces(f *file.File) *errList {
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
						errs.PushBack(&fileerr.Error{
							Message: "duplicate use",
							ErrorAnnotation: anno.Anno(f, anno.Annotation{
								Start:      spec.Path.Position,
								ToEOL:      true,
								Annotation: "duplicate",
							}),
							HintAnnotations: []fileerr.Annotation{
								anno.Anno(f, anno.Annotation{
									Start:      cmpSpec.Path.Position,
									ToEOL:      true,
									Annotation: "first use with this path",
								}),
							},
							Suggestions: []fileerr.Suggestion{{Suggestion: "remove one of these"}},
						})
					case spec.Alias != nil && cmpSpec.Alias != nil && spec.Alias.Ident == cmpSpec.Alias.Ident:
						errs.PushBack(&fileerr.Error{
							Message: "duplicate use alias",
							ErrorAnnotation: anno.Anno(f, anno.Annotation{
								Start:      spec.Alias.Position,
								Len:        len(spec.Alias.Ident),
								Annotation: "duplicate",
							}),
							HintAnnotations: []fileerr.Annotation{
								anno.Anno(f, anno.Annotation{
									Start:      cmpSpec.Alias.Position,
									Len:        len(cmpSpec.Alias.Ident),
									Annotation: "first use with this alias",
								}),
							},
							Suggestions: []fileerr.Suggestion{{Suggestion: "use a different alias for one of these"}},
						})
					default:
						errs.PushBack(&fileerr.Error{
							Message: "use namespace collision",
							ErrorAnnotation: anno.Anno(f, anno.Annotation{
								Start:      spec.Alias.Position,
								ToEOL:      true,
								Annotation: "duplicate",
							}),
							HintAnnotations: []fileerr.Annotation{
								anno.Anno(f, anno.Annotation{
									Start:      cmpSpec.Alias.Position,
									ToEOL:      true,
									Annotation: "first use with this namespace",
								}),
							},
							Suggestions: []fileerr.Suggestion{{Suggestion: "use an alias"}},
						})
					}
				}
			}
		}
	}

	return &errs
}
