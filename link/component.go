package link

import (
	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/ast"
	"github.com/mavolin/corgi/file/fileerr"
	"github.com/mavolin/corgi/internal/anno"
)

func collectComponents(p *file.Package) fileerr.List {
	errs := make(fileerr.List, 0, 48)

	p.Components = make([]*file.Component, 0, 48)

	for _, f := range p.Files {
		for _, itm := range f.Scope.Items {
			c, _ := itm.(*ast.Component)
			if c == nil {
				continue
			}

			if other := p.ComponentByName(c.Name.Ident); other != nil {
				errs = append(errs, &fileerr.Error{
					Message: "component defined twice",
					ErrorAnnotation: anno.Anno(f, anno.Annotation{
						Start:      c.Name.Position,
						Len:        len(c.Name.Ident),
						Annotation: "second definition of " + c.Name.Ident,
					}),
					HintAnnotations: []fileerr.Annotation{
						anno.Anno(other.File, anno.Annotation{
							Start:      other.Source.Name.Position,
							Len:        len(other.Source.Name.Ident),
							Annotation: "first definition of " + other.Source.Name.Ident,
						}),
					},
					Suggestions: []fileerr.Suggestion{
						{Suggestion: "rename or delete either"},
					},
				})
			}

			p.Components = append(p.Components, &file.Component{File: f, Source: c})
		}
	}

	p.Components = p.Components[:len(p.Components):len(p.Components)]

	return errs
}
