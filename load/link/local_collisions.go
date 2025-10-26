package link

import (
	"log/slog"
	"slices"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
)

// ============================================================================
// Component
// ======================================================================================

type componentCollisionCheck struct{}

func (l *linker) CheckComponentCollisions() {
	defer l.Ran(l.Pkg, componentCollisionCheck{})

	logger := l.Logger.WithGroup("checks.collisions.components")
	logger.Debug("Checking for component collisions")

	if len(l.Pkg.Components) <= 1 {
		logger.Info("One or no components, skipping")
		return
	}

	dupls := make(map[file.Identifier][]*file.Component)

	for _, comp := range l.Pkg.Components {
		if comp.Name != "" {
			dupls[comp.Name] = append(dupls[comp.Name], comp)
		}
	}

	for name, comps := range dupls {
		if len(comps) <= 1 {
			continue
		}

		logger.Error("Found component collisions",
			slog.String("name", string(name)),
			slog.Int("count", len(comps)))

		primaries := make([]diagnostic.Annotation, len(comps))
		for i, comp := range comps {
			primaries[i] = anno.Node(comp.File, comp.AST, "defined here")
		}

		l.Report(&diagnostic.Diagnostic{
			Message: "component defined multiple times",
			Primary: primaries,
		})
	}
}

// ============================================================================
// Element Spec
// ======================================================================================

type elementSpecCollisionCheck struct{}

func (l *linker) CheckElementSpecCollisions() {
	defer l.Ran(l.Pkg, elementSpecCollisionCheck{})

	logger := l.Logger.WithGroup("checks.collisions.element_specs")
	logger.Debug("Checking for element spec collisions")

	if len(l.Pkg.ElementSpecs) <= 1 {
		return
	}

	qualifiedDupls := make(map[file.CanonicalQualifiableElementName][]*file.ElementSpec)
	htmlNameDupls := make(map[file.CanonicalElementName][]*file.ElementSpec)

	// Collect all element specs
	for _, elem := range l.Pkg.ElementSpecs {
		if elem.Definition == nil || elem.AST.Name == nil {
			continue
		}

		if elem.QualifiableName == "" {
			continue
		}
		qualifiedDupls[elem.QualifiableName] = append(qualifiedDupls[elem.QualifiableName], elem)

		// don't report twice, if canonical and qualifiable name are the same
		if elem.HTMLName != file.CanonicalElementName(elem.QualifiableName) {
			htmlNameDupls[elem.HTMLName] = append(htmlNameDupls[elem.HTMLName], elem)
		}
	}

	for name, elems := range qualifiedDupls {
		if len(elems) <= 1 {
			continue
		}

		logger.Error("Found duplicate element specs",
			slog.String("qualified_name", string(name)),
			slog.Int("count", len(elems)))

		primaries := make([]diagnostic.Annotation, len(elems))
		for i, elem := range elems {
			primaries[i] = anno.Node(elem.File, elem.AST.Name, "defined here")
		}

		l.Report(&diagnostic.Diagnostic{
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
		qualifiedElems := qualifiedDupls[elems[0].QualifiableName]
		if slices.Equal(elems, qualifiedElems) {
			continue
		}

		logger.Error("Found duplicate element specs",
			slog.String("full_name", string(name)),
			slog.Int("count", len(elems)))

		primaries := make([]diagnostic.Annotation, 0, 2*len(elems))
		reportedPrefixes := make(map[*ast.ElementDefinition]bool)
		for _, elem := range elems {
			primaries = appendElementSpecLocationAnnotations(primaries, reportedPrefixes, elem, "defined here")
		}

		l.Report(&diagnostic.Diagnostic{
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

type attributeSpecCollisionCheck struct{}

func (l *linker) CheckAttributeSpecCollisions() {
	defer l.Ran(l.Pkg, attributeSpecCollisionCheck{})

	logger := l.Logger.WithGroup("checks.collisions.attribute_specs")
	logger.Debug("Checking for attribute spec collisions")

	if len(l.Pkg.AttributeSpecs) <= 1 {
		return
	}

	qualifiedDupls := make(map[file.CanonicalQualifiableAttributeName][]*file.AttributeSpec)
	htmlNameDupls := make(map[file.CanonicalAttributeName][]*file.AttributeSpec)

	for _, spec := range l.Pkg.AttributeSpecs {
		qualifiableSel := qualifiableAttrSelector(spec)
		htmlName := htmlAttrName(spec)
		if qualifiableSel == "" || htmlName == "" {
			continue
		}

		qualifiedDupls[qualifiableSel] = append(qualifiedDupls[qualifiableSel], spec)

		if htmlName != file.CanonicalAttributeName(qualifiableSel) {
			htmlNameDupls[htmlName] = append(htmlNameDupls[htmlName], spec)
		}
	}

	for sel, attrs := range qualifiedDupls {
		if len(attrs) <= 1 {
			continue
		}

		logger.Error("Found duplicate attribute specs",
			slog.String("selector", string(sel)),
			slog.Int("count", len(attrs)))

		primaries := make([]diagnostic.Annotation, len(attrs))
		for i, attr := range attrs {
			primaries[i] = anno.Node(attr.File, attr.AST.Selector, "defined here")
		}

		l.Report(&diagnostic.Diagnostic{
			Message: "multiple attributes with same qualified selector",
			Primary: primaries,
			Hints: []diagnostic.Hint{
				{Hint: "Remember that attribute names are case-insensitive."},
			},
		})
	}

	for sel, attrs := range htmlNameDupls {
		if len(attrs) <= 1 {
			continue
		}

		// Don't report again if exactly the same elements were already
		// reported for clashing qualified names.
		qualifiedElems := qualifiedDupls[qualifiableAttrSelector(attrs[0])]
		if slices.Equal(attrs, qualifiedElems) {
			continue
		}

		logger.Error("Found duplicate attribute specs",
			slog.String("selector", string(sel)),
			slog.Int("count", len(attrs)))

		primaries := make([]diagnostic.Annotation, 0, 2*len(attrs))
		reportedPrefixes := make(map[*ast.AttributeDefinition]bool)
		for _, attr := range attrs {
			primaries = appendAttrSpecLocationAnnotations(primaries, reportedPrefixes, attr, "defined here")
		}

		l.Report(&diagnostic.Diagnostic{
			Message: "multiple attributes with same html name selector",
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
