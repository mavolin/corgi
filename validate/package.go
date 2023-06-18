package validate

import (
	"fmt"

	"github.com/mavolin/corgi/corgierr"
	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/internal/anno"
	"github.com/mavolin/corgi/internal/list"
)

func packageMixinNameConflicts(fs []file.File) errList {
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

				errs.PushBack(&corgierr.Error{
					Message: "duplicate mixin in package",
					ErrorAnnotation: anno.Anno(&f, anno.Annotation{
						Start: m.Name.Position,
						Len:   len(m.Name.Ident),
						Annotation: fmt.Sprintf("mixin redeclared in `%s:%d:%d` (see error below)",
							other.File.Name, other.Mixin.Name.Line, other.Mixin.Name.Col),
					}),
				})
				errs.PushBack(&corgierr.Error{
					Message: "duplicate mixin in package",
					ErrorAnnotation: anno.Anno(&f, anno.Annotation{
						Start: other.Mixin.Name.Position,
						Len:   len(other.Mixin.Name.Ident),
						Annotation: fmt.Sprintf("mixin redeclared in `%s:%d:%d` (see error above)",
							f.Name, m.Name.Line, m.Name.Col),
					}),
				})

				foundMixins.PushBack(struct {
					File  *file.File
					Mixin file.Mixin
				}{File: &f, Mixin: m})
				break
			}
		}
	}

	return errs
}
