package link

import (
	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/ast"
	"github.com/mavolin/corgi/file/fileerr"
	"github.com/mavolin/corgi/file/walk"
	"github.com/mavolin/corgi/internal/anno"
)

func collectComponentCalls(p *file.Package) {
	p.ComponentCalls = make([]*file.ComponentCall, 0, 32*len(p.Files))

	for _, f := range p.Files {
		walk.Walk(f, f.Scope, func(_ []walk.Context, wctx walk.Context) error {
			switch itm := wctx.Item.(type) {
			case *ast.ComponentCall:
				if itm != nil {
					p.ComponentCalls = append(p.ComponentCalls, &file.ComponentCall{File: f, Source: itm})
				}
			case *ast.ArrowBlock:
				if itm != nil {
					collectComponentCallsFromText(f, itm.Lines)
				}
			case *ast.Element:
				if itm != nil {
					collectComponentCallsFromAttributes(f, itm.Attributes)
				}
			case *ast.And:
				if itm != nil {
					collectComponentCallsFromAttributes(f, itm.Attributes)
				}
			default:
				if bt, _ := file.BracketText(itm); bt != nil {
					collectComponentCallsFromText(f, bt.Lines)
				}
			}
			return nil
		})
	}

	p.ComponentCalls = p.ComponentCalls[:len(p.ComponentCalls):len(p.ComponentCalls)]
}

func collectComponentCallsFromText(f *file.File, lns []ast.TextLine) {
	for _, ln := range lns {
		for _, itm := range ln {
			switch itm := itm.(type) {
			case *ast.ComponentCallInterpolation:
				if itm != nil && itm.ComponentCall != nil {
					f.Package.ComponentCalls = append(f.Package.ComponentCalls, &file.ComponentCall{
						File:   f,
						Source: itm.ComponentCall,
					})
				}
			case *ast.ElementInterpolation:
				if itm != nil && itm.Element != nil {
					collectComponentCallsFromAttributes(f, itm.Element.Attributes)
				}
			}
		}
	}
}

func collectComponentCallsFromAttributes(f *file.File, attrColls []ast.AttributeCollection) {
	for _, attrColl := range attrColls {
		attrList, _ := attrColl.(*ast.AttributeList)
		if attrList == nil {
			continue
		}

		for _, attr := range attrList.Attributes {
			simpleAttr, _ := attr.(*ast.SimpleAttribute)
			if simpleAttr == nil {
				continue
			}

			compCallAttr, _ := simpleAttr.Value.(*ast.ComponentCallAttribute)
			if compCallAttr != nil && compCallAttr.ComponentCall != nil {
				f.Package.ComponentCalls = append(f.Package.ComponentCalls, &file.ComponentCall{
					File:   f,
					Source: compCallAttr.ComponentCall,
				})
			}
		}
	}
}

func linkLocalComponentCalls(p *file.Package) {
	for _, cc := range p.ComponentCalls {
		if cc.Namespace() != "" {
			continue
		}

		if c := p.ComponentByName(cc.Name()); c != nil {
			cc.Component = c
			continue
		}
	}
	return
}

func linkExternalComponentCalls(p *file.Package) fileerr.List {
	errs := make(fileerr.List, 0, len(p.ComponentCalls))

	for _, cc := range p.ComponentCalls {
		if cc.Component != nil {
			continue
		}

		if err := linkComponentCall(cc.File, cc); err != nil {
			errs = append(errs, err)
		}
	}

	return errs
}

func linkComponentCall(f *file.File, cc *file.ComponentCall) *fileerr.Error {
	if cc.Source.Name == nil {
		return nil
	}

	namespace := cc.Namespace()
	if namespace == "" {
		namespace = "."
	}

	missingImport := true

	for _, imp := range f.Imports {
		if imp.Namespace() != namespace {
			continue
		}
		missingImport = false

		c := imp.Package.ComponentByName(cc.Name())
		if c != nil {
			cc.Component = c
			return nil
		}
	}

	if missingImport && namespace != "." {
		return &fileerr.Error{
			Message: "missing import",
			ErrorAnnotation: anno.Anno(f, anno.Annotation{
				Start:      cc.Source.Namespace.Position,
				Len:        len(namespace),
				Annotation: "there is no import for this package",
			}),
		}
	}

	return newUnknownComponentError(f, cc)
}

func newUnknownComponentError(f *file.File, cc *file.ComponentCall) *fileerr.Error {
	var hint *fileerr.Annotation
	var suggestion *fileerr.Suggestion

	namespace := cc.Namespace()
	if namespace == "" {
		namespace = "."
	}
	imp := f.ImportByNamespace(namespace)
	if imp != nil {
		a := anno.Anno(f, anno.Annotation{
			Start:      imp.Source.Position,
			ToEOL:      true,
			Annotation: "in this package",
		})
		hint = &a
	} else if cc.Exported() {
		suggestion = &fileerr.Suggestion{Suggestion: "did you forget to add a dot import?"}
	}

	start := cc.Source.Name.Position
	if cc.Source.Namespace != nil {
		start = cc.Source.Namespace.Position
	}

	ferr := &fileerr.Error{
		Message: "unknown component",
		ErrorAnnotation: anno.Anno(f, anno.Annotation{
			Start:      start,
			End:        cc.Source.Name.Position,
			EndOffset:  len(cc.Source.Name.Ident),
			Annotation: "found no component with this name",
		}),
	}
	if hint != nil {
		ferr.HintAnnotations = []fileerr.Annotation{*hint}
	}
	if suggestion != nil {
		ferr.Suggestions = []fileerr.Suggestion{*suggestion}
	}

	return ferr
}
