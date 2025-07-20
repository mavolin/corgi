package link

import (
	"context"
	"log/slog"
	"slices"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	"github.com/mavolin/corgi/v2/internal/set"
)

func (l *linker) CheckLocalDotImportCollisions(_ context.Context) {
	logger := l.logger.WithGroup("check.local_dot_import_collisions")
	logger.Info("Checking for collisions between local symbols and dot imports")

	dotImports := make([]*file.Import, 0, 32)

	for _, f := range l.p.Files {
		dotImports = dotImports[:0]
		for _, imp := range f.Symbols.Imports {
			if imp == nil || imp.AST == nil || imp.AST.Alias == nil || imp.AST.Alias.Name != "." || imp.Package == nil {
				continue
			} else if slices.Contains(dotImports, imp) {
				continue
			}
			dotImports = append(dotImports, imp)
		}

		(&localDotImportCollisionChecker{
			dotImports: dotImports,
		}).checkFile(l, logger, f)
	}
}

type localDotImportCollisionChecker struct { // file level
	dotImports []*file.Import
}

func (c *localDotImportCollisionChecker) checkFile(l *linker, logger *slog.Logger, f *file.File) {
	logger = logger.With(slog.String("file", f.Name))
	logger.Debug("Checking file")

	if len(c.dotImports) == 0 {
		logger.Debug("No dot imports, skipping")
		return
	}

	(&localDotImportComponentCollisionChecker{
		localDotImportCollisionChecker: c,
		duplComps:                      make([]localDotImportComponentCollision, 0, 8),
	}).checkFile(l, logger, f)
	(&localDotImportElementDefinitionCollisionChecker{
		localDotImportCollisionChecker: c,
		duplElemDefs:                   make([]localDotImportElementDefinitionCollision, 0, 8),
	}).checkFile(l, logger, f)
	(&localDotImportAttributeDefinitionCollisionChecker{
		localDotImportCollisionChecker: c,
		duplAttrDefs:                   make([]localDotImportAttributeDefinitionCollision, 0, 8),
	}).checkFile(l, logger, f)
}

// ================================ Component Collisions ================================

type (
	localDotImportComponentCollisionChecker struct { // file level
		*localDotImportCollisionChecker
		duplComps []localDotImportComponentCollision
	}

	localDotImportComponentCollision struct {
		imp  *file.Import
		comp *file.Component
	}
)

func (c *localDotImportComponentCollisionChecker) checkFile(l *linker, logger *slog.Logger, f *file.File) {
	logger = logger.WithGroup("components")
	logger.Info("Checking for collisions through components")

	for _, localComp := range l.p.Components {
		if !c.shouldCheck(localComp) {
			continue
		}

		localCompName := localComp.Header().Name.Name
		logger := logger.With(
			slog.String("pos", localComp.Start().String()),
			slog.String("component", localCompName))
		logger.Debug("Checking component")
		c.resetDuplicates()

		for _, imp := range c.dotImports {
			impComp := imp.Package.ComponentByName(localCompName)
			if impComp != nil {
				c.recordDuplicate(impComp, imp)
			}
		}

		if len(c.duplComps) > 0 {
			c.reportCollision(l, logger, f, localComp, c.duplComps)
		}
	}
}

func (c *localDotImportComponentCollisionChecker) reportCollision(l *linker, logger *slog.Logger, f *file.File, local *file.Component, dupls []localDotImportComponentCollision) {
	collisionName := local.Header().Name.Name
	logger.Error("Component collision", slog.String("name", collisionName))

	primaries := make([]diagnostic.Annotation, 1, len(dupls)+1)
	primaries[0] = anno.Node(f, local.Header().Name, "defined here")
	for _, dupl := range dupls {
		primaries = append(primaries, anno.Node(f, dupl.imp.AST, "also defines `"+collisionName+"`"))
	}

	secondaries := make([]diagnostic.Annotation, 0, len(dupls))
	for _, dupl := range dupls {
		secondaries = append(secondaries, anno.Anno(dupl.comp.File, anno.Annotation{
			Highlight:  anno.HighlightNode(dupl.comp.Header().Name),
			Context:    anno.ContextLines(dupl.comp.Start(), dupl.comp.End()),
			Annotation: "defined here",
		}))
	}

	l.report(&diagnostic.Diagnostic{
		Message:   "local dot import collision: a locally defined component collides with a dot-imported component",
		Primary:   primaries,
		Secondary: secondaries,
		Hints: []diagnostic.Hint{
			{Hint: "Make all of the imports non-dot imports."},
			{Hint: "Rename the local component."},
		},
	})
}

func (c *localDotImportComponentCollisionChecker) resetDuplicates() {
	c.duplComps = c.duplComps[:0]
}

func (c *localDotImportComponentCollisionChecker) recordDuplicate(comp *file.Component, imp *file.Import) {
	c.duplComps = append(c.duplComps, localDotImportComponentCollision{
		imp:  imp,
		comp: comp,
	})
}

func (c localDotImportComponentCollisionChecker) shouldCheck(comp *file.Component) bool {
	return comp != nil && comp.Header() != nil && comp.Header().Name != nil && file.IsExported(comp.Header().Name.Name)
}

// =========================== Element Spec Collisions ============================

type (
	localDotImportElementDefinitionCollisionChecker struct { // file level
		*localDotImportCollisionChecker
		duplElemDefs []localDotImportElementDefinitionCollision
	}
	localDotImportElementDefinitionCollision struct {
		imp  *file.Import
		elem *file.ElementSpec
	}
)

func (c *localDotImportElementDefinitionCollisionChecker) checkFile(l *linker, logger *slog.Logger, f *file.File) {
	logger = logger.WithGroup("element_definitions")
	logger.Info("Checking for collisions through element definitions")

	for _, localElemDef := range l.p.ElementDefinitions {
		localName := localElemDef.FullName()
		logger := logger.With(
			slog.String("pos", localElemDef.AST.Start().String()),
			slog.String("element", localName))
		logger.Debug("Checking element definition")
		c.resetDuplicates()

		for _, imp := range c.dotImports {
			impElemDef := imp.Package.ElementDefinitionByQualifiedName(localName)
			if impElemDef != nil {
				c.recordDuplicate(impElemDef, imp)
			}
		}

		if len(c.duplElemDefs) > 0 {
			c.reportCollision(l, logger, f, localElemDef, c.duplElemDefs)
		}
	}
}

func (c *localDotImportElementDefinitionCollisionChecker) reportCollision(l *linker, logger *slog.Logger, f *file.File, local *file.ElementSpec, dupls []localDotImportElementDefinitionCollision) {
	collisionName := local.FullName()
	logger.Error("Element definition collision", slog.String("name", collisionName))

	primaries := make([]diagnostic.Annotation, 0, len(dupls)+2)
	reportedPrefixes := set.NewSliceSet[*ast.ElementDefinition](len(dupls) + 1)
	primaries = c.appendCollisionDiagnostic(primaries, local, reportedPrefixes)
	for _, dupl := range dupls {
		primaries = append(primaries, anno.Node(f, dupl.imp.AST, "also defines `"+collisionName+"`"))
	}

	secondaries := make([]diagnostic.Annotation, 0, 2*len(dupls))
	for _, dupl := range dupls {
		secondaries = c.appendCollisionDiagnostic(secondaries, dupl.elem, reportedPrefixes)
	}

	l.report(&diagnostic.Diagnostic{
		Message:   "local dot import collision: a locally defined element collides with a dot-imported element",
		Primary:   primaries,
		Secondary: secondaries,
		Hints: []diagnostic.Hint{
			{Hint: "Make all of the imports non-dot imports."},
		},
	})
}

func (c localDotImportElementDefinitionCollisionChecker) appendCollisionDiagnostic(as []diagnostic.Annotation, dupl *file.ElementSpec, reportedPrefixes *set.SliceSet[*ast.ElementDefinition]) []diagnostic.Annotation {
	if dupl.Definition.LParen == nil && dupl.Definition.Prefix != nil {
		return append(as,
			anno.Range(dupl.File, dupl.Definition.Prefix.Start(), dupl.AST.Name.End(), "defined here"))
	}

	if dupl.Definition.Prefix != nil && reportedPrefixes.Contains(dupl.Definition) {
		reportedPrefixes.Add(dupl.Definition)
		as = append(as, anno.Node(dupl.File, dupl.Definition.Prefix, "with this prefix"))
	}
	return append(as, anno.Node(dupl.File, dupl.AST.Name, "defined here"))
}

func (c *localDotImportElementDefinitionCollisionChecker) resetDuplicates() {
	c.duplElemDefs = c.duplElemDefs[:0]
}

func (c *localDotImportElementDefinitionCollisionChecker) recordDuplicate(elem *file.ElementSpec, imp *file.Import) {
	c.duplElemDefs = append(c.duplElemDefs, localDotImportElementDefinitionCollision{
		imp:  imp,
		elem: elem,
	})
}

func (c localDotImportElementDefinitionCollisionChecker) shouldCheck(elem *file.ElementSpec) bool {
	return elem != nil && elem.AST != nil && elem.FullName() != ""
}

// ========================== Attribute Spec Collisions ===========================

type (
	localDotImportAttributeDefinitionCollisionChecker struct { // file level
		*localDotImportCollisionChecker
		duplAttrDefs []localDotImportAttributeDefinitionCollision
	}
	localDotImportAttributeDefinitionCollision struct {
		imp      *file.Import
		attr     *file.AttributeSpec
		selector string
	}
)

func (c *localDotImportAttributeDefinitionCollisionChecker) checkFile(l *linker, logger *slog.Logger, f *file.File) {
	logger = logger.WithGroup("attribute_definition")
	logger.Info("Checking for collisions through attribute definitions")

	for _, localAttrDef := range l.p.AttributeDefinitions {
		aSel := c.basicSelector(localAttrDef)
		if aSel == nil {
			continue
		}

		aName, aSelector := c.attrName(localAttrDef)

		logger := logger.With(
			slog.String("pos", localAttrDef.AST.Start().String()),
			slog.String("selector", aSelector))
		logger.Debug("Checking attribute definition")

		c.resetDuplicates()

		for _, imp := range c.dotImports {
			for _, impAttrDef := range imp.Package.AttributeDefinitions {
				impSel := c.basicSelector(impAttrDef)
				if impSel == nil {
					continue
				}

				impName, impSelector := c.attrName(impAttrDef)

				if aSel.Wildcard {
					if impSel.Wildcard {
						if aName != impName {
							continue
						}
					} else {
						if !localAttrDef.MatchesFullName(impName) {
							continue
						}
					}
				} else {
					if impSel.Wildcard {
						if !impAttrDef.MatchesFullName(aName) {
							continue
						}
					} else {
						if aName != impName {
							continue
						}
					}
				}

				c.recordDuplicate(imp, impAttrDef, impSelector)
				break
			}
		}

		if len(c.duplAttrDefs) > 0 {
			a := localDotImportAttributeDefinitionCollision{
				attr:     localAttrDef,
				selector: aSelector,
			}
			c.reportCollision(l, logger, f, a, c.duplAttrDefs)
		}
	}
}

func (c *localDotImportAttributeDefinitionCollisionChecker) reportCollision(l *linker, logger *slog.Logger, f *file.File, local localDotImportAttributeDefinitionCollision, dupls []localDotImportAttributeDefinitionCollision) {
	logger.Error("Attribute definition collision", slog.String("selector", local.selector))

	primaries := make([]diagnostic.Annotation, 0, len(dupls)+2)
	reportedPrefixes := set.NewSliceSet[*ast.AttributeDefinition](len(dupls) + 1)
	primaries = c.appendCollisionDiagnostic(primaries, local.attr, reportedPrefixes)
	for _, dupl := range dupls {
		primaries = append(primaries, anno.Node(f, dupl.imp.AST, "also defines `"+dupl.selector+"`"))
	}

	secondaries := make([]diagnostic.Annotation, 0, 2*len(dupls))
	for _, dupl := range dupls {
		secondaries = c.appendCollisionDiagnostic(secondaries, dupl.attr, reportedPrefixes)
	}

	l.report(&diagnostic.Diagnostic{
		Message:   "local dot import collision: a locally defined attribute collides with a dot-imported attribute",
		Primary:   primaries,
		Secondary: secondaries,
		Hints: []diagnostic.Hint{
			{Hint: "Make all but one of the imports non-dot imports."},
		},
	})
}

func (c localDotImportAttributeDefinitionCollisionChecker) appendCollisionDiagnostic(secondaries []diagnostic.Annotation, attr *file.AttributeSpec, reportedPrefixes *set.SliceSet[*ast.AttributeDefinition]) []diagnostic.Annotation {
	if attr.Definition.LParen == nil && attr.Definition.Prefix != nil {
		return append(secondaries,
			anno.Range(attr.File, attr.Definition.Prefix.Start(), attr.AST.Selector.End(), "defined here"))
	}

	if attr.Definition.Prefix != nil && reportedPrefixes.Contains(attr.Definition) {
		reportedPrefixes.Add(attr.Definition)
		secondaries = append(secondaries, anno.Node(attr.File, attr.Definition.Prefix, "with this prefix"))
	}
	return append(secondaries, anno.Node(attr.File, attr.AST.Selector, "defined here"))
}

func (c *localDotImportAttributeDefinitionCollisionChecker) resetDuplicates() {
	c.duplAttrDefs = c.duplAttrDefs[:0]
}

func (c *localDotImportAttributeDefinitionCollisionChecker) recordDuplicate(imp *file.Import, attr *file.AttributeSpec, selector string) {
	c.duplAttrDefs = append(c.duplAttrDefs, localDotImportAttributeDefinitionCollision{
		imp:      imp,
		attr:     attr,
		selector: selector,
	})
}

func (c localDotImportAttributeDefinitionCollisionChecker) basicSelector(attr *file.AttributeSpec) *ast.BasicAttributeSelector {
	if attr == nil || attr.AST == nil {
		return nil
	}
	sel, _ := attr.AST.Selector.(*ast.BasicAttributeSelector)
	if sel == nil || sel.Name == "" {
		return nil
	}
	return sel
}

func (c localDotImportAttributeDefinitionCollisionChecker) attrName(attr *file.AttributeSpec) (name, selector string) {
	sel := attr.AST.Selector.(*ast.BasicAttributeSelector)
	if attr.Definition != nil && attr.Definition.Prefix != nil {
		name = attr.Definition.Prefix.Name + sel.Name
	} else {
		name = sel.Name
	}

	if sel.Wildcard {
		selector = name + "*"
	} else {
		selector = name
	}

	return name, selector
}
