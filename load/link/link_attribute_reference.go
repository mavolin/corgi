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

func (l *linker) LinkAttributeReferences() {
	logger := l.logger.WithGroup("links.attribute_references")
	logger.Debug("Linking attribute references")

	for _, f := range l.p.Files {
		logger := logger.With(slog.String("file", f.Name))

		for _, ref := range f.AttributeReferences {
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

			if ref.AST.Package != nil {
				l.linkQualifiedAttributeReference(logger, f, ref)
			} else {
				l.linkUnqualifiedAttributeReference(logger, f, ref)
			}
		}
	}
}

func (l *linker) linkUnqualifiedAttributeReference(logger *slog.Logger, f *file.File, ref *file.AttributeReference) {
	name := ref.AST.Name.Name

	var (
		bestMatchImport         string
		equalSpecificityMatches []*file.AttributeSpec
	)

	// search in current package
	packageMatches := f.Package.AttributeSpecByFullName(name)
	equalSpecificityMatches = packageMatches

	// search in dot imports
	var ignoreError bool
	for _, imp := range f.Imports {
		switch {
		case !imp.Explicit() || imp.Namespace != ".":
			continue
		case imp.LoadedWithErrors:
			ignoreError = true
			continue
		case imp.Package == nil:
			continue
		}

		packageMatches = imp.Package.AttributeSpecByFullName(name)
		if len(equalSpecificityMatches) == 0 || equalSpecificityMatches[0].Specificity < packageMatches[0].Specificity {
			equalSpecificityMatches = packageMatches
			bestMatchImport = imp.Path
		} else if equalSpecificityMatches[0].Specificity == packageMatches[0].Specificity {
			equalSpecificityMatches = append(equalSpecificityMatches, packageMatches...)
		}
	}

	if len(equalSpecificityMatches) == 1 {
		ref.Spec = equalSpecificityMatches[0]
		if bestMatchImport == "" {
			logger.Debug("Found attribute definition within package")
		} else {
			logger.Debug("Found attribute definition within package or dot import",
				slog.String("import", bestMatchImport))
		}
		return
	} else if len(equalSpecificityMatches) > 1 {
		if ignoreError {
			logger.Warn("Found multiple attribute definitions with same specificity, but at least one dot import was not loaded; not reporting error")
			return
		}

		logger.Error("Found multiple attribute definitions with same specificity",
			slog.Int("count", len(equalSpecificityMatches)),
			slog.Int("specificity", equalSpecificityMatches[0].Specificity))

		l.report(&diagnostic.Diagnostic{
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
	if builtinImp := f.BuiltinImport(); builtinImp != nil {
		packageMatches = builtinImp.Package.AttributeSpecByFullName(name)
		if len(packageMatches) == 1 {
			logger.Debug("Found attribute definition within builtin package")
			ref.Spec = packageMatches[0]
		} else if len(packageMatches) > 1 {
			logger.Error("Found multiple attribute definitions with same specificity in builtin package")
			l.report(&diagnostic.Diagnostic{
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
		}
		return
	}

	logger.Debug("Not defined explicitly, analyzer needs to determine whether there is explicit typing")
}

func (l *linker) linkQualifiedAttributeReference(logger *slog.Logger, f *file.File, ref *file.AttributeReference) {
	imp := f.ImportByNamespace(ref.AST.Package.Name)
	if imp == nil || !imp.Explicit() {
		logger.Error("Could not find import for package")
		l.reportMissingImport(f, ref.AST.Package.Name, &diagnostic.Diagnostic{
			Message: "attribute: unresolved reference to package",
			Primary: []diagnostic.Annotation{
				anno.Node(f, ref.AST.Package, "missing import for this package"),
			},
		})
		return
	}
	if imp.LoadedWithErrors {
		logger.Warn("Could not resolve reference, but import was not loaded, not reporting error")
		return
	}

	matches := imp.Package.AttributeSpecByQualifiedName(ref.AST.Name.Name)
	switch {
	case len(matches) == 0:
		logger.Error("Could not resolve reference")
		l.report(&diagnostic.Diagnostic{
			Message: "attribute: unresolved reference",
			Primary: []diagnostic.Annotation{
				anno.Node(f, ref.AST, "attribute is not defined in package"),
			},
		})
	case len(matches) == 1:
		logger.Debug("Found attribute")
		ref.Spec = matches[0]
	default:
		logger.Error("Found multiple attribute definitions with same specificity",
			slog.Int("count", len(matches)),
			slog.Int("specificity", matches[0].Specificity))
		l.report(&diagnostic.Diagnostic{
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
