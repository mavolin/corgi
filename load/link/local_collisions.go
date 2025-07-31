package link

import (
	"log/slog"
	"slices"
	"strings"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
)

// ============================================================================
// Component
// ======================================================================================

func (l *linker) CheckComponentCollisions() {
	logger := l.logger.WithGroup("checks.collisions.components")
	logger.Debug("Checking for component collisions")

	if len(l.p.Components) <= 1 {
		logger.Info("One or no components, skipping")
		return
	}

	dupls := make(map[componentName][]*file.Component)

	for _, comp := range l.p.Components {
		if comp.AST.Header == nil || comp.AST.Header.Name == nil {
			continue
		}

		name := comp.AST.Header.Name.Name
		dupls[name] = append(dupls[name], comp)
	}

	for name, comps := range dupls {
		if len(comps) <= 1 {
			continue
		}

		logger.Error("Found component collisions",
			slog.String("name", name),
			slog.Int("count", len(comps)))

		primaries := make([]diagnostic.Annotation, len(comps))
		for i, comp := range comps {
			primaries[i] = anno.Node(comp.File, comp.AST, "defined here")
		}

		l.report(&diagnostic.Diagnostic{
			Message: "component defined multiple times",
			Primary: primaries,
		})
	}
}

// ============================================================================
// Element Spec
// ======================================================================================

func (l *linker) CheckElementSpecCollisions() {
	logger := l.logger.WithGroup("checks.collisions.element_specs")
	logger.Debug("Checking for element spec collisions")

	if len(l.p.ElementSpecs) <= 1 {
		return
	}

	qualifiedDupls := make(map[elementName][]*file.ElementSpec)
	htmlNameDupls := make(map[fullElementName][]*file.ElementSpec)

	// Collect all element specs
	for _, elem := range l.p.ElementSpecs {
		if elem.Definition == nil || elem.AST.Name == nil {
			continue
		}

		qualName, htmlName := strings.ToLower(elem.QualifiedName()), strings.ToLower(elem.HTMLName())
		if qualName == "" {
			continue
		}
		qualifiedDupls[qualName] = append(qualifiedDupls[qualName], elem)

		if htmlName != qualName {
			htmlNameDupls[htmlName] = append(htmlNameDupls[htmlName], elem)
		}
	}

	for name, elems := range qualifiedDupls {
		if len(elems) <= 1 {
			continue
		}

		logger.Error("Found duplicate element specs",
			slog.String("qualified_name", name),
			slog.Int("count", len(elems)))

		primaries := make([]diagnostic.Annotation, len(elems))
		for i, elem := range elems {
			primaries[i] = anno.Node(elem.File, elem.AST.Name, "defined here")
		}

		l.report(&diagnostic.Diagnostic{
			Message: "multiple elements with same qualified name",
			Primary: primaries,
		})
	}

	for name, elems := range htmlNameDupls {
		if len(elems) <= 1 {
			continue
		}

		// Don't report again if exactly the same elements were already
		// reported for clashing qualified names.
		qualifiedElems := qualifiedDupls[strings.ToLower(elems[0].QualifiedName())]
		if slices.Equal(elems, qualifiedElems) {
			continue
		}

		logger.Error("Found duplicate element specs",
			slog.String("full_name", name),
			slog.Int("count", len(elems)))

		primaries := make([]diagnostic.Annotation, 0, 2*len(elems))
		reportedPrefixes := make(map[*ast.ElementDefinition]bool)
		for _, elem := range elems {
			primaries = appendElementSpecLocationAnnotations(primaries, reportedPrefixes, elem, "defined here")
		}

		l.report(&diagnostic.Diagnostic{
			Message: "multiple elements with same html name",
			Primary: primaries,
		})
	}
}

func appendElementSpecLocationAnnotations(
	annos []diagnostic.Annotation, reportedPrefixes map[*ast.ElementDefinition]bool, elem *file.ElementSpec, text string,
) []diagnostic.Annotation {
	if elem.Definition.LParen == nil && elem.Definition.Prefix != nil {
		return append(annos, anno.Range(elem.File, elem.Definition.Prefix.Start(), elem.AST.Name.End(), text))
	}

	if elem.Definition.Prefix != nil && (reportedPrefixes == nil || !reportedPrefixes[elem.Definition]) {
		if reportedPrefixes != nil {
			reportedPrefixes[elem.Definition] = true
		}
		annos = append(annos, anno.Node(elem.File, elem.Definition.Prefix, "with this prefix"))
	}
	return append(annos, anno.Node(elem.File, elem.AST.Name, text))
}

// ============================================================================
// Attribute Spec
// ======================================================================================

func (l *linker) CheckAttributeSpecCollisions() {
	logger := l.logger.WithGroup("checks.collisions.attribute_specs")

	if len(l.p.AttributeSpecs) <= 1 {
		logger.Info("One or no attribute spec, skipping")
		return
	}

	qualifiedDupls := make(map[attributeSelector][]*file.AttributeSpec)
	fullDupls := make(map[fullAttributeSelector][]*file.AttributeSpec)

	for _, attr := range l.p.AttributeSpecs {
		info := attrSpecInfo(attr)
		if info == nil {
			continue
		}

		sel := info.qualifiedSelector()
		qualifiedDupls[sel] = append(qualifiedDupls[sel], attr)

		fullSel := info.htmlNameSelector()
		if fullSel != sel {
			fullDupls[fullSel] = append(fullDupls[fullSel], attr)
		}
	}

	for sel, attrs := range qualifiedDupls {
		if len(attrs) <= 1 {
			continue
		}

		logger.Error("Found duplicate attribute specs",
			slog.String("selector", sel),
			slog.Int("count", len(attrs)))

		primaries := make([]diagnostic.Annotation, len(attrs))
		for i, attr := range attrs {
			primaries[i] = anno.Node(attr.File, attr.AST.Selector, "defined here")
		}

		l.report(&diagnostic.Diagnostic{
			Message: "multiple attributes with same qualified selector",
			Primary: primaries,
			Hints: []diagnostic.Hint{
				{Hint: "Remember that attribute names are case-insensitive."},
			},
		})
	}

	for sel, attrs := range fullDupls {
		if len(attrs) <= 1 {
			continue
		}

		logger.Error("Found duplicate attribute specs",
			slog.String("selector", sel),
			slog.Int("count", len(attrs)))

		primaries := make([]diagnostic.Annotation, 0, 2*len(attrs))
		reportedPrefixes := make(map[*ast.AttributeDefinition]bool)
		for _, attr := range attrs {
			primaries = appendAttrSpecLocationAnnotations(primaries, reportedPrefixes, attr, "defined here")
		}

		l.report(&diagnostic.Diagnostic{
			Message: "multiple attributes with same full selector",
			Primary: primaries,
			Hints: []diagnostic.Hint{
				{Hint: "Remember that attribute names are case-insensitive."},
			},
		})
	}
}

func appendAttrSpecLocationAnnotations(
	annos []diagnostic.Annotation, reportedPrefixes map[*ast.AttributeDefinition]bool, attr *file.AttributeSpec, text string,
) []diagnostic.Annotation {
	if attr.Definition.LParen == nil && attr.Definition.Prefix != nil {
		return append(annos, anno.Range(attr.File, attr.Definition.Prefix.Start(), attr.AST.Selector.End(), text))
	}

	if attr.Definition.Prefix != nil && (reportedPrefixes == nil || !reportedPrefixes[attr.Definition]) {
		if reportedPrefixes != nil {
			reportedPrefixes[attr.Definition] = true
		}
		annos = append(annos,
			anno.Node(attr.File, attr.Definition.Prefix, "with this prefix"))
	}
	return append(annos, anno.Node(attr.File, attr.AST.Selector, text))
}
