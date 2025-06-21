package link

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	"github.com/mavolin/corgi/v2/internal/set"
)

const ambiguousAttributeReferenceExplanation = "There are multiple regular expression selectors that all match this attribute and " +
	"therefore it is unclear which one to use. " +
	"Refine your regular expressions so that only one matches to resolve this ambiguity."

func (l *linker) LinkAttributeReferences(ctx context.Context) {
	logger := l.logger.WithGroup("attribute_references")
	logger.Info("Linking attribute references")

	for _, f := range l.p.Files {
		logger := logger.With(slog.String("file", f.Name))
		logger.Debug("Linking file")

		for _, ref := range f.AttributeReferences {
			if ref.AST.Name == nil {
				continue
			}

			var name string
			if ref.AST.Package != nil {
				name = ref.AST.Package.Ident + "." + ref.AST.Name.Name
			} else {
				name = ref.AST.Name.Name
			}

			logger := logger.With(
				slog.String("name", name),
				slog.String("pos", ref.AST.Start().String()))
			logger.Debug("Linking attribute reference")

			if ref.AST.Package != nil {
				l.linkQualifiedAttributeReference(ctx, logger, f, ref)
			} else {
				l.linkUnqualifiedAttributeReference(ctx, logger, f, ref)
			}
		}
	}
}

func (l *linker) linkUnqualifiedAttributeReference(_ context.Context, logger *slog.Logger, f *file.File, ref *file.AttributeReference) {
	logger.Debug("Unqualified attribute reference: local, builtin or dot import attribute definition")

	name := ref.AST.Name.Name

	var (
		bestMatchImport         string
		equalSpecificityMatches []*file.AttributeDefinition
	)

	packageMatches := f.Package.AttributeDefinitionByFullName(name)
	equalSpecificityMatches = packageMatches

	for _, imp := range f.Symbols.Imports {
		if imp.Namespace() != "." || imp.Package == nil {
			continue
		}

		packageMatches = imp.Package.AttributeDefinitionByFullName(name)
		if len(equalSpecificityMatches) == 0 || equalSpecificityMatches[0].Specificity < packageMatches[0].Specificity {
			equalSpecificityMatches = packageMatches
			bestMatchImport = imp.ImportPath()
		} else if equalSpecificityMatches[0].Specificity == packageMatches[0].Specificity {
			equalSpecificityMatches = append(equalSpecificityMatches, packageMatches...)
		}
	}

	if len(equalSpecificityMatches) == 1 {
		ref.Definition = equalSpecificityMatches[0]
		if bestMatchImport == "" {
			logger.Debug("Found attribute definition within package")
		} else {
			logger.Debug("Found attribute definition within package or dot import",
				slog.String("import", bestMatchImport))
		}
		return
	} else if len(equalSpecificityMatches) > 1 {
		logger.Error("Found multiple attribute definitions with same specificity",
			slog.Int("count", len(equalSpecificityMatches)),
			slog.Int("specificity", equalSpecificityMatches[0].Specificity))

		l.report(&diagnostic.Diagnostic{
			Message: "attribute: ambiguous reference",
			Primary: []diagnostic.Annotation{
				anno.Node(f, ref.AST,
					"there are multiple attribute selectors with the same specificity that match this attribute"),
			},
			Secondary:   equalSpecificityAnnotations(equalSpecificityMatches),
			Explanation: ambiguousAttributeReferenceExplanation,
			Docs:        "attribute-specificity",
		})
		return
	}

	if l.builtin != nil {
		packageMatches = l.builtin.AttributeDefinitionByFullName(name)
		if len(packageMatches) == 1 {
			logger.Debug("Found attribute definition within builtin package")
			ref.Definition = packageMatches[0]
		} else if len(packageMatches) > 1 {
			logger.Error("Found multiple attribute definitions with same specificity in builtin package")
			l.report(&diagnostic.Diagnostic{
				Message: "attribute: ambiguous reference",
				Primary: []diagnostic.Annotation{
					anno.Node(f, ref.AST,
						"there are multiple attribute selectors with the same specificity that match this attribute"),
				},
				Secondary:   equalSpecificityAnnotations(packageMatches),
				Explanation: ambiguousAttributeReferenceExplanation,
				Hints: []diagnostic.Hint{
					{
						Hint: "If you are using the corgi stdlib builtin package, you shouldn't see this error. " +
							"Please open an issue.",
					},
				},
				Docs: "attribute-specificity",
			})
		}
		return
	}

	logger.Debug("Not defined explicitly, analyzer needs to determine whether there is explicit typing")
}

func (l *linker) linkQualifiedAttributeReference(_ context.Context, logger *slog.Logger, f *file.File, ref *file.AttributeReference) {
	logger.Debug("Qualified attribute reference: external attribute definition")

	imp := f.ImportByNamespace(ref.AST.Package.Ident)
	if imp == nil {
		logger.Error("Could not find import for package")

		if l.reportedMissingImports[f].Add(ref.AST.Package.Ident) {
			l.report(&diagnostic.Diagnostic{
				Message: "attribute: unresolved reference to package",
				Primary: []diagnostic.Annotation{
					anno.Node(f, ref.AST.Package, "missing import for this package"),
				},
			})
		}
		return
	}
	if imp.Package == nil {
		logger.Warn("Could not resolve reference, but import was not loaded, not reporting error")
		return
	}

	matches := imp.Package.AttributeDefinitionByQualifiedName(ref.AST.Name.Name)
	if len(matches) == 0 {
		logger.Error("Could not resolve reference")
		l.report(&diagnostic.Diagnostic{
			Message: "attribute: unresolved reference",
			Primary: []diagnostic.Annotation{
				anno.Node(f, ref.AST, "attribute is not defined in package"),
			},
		})
	} else if len(matches) == 1 {
		logger.Debug("Found attribute")
		ref.Definition = matches[0]
	} else {
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

func equalSpecificityAnnotations(defs []*file.AttributeDefinition) []diagnostic.Annotation {
	reportedPrefixes := set.NewSliceSet[*ast.AttributeDefinition](len(defs))
	as := make([]diagnostic.Annotation, 0, 2*len(defs))
	for _, def := range defs {
		if def.Definition.Prefix != nil && reportedPrefixes.Add(def.Definition) {
			as = append(as, anno.Node(def.File, def.Definition.Prefix, "with this prefix"))
		}
		as = append(as,
			anno.Node(def.File, def.AST.Selector, fmt.Sprint("this selector has a specificity of ", def.Specificity)))
	}
	return as
}
