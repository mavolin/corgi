package validate

import (
	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/fileerr"
	"github.com/mavolin/corgi/internal/anno"
	"github.com/mavolin/corgi/internal/list"
)

func libraryMixinNameConflicts(fs []*file.File) *errList {
	var errs errList

	var foundMixins list.List[struct {
		File  *file.File
		Mixin file.Mixin
	}]

	for _, f := range fs {
		for _, itm := range f.Scope {
			m, ok := itm.(file.Mixin)
			if !ok {
				continue
			}

			for otherE := foundMixins.Front(); otherE != nil; otherE = otherE.Next() {
				other := otherE.V()
				if m.Name.Ident != other.Mixin.Name.Ident || f.Name == other.File.Name {
					continue
				}

				errs.PushBack(&fileerr.Error{
					Message: "duplicate mixin in package",
					ErrorAnnotation: anno.Anno(f, anno.Annotation{
						Start:      m.Name.Position,
						Len:        len(m.Name.Ident),
						Annotation: "same mixin name used here",
					}),
					HintAnnotations: []fileerr.Annotation{
						anno.Anno(other.File, anno.Annotation{
							Start:      other.Mixin.Name.Position,
							Len:        len(other.Mixin.Name.Ident),
							Annotation: "and here",
						}),
					},
				})

				foundMixins.PushBack(struct {
					File  *file.File
					Mixin file.Mixin
				}{File: f, Mixin: m})
				break
			}
		}
	}

	return &errs
}
