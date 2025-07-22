package link

import (
	"context"
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
)

func (l *linker) CheckDotImportCollisions(_ context.Context) {
	logger := l.logger.WithGroup("check.dot_import_collisions")
	logger.Debug("Checking for dot import collisions")

	dotImports := make([]*file.Import, 0, 32)

	for _, f := range l.p.Files {
		dotImports = dotImports[:0]
		for _, imp := range f.Imports {
			if imp == nil || imp.AST == nil || imp.AST.Alias == nil || imp.AST.Alias.Name != "." || imp.Package == nil {
				continue
			}
			dotImports = append(dotImports, imp)
		}
		if len(dotImports) <= 1 {
			continue
		}

		logger = logger.With(slog.String("file", f.Name))

		l.checkDotImportComponentCollision(logger, f, dotImports)
		l.checkDotImportElementSpecCollisions(logger, f, dotImports)
		l.checkDotImportAttributeSpecCollision(logger, f, dotImports)
	}
}

// ============================================================================
// Component Collisions
// ======================================================================================

func (l *linker) checkDotImportComponentCollision(logger *slog.Logger, f *file.File, dotImports []*file.Import) {
	logger = logger.WithGroup("components")

	dupls := make(map[string][]*file.Component)

	for _, imp := range dotImports {
		addComponentsFromPackage(dupls, imp.Package)
	}
	addComponentsFromPackage(dupls, f.Package)

	for name, comps := range dupls {
		if len(comps) < 2 {
			continue // no collision
		}

		logger := logger.With(slog.String("name", name))
		logger.Error("Component collision", slog.Int("count", len(comps)))

		primaries := make([]diagnostic.Annotation, len(comps))
		secondaries := make([]diagnostic.Annotation, 0, len(comps))
		for i, comp := range comps {
			if comp.File.Package == f.Package {
				primaries[i] = anno.Node(f, comp.DefinedAST, "`"+name+"` defined locally")
			} else {
				primaries[i] = anno.Node(f, f.ImportByPackage(comp.File.Package).AST, "defines `"+name+"`")
				secondaries = append(secondaries, anno.Anno(comp.File, anno.Annotation{
					Highlight:  anno.HighlightNode(comp.DefinedAST),
					Context:    anno.ContextLines(comp.DefinedAST.Start(), comp.DefinedAST.Header.End()),
					Annotation: "defined here",
				}))
			}
		}

		l.report(&diagnostic.Diagnostic{
			Message:   "dot import collision: multiple definitions for component of the same name",
			Primary:   primaries,
			Secondary: secondaries,
			Hints: []diagnostic.Hint{
				{Hint: "Use import aliases instead."},
			},
		})
	}
}

func addComponentsFromPackage(dupls map[string][]*file.Component, p *file.Package) {
Components:
	for _, comp := range p.Components {
		if comp.DefinedAST.Header == nil || comp.DefinedAST.Header.Name == nil {
			continue
		}
		name := comp.DefinedAST.Header.Name.Name
		if !file.IsExported(name) {
			continue
		}

		for _, dupl := range dupls[name] {
			if dupl.File.Package == comp.File.Package {
				continue Components // only report one collision per package
			}
		}
		dupls[name] = append(dupls[name], comp)
	}
}

// ============================================================================
// Element Spec Collisions
// ======================================================================================

func (l *linker) checkDotImportElementSpecCollisions(logger *slog.Logger, f *file.File, dotImports []*file.Import) {
	logger = logger.WithGroup("element_definitions")

	dupls := make(map[string][]*file.ElementSpec)
	for _, imp := range dotImports {
		addElementSpecsFromPackage(dupls, imp.Package)
	}
	addElementSpecsFromPackage(dupls, f.Package)

	for name, elems := range dupls {
		if len(elems) < 2 {
			continue // no collision
		}

		logger.Error("Element definition collision",
			slog.String("name", name),
			slog.Int("count", len(elems)))

		primaries := make([]diagnostic.Annotation, 0, len(elems)+1)
		secondaries := make([]diagnostic.Annotation, 0, 2*len(elems))
		for _, elem := range elems {
			if elem.File.Package == f.Package {
				primaries = appendElementSpecLocationAnnotations(primaries, elem, "`"+name+"` defined locally")
			} else {
				primaries = append(primaries, anno.Node(f, f.ImportByPackage(elem.File.Package).AST, "defines `"+name+"`"))
				secondaries = appendElementSpecLocationAnnotations(secondaries, elem, "defined here")
			}
		}

		l.report(&diagnostic.Diagnostic{
			Message:   "dot import collision: multiple definitions for element of the same name",
			Primary:   primaries,
			Secondary: secondaries,
			Hints: []diagnostic.Hint{
				{Hint: "Use import aliases instead."},
				{Hint: "Remember that element names are case-insensitive."},
			},
		})
	}
}

func addElementSpecsFromPackage(dupls map[string][]*file.ElementSpec, p *file.Package) {
Specs:
	for _, elem := range p.ElementSpecs {
		name := elem.FullName()
		if name == "" {
			continue
		}

		for _, dupl := range dupls[name] {
			if dupl.File.Package == elem.File.Package {
				continue Specs // only report one collision per package
			}
		}
		dupls[name] = append(dupls[name], elem)
	}
}

func appendElementSpecLocationAnnotations(annos []diagnostic.Annotation, elem *file.ElementSpec, text string) []diagnostic.Annotation {
	if elem.Definition.LParen == nil && elem.Definition.Prefix != nil {
		return append(annos, anno.Range(elem.File, elem.Definition.Prefix.Start(), elem.AST.Name.End(), text))
	}

	if elem.Definition.Prefix != nil {
		annos = append(annos, anno.Node(elem.File, elem.Definition.Prefix, "with this prefix"))
	}
	return append(annos, anno.Node(elem.File, elem.AST.Name, text))
}

// ============================================================================
// Attribute Spec Collisions
// ======================================================================================

func (l *linker) checkDotImportAttributeSpecCollision(logger *slog.Logger, f *file.File, dotImports []*file.Import) {
	logger = logger.WithGroup("attribute_definitions")

	dupls := make(map[string][]*file.AttributeSpec)
	for _, imp := range dotImports {
		addAttributeSpecsFromPackage(dupls, imp.Package)
	}
	addAttributeSpecsFromPackage(dupls, f.Package)

	for sel, attrs := range dupls {
		if len(attrs) < 2 {
			continue // no collision
		}

		logger.Error("Attribute definition collision",
			slog.String("selector", sel),
			slog.Int("count", len(attrs)))

		primaries := make([]diagnostic.Annotation, 0, len(attrs))
		secondaries := make([]diagnostic.Annotation, 0, len(attrs))
		for _, attr := range attrs {
			if attr.File.Package == f.Package {
				primaries = appendAttrSpecLocationAnnotations(primaries, attr, "`"+sel+"` defined locally")
			} else {
				primaries = append(primaries, anno.Node(f, f.ImportByPackage(attr.File.Package).AST, "defines `"+sel+"`"))
				secondaries = appendAttrSpecLocationAnnotations(secondaries, attr, "defined here")
			}
		}

		l.report(&diagnostic.Diagnostic{
			Message:   "dot import collision: multiple definitions for attribute of the same name",
			Primary:   primaries,
			Secondary: secondaries,
			Hints: []diagnostic.Hint{
				{Hint: "Use import aliases instead."},
				{Hint: "Remember that attribute names are case-insensitive."},
			},
		})
	}
}

func addAttributeSpecsFromPackage(dupls map[string][]*file.AttributeSpec, p *file.Package) {
Specs:
	for _, attr := range p.AttributeSpecs {
		info := attrSpecInfo(attr)
		if info == nil {
			continue
		}
		sel := info.fullSelector()

		for _, dupl := range dupls[sel] {
			if dupl.File.Package == attr.File.Package {
				continue Specs // only report one collision per package
			}
		}
		dupls[sel] = append(dupls[sel], attr)
	}
}

func appendAttrSpecLocationAnnotations(annos []diagnostic.Annotation, attr *file.AttributeSpec, text string) []diagnostic.Annotation {
	if attr.Definition.LParen == nil && attr.Definition.Prefix != nil {
		return append(annos, anno.Range(attr.File, attr.Definition.Prefix.Start(), attr.AST.Selector.End(), text))
	}

	if attr.Definition.Prefix != nil {
		annos = append(annos,
			anno.Node(attr.File, attr.Definition.Prefix, "with this prefix"))
	}
	return append(annos, anno.Node(attr.File, attr.AST.Selector, text))
}
