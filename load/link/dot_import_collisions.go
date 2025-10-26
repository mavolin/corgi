package link

import (
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
)

// ============================================================================
// Component
// ======================================================================================

type dotImportComponentCollisionCheck struct{}

func (l *linker) CheckDotImportComponentCollisions() {
	defer l.Ran(l.Pkg, dotImportComponentCollisionCheck{})
	l.Require(l.Pkg, importsLoaded{})

	logger := l.Logger.WithGroup("checks.dot_import_collisions.components")
	logger.Debug("Checking for component collisions through dot imports")

	for _, f := range l.Pkg.Files {
		dotImports := l.dotImports[f]
		if len(dotImports) == 0 {
			continue
		}

		logger := logger.With(slog.String("file", string(f.Name)))

		dupls := make(map[file.Identifier][]*file.Component)
		for _, imp := range dotImports {
			addComponentsFromPackage(dupls, imp.Package)
		}
		addComponentsFromPackage(dupls, f.Package)

		for name, comps := range dupls {
			if len(comps) < 2 {
				continue // no collision
			}

			logger := logger.With(slog.String("name", string(name)))
			logger.Error("Component collision", slog.Int("count", len(comps)))

			primaries := make([]diagnostic.Annotation, len(comps))
			secondaries := make([]diagnostic.Annotation, 0, len(comps))
			for i, comp := range comps {
				if comp.File.Package == f.Package {
					primaries[i] = anno.Node(comp.File, comp.AST, "`"+string(name)+"` defined locally")
				} else {
					primaries[i] = anno.Node(f, findImport(dotImports, comp.File.Package).AST, "defines `"+string(name)+"`")
					secondaries = append(secondaries, anno.Anno(comp.File, anno.Annotation{
						Highlight:  anno.HighlightNode(comp.AST),
						Context:    anno.ContextNode(comp.AST.Header),
						Annotation: "defined here",
					}))
				}
			}

			l.Report(&diagnostic.Diagnostic{
				Message:   "dot import collision: multiple definitions for component of the same name",
				Primary:   primaries,
				Secondary: secondaries,
				Hints: []diagnostic.Hint{
					{Hint: "Use import aliases instead."},
				},
			})
		}
	}
}

func addComponentsFromPackage(dupls map[file.Identifier][]*file.Component, p *file.Package) {
	if p == nil || p.PackageSymbols == nil || p.Components == nil {
		return
	}

Components:
	for _, comp := range p.Components {
		if comp.AST.Header == nil || comp.AST.Header.Name == nil {
			continue
		}
		if !comp.Name.Exported() {
			continue
		}

		for _, dupl := range dupls[comp.Name] {
			if dupl.File.Package == comp.File.Package {
				// don't report collisions within the same package or report
				// the same component multiple times if there are duplicate
				// dot imports
				continue Components
			}
		}
		dupls[comp.Name] = append(dupls[comp.Name], comp)
	}
}

// ============================================================================
// Element Spec
// ======================================================================================

type dotImportElementSpecCollisionCheck struct{}

func (l *linker) CheckDotImportElementSpecCollisions() {
	defer l.Ran(l.Pkg, dotImportElementSpecCollisionCheck{})
	l.Require(l.Pkg, importsLoaded{})

	logger := l.Logger.WithGroup("checks.dot_import_collisions.element_specs")
	logger.Debug("Checking for element spec collisions through dot imports")

	for _, f := range l.Pkg.Files {
		dotImports := l.dotImports[f]
		if len(dotImports) == 0 {
			continue
		}

		logger := logger.With(slog.String("file", string(f.Name)))

		dupls := make(map[file.CanonicalElementName][]*file.ElementSpec)
		for _, imp := range dotImports {
			addElementSpecsFromPackage(dupls, imp.Package)
		}
		addElementSpecsFromPackage(dupls, f.Package)

		for name, elems := range dupls {
			if len(elems) < 2 {
				continue // no collision
			}

			logger.Error("Element definition collision",
				slog.String("name", string(name)),
				slog.Int("count", len(elems)))

			primaries := make([]diagnostic.Annotation, 0, len(elems)+1)
			secondaries := make([]diagnostic.Annotation, 0, 2*len(elems))
			for _, elem := range elems {
				if elem.File.Package == f.Package {
					primaries = appendElementSpecLocationAnnotations(primaries, nil, elem, "`"+elem.StylizedHTMLName+"` defined locally")
				} else {
					primaries = append(primaries, anno.Node(f, findImport(dotImports, elem.File.Package).AST, "defines `"+elem.StylizedHTMLName+"`"))
					secondaries = appendElementSpecLocationAnnotations(secondaries, nil, elem, "defined here")
				}
			}

			l.Report(&diagnostic.Diagnostic{
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
}

func addElementSpecsFromPackage(dupls map[file.CanonicalElementName][]*file.ElementSpec, p *file.Package) {
	if p == nil || p.PackageSymbols == nil || p.ElementSpecs == nil {
		return
	}

Specs:
	for _, elem := range p.ElementSpecs {
		if elem.HTMLName == "" {
			continue
		}

		for _, dupl := range dupls[elem.HTMLName] {
			if dupl.File.Package == elem.File.Package {
				// don't report collisions within the same package or report
				// the same spec multiple times if there are duplicate
				// dot imports
				continue Specs
			}
		}
		dupls[elem.HTMLName] = append(dupls[elem.HTMLName], elem)
	}
}

// ============================================================================
// Attribute Spec
// ======================================================================================

type dotImportAttributeSpecCollisionCheck struct{}

func (l *linker) CheckDotImportAttributeSpecCollisions() {
	defer l.Ran(l.Pkg, dotImportAttributeSpecCollisionCheck{})
	l.Require(l.Pkg, importsLoaded{})

	logger := l.Logger.WithGroup("checks.dot_import_collisions.attribute_specs")
	logger.Debug("Checking for attribute spec collisions through dot imports")

	for _, f := range l.Pkg.Files {
		dotImports := l.dotImports[f]
		if len(dotImports) == 0 {
			continue
		}

		logger := logger.With(slog.String("file", string(f.Name)))

		dupls := make(map[file.CanonicalAttributeName][]*file.AttributeSpec)
		for _, imp := range dotImports {
			addAttributeSpecsFromPackage(dupls, imp.Package)
		}
		addAttributeSpecsFromPackage(dupls, f.Package)

		for sel, attrs := range dupls {
			if len(attrs) < 2 {
				continue // no collision
			}

			logger.Error("Attribute definition collision",
				slog.String("selector", string(sel)),
				slog.Int("count", len(attrs)))

			primaries := make([]diagnostic.Annotation, 0, len(attrs))
			secondaries := make([]diagnostic.Annotation, 0, len(attrs))
			for _, attr := range attrs {
				if attr.File.Package == f.Package {
					primaries = appendAttrSpecLocationAnnotations(primaries, nil, attr, "`"+string(sel)+"` defined locally")
				} else {
					primaries = append(primaries, anno.Node(f, findImport(dotImports, attr.File.Package).AST, "defines `"+string(sel)+"`"))
					secondaries = appendAttrSpecLocationAnnotations(secondaries, nil, attr, "defined here")
				}
			}

			l.Report(&diagnostic.Diagnostic{
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
}

func addAttributeSpecsFromPackage(dupls map[file.CanonicalAttributeName][]*file.AttributeSpec, p *file.Package) {
	if p == nil || p.PackageSymbols == nil || p.AttributeSpecs == nil {
		return
	}

Specs:
	for _, spec := range p.AttributeSpecs {
		sel := htmlAttrName(spec)
		if sel == "" {
			continue
		}

		for _, dupl := range dupls[sel] {
			if dupl.File.Package == spec.File.Package {
				// don't report collisions within the same package or report
				// the same spec multiple times if there are duplicate
				// dot imports
				continue Specs
			}
		}
		dupls[sel] = append(dupls[sel], spec)
	}
}

// ============================================================================
// Helpers
// ======================================================================================

func findImport(imps []*file.Import, p *file.Package) *file.Import {
	for _, imp := range imps {
		if imp.Package == p {
			return imp
		}
	}
	return nil
}
