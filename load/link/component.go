package link

import (
	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/ast"
	"github.com/mavolin/corgi/file/fileerr"
	"github.com/mavolin/corgi/file/fileerr/anno"
)

func collectComponents(ctx *context, p *file.Package) {
	p.Components = make([]*file.Component, 0, 48)

	for _, f := range p.Files {
		for _, itm := range f.Scope.Nodes {
			c, _ := itm.(*ast.Component)
			if c == nil {
				continue
			}

			// only do the check if the component has a name (i.e. no parsing error)
			if c.Name != nil && c.Name.Ident != "" {
				if other := p.ComponentByName(c.Name.Ident); other != nil {
					ctx.err(&fileerr.Error{
						Message:         "component defined twice",
						ErrorAnnotation: anno.Node(f, c, "second definition of "+c.Name.Ident),
						HintAnnotations: []fileerr.Annotation{
							anno.Node(other.File, other.Component, "first definition of "+other.Component.Name.Ident),
						},
						Suggestions: []fileerr.Suggestion{
							{Suggestion: "rename or delete either"},
						},
					})
				}
			}

			p.Components = append(p.Components, &file.Component{File: f, Component: c})
		}
	}

	p.Components = p.Components[:len(p.Components):len(p.Components)]
}
