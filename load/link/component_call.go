package link

import (
	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/ast"
	"github.com/mavolin/corgi/file/fileerr"
	"github.com/mavolin/corgi/file/fileerr/anno"
	"github.com/mavolin/corgi/file/walk"
)

func collectComponentCalls(ctx *context, p *file.Package) {
	p.ComponentCalls = make([]*file.ComponentCall, 0, 32*len(p.Files))

	for _, f := range p.Files {
		walk.Walk(f, f.Scope, func(_ []walk.Context, wctx walk.Context) error {
			switch itm := wctx.Node.(type) {
			case *ast.ComponentCall:
				if itm != nil {
					p.ComponentCalls = append(p.ComponentCalls, &file.ComponentCall{File: f, ComponentCall: itm})
				}
			case *ast.ArrowBlock:
				if itm != nil {
					collectComponentCallsFromText(ctx, f, itm.Lines)
				}
			case *ast.Element:
				if itm != nil {
					collectComponentCallsFromAttributes(ctx, f, itm.Attributes)
				}
			case *ast.And:
				if itm != nil {
					collectComponentCallsFromAttributes(ctx, f, itm.Attributes)
				}
			default:
				if bt, _ := file.BracketText(itm); bt != nil {
					collectComponentCallsFromText(ctx, f, bt.Lines)
				}
			}
			return nil
		})
	}

	p.ComponentCalls = p.ComponentCalls[:len(p.ComponentCalls):len(p.ComponentCalls)]
}

func collectComponentCallsFromText(ctx *context, f *file.File, lns []ast.TextLine) {
	for _, ln := range lns {
		for _, itm := range ln {
			switch itm := itm.(type) {
			case *ast.ComponentCallInterpolation:
				if itm != nil && itm.ComponentCall != nil {
					f.Package.ComponentCalls = append(f.Package.ComponentCalls, &file.ComponentCall{
						File:          f,
						ComponentCall: itm.ComponentCall,
					})
				}
			case *ast.ElementInterpolation:
				if itm != nil && itm.Element != nil {
					collectComponentCallsFromAttributes(ctx, f, itm.Element.Attributes)
				}
			}
		}
	}
}

func collectComponentCallsFromAttributes(_ *context, f *file.File, attrColls []ast.AttributeCollection) {
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

			compCallAttr, _ := simpleAttr.Value.(*ast.ComponentCallAttributeValue)
			if compCallAttr != nil && compCallAttr.ComponentCall != nil {
				f.Package.ComponentCalls = append(f.Package.ComponentCalls, &file.ComponentCall{
					File:          f,
					ComponentCall: compCallAttr.ComponentCall,
				})
			}
		}
	}
}

func linkLocalComponentCalls(_ *context, p *file.Package) {
	for _, cc := range p.ComponentCalls {
		if cc.Namespace != nil || cc.Name == nil {
			continue
		}

		if c := p.ComponentByName(cc.Name.Ident); c != nil {
			cc.Component = c
			continue
		}
	}
	return
}

func linkExternalComponentCalls(ctx *context, p *file.Package) {
	for _, cc := range p.ComponentCalls {
		if cc.Component != nil {
			continue
		}

		linkComponentCall(ctx, cc.File, cc)
	}
}

func linkComponentCall(ctx *context, f *file.File, cc *file.ComponentCall) {
	if cc.Name == nil || cc.Name.Ident == "" || (cc.Namespace != nil && cc.Namespace.Ident == "") {
		return
	}

	namespace := "."
	if cc.Namespace != nil {
		namespace = cc.Namespace.Ident
	}

	missingImport := true

	for _, imp := range f.Imports {
		if imp.Namespace() != namespace {
			continue
		}
		missingImport = false

		c := imp.Package.ComponentByName(cc.Name.Ident)
		if c != nil {
			cc.Component = c
			return
		}
	}

	if missingImport && namespace != "." {
		ctx.err(&fileerr.Error{
			Message:         "missing import",
			ErrorAnnotation: anno.Node(f, cc.Namespace, "there is no import for this package"),
		})
		return
	}

	ctx.err(newUnknownComponentError(f, cc))
}

func newUnknownComponentError(f *file.File, cc *file.ComponentCall) *fileerr.Error {
	var hint *fileerr.Annotation
	var suggestion *fileerr.Suggestion

	namespace := "."
	if cc.Namespace != nil {
		namespace = cc.Namespace.Ident
	}
	imp := f.ImportByNamespace(namespace)
	if imp != nil {
		a := anno.Node(f, imp.ImportSpec, "in this package")
		hint = &a
	} else if file.IsExported(cc.Name.Ident) {
		suggestion = &fileerr.Suggestion{Suggestion: "did you forget to add a dot import?"}
	}

	ferr := &fileerr.Error{
		Message:         "unknown component",
		ErrorAnnotation: anno.Node(f, cc, "found no component with this name"),
	}
	if hint != nil {
		ferr.HintAnnotations = []fileerr.Annotation{*hint}
	}
	if suggestion != nil {
		ferr.Suggestions = []fileerr.Suggestion{*suggestion}
	}

	return ferr
}
