package link

import (
	"context"
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	"github.com/mavolin/corgi/v2/internal/set"
)

func (l *linker) CheckDotImportCollisions(_ context.Context) {
	logger := l.logger.WithGroup("check.dot_import_collisions")
	logger.Info("Checking for dot import collisions")

	dotImports := make([]*file.Import, 0, 32)

	for _, f := range l.p.Files {
		dotImports = dotImports[:0]
		for _, imp := range f.Symbols.Imports {
			if imp == nil || imp.AST == nil || imp.AST.Alias == nil || imp.AST.Alias.Ident != "." || imp.Package == nil {
				continue
			}
			dotImports = append(dotImports, imp)
		}

		(&dotImportCollisionChecker{
			dotImports: dotImports,
		}).checkFile(l, logger, f)
	}
}

type dotImportCollisionChecker struct { // file level
	dotImports []*file.Import
}

func (c *dotImportCollisionChecker) checkFile(l *linker, logger *slog.Logger, f *file.File) {
	logger = logger.With(slog.String("file", f.Name))
	logger.Debug("Checking file")

	if len(c.dotImports) <= 1 {
		logger.Debug("One or no dot imports, skipping")
		return
	}

	(&dotImportComponentCollisionChecker{
		dotImportCollisionChecker: c,
		checked:                   l.takeStringSet(),
		duplComps:                 make([]dotImportComponentCollision, 0, 8),
	}).checkFile(l, logger, f)
	(&dotImportElementDefinitionCollisionChecker{
		dotImportCollisionChecker: c,
		checked:                   l.takeStringSet(),
		duplElemDefs:              make([]dotImportElementDefinitionCollision, 0, 8),
	}).checkFile(l, logger, f)
	(&dotImportAttributeDefinitionCollisionChecker{
		dotImportCollisionChecker: c,
		reported:                  set.NewSliceSet[*file.AttributeDefinition](32),
		duplAttrDefs:              make([]dotImportAttributeDefinitionCollision, 0, 8),
	}).checkFile(l, logger, f)
}

// ============================================================================
// Component Collisions
// ======================================================================================

type (
	dotImportComponentCollisionChecker struct { // file level
		*dotImportCollisionChecker
		checked   set.Set[componentName]
		duplComps []dotImportComponentCollision
	}

	dotImportComponentCollision struct {
		imp  *file.Import
		comp *file.Component
	}
)

func (c *dotImportComponentCollisionChecker) checkFile(l *linker, logger *slog.Logger, f *file.File) {
	logger = logger.WithGroup("components")
	logger.Debug("Checking for dot import collisions through components")

	for aImpI, aImp := range c.dotImports[:len(c.dotImports)-1] {
		logger := logger.With(slog.String("import", aImp.Package.ImportPath))
		logger.Debug("Checking import")

		for _, aComp := range aImp.Package.Components {
			if !c.shouldCheck(aComp) {
				continue
			}

			aCompName := aComp.AST.Header.Name.Ident
			logger := logger.With(
				slog.String("component_pos", aComp.AST.Start().String()),
				slog.String("component", aCompName))
			logger.Debug("Checking component")
			if c.checked.Contains(aCompName) {
				logger.Debug("Already checked components with that name")
				continue
			}
			c.checked.Add(aCompName)
			c.resetDuplicates()

			for _, bImp := range c.dotImports[aImpI+1:] {
				if aImp.Package.ImportPath == bImp.Package.ImportPath { // duplicate import, reported elsewhere
					continue
				}

				bComp := bImp.Package.ComponentByName(aCompName)
				if bComp != nil {
					c.recordDuplicate(bComp, bImp)
				}
			}

			if len(c.duplComps) > 0 {
				a := dotImportComponentCollision{
					imp:  aImp,
					comp: aComp,
				}
				c.reportCollision(l, logger, f, a, c.duplComps)
			}
		}
	}
}

func (c *dotImportComponentCollisionChecker) reportCollision(l *linker, logger *slog.Logger, f *file.File, first dotImportComponentCollision, dupls []dotImportComponentCollision) {
	collisionName := first.comp.AST.Header.Name.Ident
	logger.Error("Component collision", slog.String("name", collisionName))

	primaries := make([]diagnostic.Annotation, 1, len(dupls)+1)
	primaries[0] = anno.Node(f, first.imp.AST, "defines `"+collisionName+"`")
	for _, dupl := range dupls {
		primaries = append(primaries, anno.Node(f, dupl.imp.AST, "defines `"+collisionName+"`"))
	}

	secondaries := make([]diagnostic.Annotation, 1, len(dupls)+1)
	secondaries[0] = anno.Anno(first.comp.File, anno.Annotation{
		Highlight:  anno.HighlightNode(first.comp.AST.Header.Name),
		Context:    anno.ContextLines(first.comp.AST.Start(), first.comp.AST.End()),
		Annotation: "defined here",
	})
	for _, dupl := range dupls {
		secondaries = append(secondaries, anno.Anno(dupl.comp.File, anno.Annotation{
			Highlight:  anno.HighlightNode(dupl.comp.AST.Header.Name),
			Context:    anno.ContextLines(dupl.comp.AST.Start(), dupl.comp.AST.End()),
			Annotation: "defined here",
		}))
	}

	l.report(&diagnostic.Diagnostic{
		Message:   "dot import collision: multiple definitions for component of the same name",
		Primary:   primaries,
		Secondary: secondaries,
		Hints: []diagnostic.Hint{
			{Hint: "Make all but one of the imports non-dot imports."},
		},
	})
}

func (c *dotImportComponentCollisionChecker) resetDuplicates() {
	c.duplComps = c.duplComps[:0]
}

func (c *dotImportComponentCollisionChecker) recordDuplicate(comp *file.Component, imp *file.Import) {
	c.duplComps = append(c.duplComps, dotImportComponentCollision{
		imp:  imp,
		comp: comp,
	})
}

func (c dotImportComponentCollisionChecker) shouldCheck(comp *file.Component) bool {
	return comp != nil && comp.AST.Header != nil && comp.AST.Header.Name != nil && file.IsExported(comp.AST.Header.Name.Ident)
}

// ============================================================================
// Element Definition Collisions
// ======================================================================================

type (
	dotImportElementDefinitionCollisionChecker struct { // file level
		*dotImportCollisionChecker
		checked      set.Set[fullElementName]
		duplElemDefs []dotImportElementDefinitionCollision
	}
	dotImportElementDefinitionCollision struct {
		imp  *file.Import
		elem *file.ElementDefinition
	}
)

func (c *dotImportElementDefinitionCollisionChecker) checkFile(l *linker, logger *slog.Logger, f *file.File) {
	logger = logger.WithGroup("element_definitions")
	logger.Debug("Checking for dot import collisions through element definitions")

	for aImpI, aImp := range c.dotImports[:len(c.dotImports)-1] {
		logger := logger.With(slog.String("import", aImp.Package.ImportPath))
		logger.Debug("Checking import")

		for _, aElemDef := range aImp.Package.ElementDefinitions {
			if !c.shouldCheck(aElemDef) {
				continue
			}

			aName := aElemDef.FullName()
			logger := logger.With(
				slog.String("spec_file", aElemDef.File.Name),
				slog.String("spec_pos", aElemDef.AST.Start().String()),
				slog.String("element", aName))
			logger.Debug("Checking element definition spec")
			if c.checked.Contains(aName) {
				logger.Debug("Already checked element definitions with that name")
				continue
			}
			c.checked.Add(aName)
			c.resetDuplicates()

			for _, bImp := range c.dotImports[aImpI+1:] {
				if aImp.Package.ImportPath == bImp.Package.ImportPath { // duplicate import, reported elsewhere
					continue
				}

				bElemDef := bImp.Package.ElementDefinitionByFullName(aName)
				if bElemDef != nil {
					c.recordDuplicate(bElemDef, bImp)
				}
			}

			if len(c.duplElemDefs) > 0 {
				a := dotImportElementDefinitionCollision{
					imp:  aImp,
					elem: aElemDef,
				}
				c.reportCollision(l, logger, f, a, c.duplElemDefs)
			}
		}
	}
}

func (c *dotImportElementDefinitionCollisionChecker) reportCollision(l *linker, logger *slog.Logger, f *file.File, first dotImportElementDefinitionCollision, dupls []dotImportElementDefinitionCollision) {
	collisionName := first.elem.FullName()
	logger.Error("Element definition collision", slog.String("name", collisionName))

	primaries := make([]diagnostic.Annotation, 1, len(dupls)+1)
	primaries[0] = anno.Node(f, first.imp.AST, "defines `"+collisionName+"`")
	for _, dupl := range dupls {
		primaries = append(primaries, anno.Node(f, dupl.imp.AST, "defines `"+collisionName+"`"))
	}

	secondaries := make([]diagnostic.Annotation, 0, 2*(len(dupls)+1))
	reportedPrefixes := set.NewSliceSet[*ast.ElementDefinition](len(dupls) + 1)
	secondaries = c.appendCollisionDiagnosticSecondary(secondaries, first, reportedPrefixes)
	for _, dupl := range dupls {
		secondaries = c.appendCollisionDiagnosticSecondary(secondaries, dupl, reportedPrefixes)
	}

	l.report(&diagnostic.Diagnostic{
		Message:   "dot import collision: multiple definitions for element of the same name",
		Primary:   primaries,
		Secondary: secondaries,
		Hints: []diagnostic.Hint{
			{Hint: "Make all but one of the imports non-dot imports."},
			{Hint: "Remember that element definitions are case-insensitive."},
		},
	})
}

func (c dotImportElementDefinitionCollisionChecker) appendCollisionDiagnosticSecondary(secondaries []diagnostic.Annotation, dupl dotImportElementDefinitionCollision, reportedPrefixes *set.SliceSet[*ast.ElementDefinition]) []diagnostic.Annotation {
	if dupl.elem.Definition.LParen == nil && dupl.elem.Definition.Prefix != nil {
		return append(secondaries,
			anno.Range(dupl.elem.File, dupl.elem.Definition.Prefix.Start(), dupl.elem.AST.Name.End(), "defined here"))
	}

	if dupl.elem.Definition.Prefix != nil && reportedPrefixes.Add(dupl.elem.Definition) {
		secondaries = append(secondaries,
			anno.Node(dupl.elem.File, dupl.elem.Definition.Prefix, "with this prefix"))
	}
	return append(secondaries, anno.Node(dupl.elem.File, dupl.elem.AST.Name, "defined here"))
}

func (c *dotImportElementDefinitionCollisionChecker) resetDuplicates() {
	c.duplElemDefs = c.duplElemDefs[:0]
}

func (c *dotImportElementDefinitionCollisionChecker) recordDuplicate(elem *file.ElementDefinition, imp *file.Import) {
	c.duplElemDefs = append(c.duplElemDefs, dotImportElementDefinitionCollision{
		imp:  imp,
		elem: elem,
	})
}

func (c dotImportElementDefinitionCollisionChecker) shouldCheck(elem *file.ElementDefinition) bool {
	return elem != nil && elem.AST != nil && elem.FullName() != ""
}

// ============================================================================
// Attribute Definition Collisions
// ======================================================================================

type (
	dotImportAttributeDefinitionCollisionChecker struct { // file level
		*dotImportCollisionChecker
		reported     set.Set[*file.AttributeDefinition]
		duplAttrDefs []dotImportAttributeDefinitionCollision
	}
	dotImportAttributeDefinitionCollision struct {
		imp      *file.Import
		attr     *file.AttributeDefinition
		selector string
	}
)

func (c *dotImportAttributeDefinitionCollisionChecker) checkFile(l *linker, logger *slog.Logger, f *file.File) {
	logger = logger.WithGroup("attribute_definitions")
	logger.Debug("Checking for dot import collisions through attribute definitions")

	for aImpI, aImp := range c.dotImports[:len(c.dotImports)-1] {
		logger := logger.With(slog.String("import", aImp.Package.ImportPath))
		logger.Debug("Checking import")

		for _, aAttrDef := range aImp.Package.AttributeDefinitions {
			aInfo := attrDefInfo(aAttrDef)
			if aInfo == nil {
				continue
			}

			logger := logger.With(
				slog.String("spec_file", aAttrDef.File.Name),
				slog.String("spec_pos", aAttrDef.AST.Start().String()),
				slog.String("full_selector", aInfo.fullSelector()))
			logger.Debug("Checking attribute definition spec")

			if c.reported.Contains(aAttrDef) {
				logger.Debug("Already reported, skipping")
				continue
			}

			c.resetDuplicates()

			for _, bImp := range c.dotImports[aImpI+1:] {
				if aImp.Package.ImportPath == bImp.Package.ImportPath { // duplicate import, reported elsewhere
					continue
				}

				for _, bAttrDef := range bImp.Package.AttributeDefinitions {
					bInfo := attrDefInfo(bAttrDef)
					if bInfo == nil {
						continue
					}

					// only report collisions for attribute definitions with
					// the same specificity
					if aInfo.fullSelector() == bInfo.fullSelector() {
						c.recordDuplicate(bImp, bAttrDef, bInfo.fullSelector())
						c.reported.Add(bAttrDef)
						break // one per file
					}
				}
			}

			if len(c.duplAttrDefs) > 0 {
				a := dotImportAttributeDefinitionCollision{
					imp:      aImp,
					attr:     aAttrDef,
					selector: aInfo.fullSelector(),
				}
				c.reportCollision(l, logger, f, a, c.duplAttrDefs)
			}
		}
	}
}

func (c *dotImportAttributeDefinitionCollisionChecker) reportCollision(l *linker, logger *slog.Logger, f *file.File, first dotImportAttributeDefinitionCollision, dupls []dotImportAttributeDefinitionCollision) {
	logger.Error("Attribute definition collision", slog.String("selector", first.selector))

	primaries := make([]diagnostic.Annotation, 1, len(dupls)+1)
	primaries[0] = anno.Node(f, first.imp.AST, "defines `"+first.selector+"`")
	for _, dupl := range dupls {
		primaries = append(primaries, anno.Node(f, dupl.imp.AST, "defines `"+dupl.selector+"`"))
	}

	secondaries := make([]diagnostic.Annotation, 0, 2*(len(dupls)+1))
	secondaries = c.appendCollisionDiagnosticSecondary(secondaries, first)
	for _, dupl := range dupls {
		secondaries = c.appendCollisionDiagnosticSecondary(secondaries, dupl)
	}

	l.report(&diagnostic.Diagnostic{
		Message:   "dot import collision: multiple definitions for attribute of the same name",
		Primary:   primaries,
		Secondary: secondaries,
		Hints: []diagnostic.Hint{
			{Hint: "Make all but one of the imports non-dot imports."},
			{Hint: "Remember that attribute names are case-insensitive."},
		},
	})
}

func (c dotImportAttributeDefinitionCollisionChecker) appendCollisionDiagnosticSecondary(secondaries []diagnostic.Annotation, dupl dotImportAttributeDefinitionCollision) []diagnostic.Annotation {
	if dupl.attr.Definition.LParen == nil && dupl.attr.Definition.Prefix != nil {
		return append(secondaries,
			anno.Range(dupl.attr.File, dupl.attr.Definition.Prefix.Start(), dupl.attr.AST.Selector.End(), "defined here"))
	}

	if dupl.attr.Definition.Prefix != nil {
		secondaries = append(secondaries,
			anno.Node(dupl.attr.File, dupl.attr.Definition.Prefix, "with this prefix"))
	}
	return append(secondaries, anno.Node(dupl.attr.File, dupl.attr.AST.Selector, "defined here"))
}

func (c *dotImportAttributeDefinitionCollisionChecker) resetDuplicates() {
	c.duplAttrDefs = c.duplAttrDefs[:0]
}

func (c *dotImportAttributeDefinitionCollisionChecker) recordDuplicate(imp *file.Import, attr *file.AttributeDefinition, selector string) {
	c.duplAttrDefs = append(c.duplAttrDefs, dotImportAttributeDefinitionCollision{
		imp:      imp,
		attr:     attr,
		selector: selector,
	})
}
