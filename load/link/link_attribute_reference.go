package link

import (
	"fmt"
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
)

const ambiguousAttributeReferenceExplanation = "There are multiple regular expression selectors that all match this attribute and " +
	"therefore it is unclear which one to use. " +
	"Refine your regular expressions so that only one matches to resolve this ambiguity."

type attributeReferenceLinked struct{}

func (l *linker) LinkAttributeReferences() {
	defer l.Ran(l.Pkg, attributeReferenceLinked{})
	l.Require(l.Pkg, importsLoaded{})
	l.Require(l.Pkg, dotImportAttributeSpecCollisionCheck{})
	l.Require(l.Pkg, attributeSpecCollisionCheck{})
	l.Require(l.Pkg, attributeRuleCollisionCheck{})

	logger := l.Logger.WithGroup("links.attribute_references")
	logger.Debug("Linking attribute references")

	for _, f := range l.Pkg.Files {
		logger := logger.With(slog.String("file", string(f.Name)))

		for _, ref := range f.AttributeReferences {
			ref.Linked = true
			if ref.AST.Name == nil {
				continue
			}

			var name string
			if ref.AST.Package != nil {
				name = ref.AST.Package.Name + "." + ref.AST.Name.Name
			} else {
				name = ref.AST.Name.Name
			}

			logger := logger.With(
				slog.String("name", name),
				slog.String("pos", ref.AST.Start().String()))

			if ref.Qualified() {
				l.linkQualifiedAttributeReference(logger, f, ref)
				if ref.Spec.NotZero() {
					ref.HTMLName.SetResult(ref.Spec.Result().Prefix + file.CanonicalAttributeName(ref.AST.Name.Name))
				}
			} else {
				l.linkUnqualifiedAttributeReference(logger, f, ref)
				if ref.Spec.NotZero() {
					ref.HTMLName.SetResult(file.CanonicalAttributeName(ref.AST.Name.CanonicalName))
				}
			}
		}
	}
}

func (l *linker) linkUnqualifiedAttributeReference(logger *slog.Logger, f *file.File, ref *file.AttributeReference) {
	var (
		equalSpecificityMatches []*file.AttributeSpec
		bestImport              *file.Import
	)

	// search in current package
	equalSpecificityMatches = f.Package.AttributeSpecByHTMLName(ref.UnqualifiedName)

	// search in dot imports
	var ignoreError bool
	for _, imp := range l.dotImports[f] {
		if imp.Package == nil || imp.Package.PackageSymbols == nil {
			ignoreError = true
		}
		switch {
		case !imp.Explicit() || imp.Qualifier != "":
			continue
		case imp.Package == nil || imp.Package.PackageSymbols == nil:
			continue
		}

		packageMatches := imp.Package.AttributeSpecByHTMLName(ref.UnqualifiedName)
		if packageMatches == nil {
			continue
		}

		if len(equalSpecificityMatches) == 0 || equalSpecificityMatches[0].Specificity < packageMatches[0].Specificity {
			equalSpecificityMatches = packageMatches
			bestImport = imp
		} else if equalSpecificityMatches[0].Specificity == packageMatches[0].Specificity {
			equalSpecificityMatches = append(equalSpecificityMatches, packageMatches...)
		}
	}

	if len(equalSpecificityMatches) == 1 {
		ref.Spec.SetResult(equalSpecificityMatches[0])
		if bestImport != nil {
			bestImport.Forward = true
		}
		return
	} else if len(equalSpecificityMatches) > 1 {
		ref.Spec.SetFailed()
		if ignoreError {
			logger.Debug("Found multiple attribute definitions with same specificity, but at least one dot import was not loaded without error: not reporting error")
			return
		}

		logger.Error("Found multiple attribute definitions with same specificity",
			slog.Int("count", len(equalSpecificityMatches)),
			slog.Int("specificity", equalSpecificityMatches[0].Specificity))

		l.Report(&diagnostic.Diagnostic{
			Message: "attribute: ambiguous reference",
			Primary: []diagnostic.Annotation{
				anno.Node(f, ref.AST, "there are multiple attribute selectors with the same specificity that match this attribute"),
			},
			Secondary:   equalSpecificityAnnotations(equalSpecificityMatches),
			Explanation: ambiguousAttributeReferenceExplanation,
			Docs:        "attribute-specificity",
		})
		return
	}

	// search in builtin package
	builtinImp := f.BuiltinImport()
	if builtinImp == nil || builtinImp.Package == nil || builtinImp.Package.PackageSymbols == nil {
		// Not defined explicitly, analyzer needs to determine whether there is explicit typing
		ref.Spec.SetZero()
		return
	}

	packageMatches := builtinImp.Package.AttributeSpecByHTMLName(ref.UnqualifiedName)
	if len(packageMatches) == 1 {
		ref.Spec.SetResult(packageMatches[0])
		builtinImp.Forward = true
		return
	} else if len(packageMatches) > 1 {
		ref.Spec.SetFailed()
		logger.Error("Found multiple attribute definitions with same specificity in builtin package")
		l.Report(&diagnostic.Diagnostic{
			Type:    diagnostic.InternalError,
			Message: "builtin: attribute: ambiguous reference",
			Primary: []diagnostic.Annotation{
				anno.Node(f, ref.AST, "there are multiple attribute selectors with the same specificity that match this attribute"),
			},
			Secondary:   equalSpecificityAnnotations(packageMatches),
			Explanation: ambiguousAttributeReferenceExplanation,
			Hints: []diagnostic.Hint{
				{
					Hint: "If you are using the corgi stdlib builtin package, you shouldn't see this error. " +
						"Please open an issue, this is a bug.",
				},
			},
			Docs: "attribute-specificity",
		})
		return
	}

	ref.Spec.SetZero()
}

func (l *linker) linkQualifiedAttributeReference(logger *slog.Logger, f *file.File, ref *file.AttributeReference) {
	if ref.Qualifier == "" {
		return
	}

	imp := f.ImportByQualifier(ref.Qualifier)
	if imp == nil {
		ref.Spec.SetFailed()
		logger.Error("Could not find import for package")
		l.ReportMissingImport(f, ref.Qualifier, &diagnostic.Diagnostic{
			Message: "attribute: unresolved reference to package",
			Primary: []diagnostic.Annotation{
				anno.Node(f, ref.AST.Package, "missing import for this package"),
			},
		})
		return
	}

	if !l.implicitImportCheck(logger, f, imp, ref.AST.Package, "an", "attribute") {
		ref.Spec.SetFailed()
		return
	}

	var matches []*file.AttributeSpec
	if imp.Package != nil && imp.Package.PackageSymbols != nil {
		matches = imp.Package.AttributeSpecByQualifiableName(ref.QualifiableName)
		if len(matches) == 1 {
			ref.Spec.SetResult(matches[0])
			imp.Forward = true
			return
		}
	}

	ref.Spec.SetFailed()

	if imp.Package == nil || imp.Package.PackageSymbols == nil {
		logger.Debug("Couldn't resolve reference, but package was loaded with errors: not reporting error")
		return
	}

	if len(matches) == 0 {
		logger.Error("Could not resolve reference")
		l.Report(&diagnostic.Diagnostic{
			Message: "attribute: unresolved reference",
			Primary: []diagnostic.Annotation{
				anno.Node(f, ref.AST, "could not resolve reference"),
			},
			Secondary: []diagnostic.Annotation{
				anno.Node(f, imp.AST, "searching in this package"),
			},
		})
		return
	}
	logger.Error("Found multiple attribute definitions with same specificity",
		slog.Int("count", len(matches)),
		slog.Int("specificity", matches[0].Specificity))
	l.Report(&diagnostic.Diagnostic{
		Message: "attribute: ambiguous reference",
		Primary: []diagnostic.Annotation{
			anno.Node(f, ref.AST,
				"there are multiple attribute selectors with the same specificity that match this attribute"),
		},
		Secondary:   equalSpecificityAnnotations(matches),
		Explanation: ambiguousAttributeReferenceExplanation,
		Docs:        "attribute-specificity",
	})
}

func equalSpecificityAnnotations(defs []*file.AttributeSpec) []diagnostic.Annotation {
	reportedPrefixes := make(map[*ast.AttributeDefinition]bool)
	as := make([]diagnostic.Annotation, 0, 2*len(defs))
	for _, def := range defs {
		if def.Definition.Prefix != nil && !reportedPrefixes[def.Definition] {
			reportedPrefixes[def.Definition] = true
			as = append(as, anno.Node(def.File, def.Definition.Prefix, "with this prefix"))
		}
		as = append(as,
			anno.Node(def.File, def.AST.Selector, fmt.Sprint("this selector has a specificity of ", def.Specificity)))
	}
	return as
}
