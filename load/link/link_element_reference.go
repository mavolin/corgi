package link

import (
	"context"
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
)

func (l *linker) LinkElementReferences(ctx context.Context) {
	logger := l.logger.WithGroup("element_references")
	logger.Info("Linking element references")

	for _, f := range l.p.Files {
		logger := logger.With(slog.String("file", f.Name))
		logger.Debug("Linking file")

		for _, ref := range f.ElementReferences {
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
			logger.Debug("Linking element reference")

			if ref.AST.Package != nil {
				l.linkQualifiedElementReference(ctx, logger, f, ref)
			} else {
				l.linkUnqualifiedElementReference(ctx, logger, f, ref)
			}
		}
	}
}

func (l *linker) linkUnqualifiedElementReference(_ context.Context, logger *slog.Logger, f *file.File, ref *file.ElementReference) {
	logger.Debug("Unqualified element reference: local, builtin or dot import element definition")

	name := ref.AST.Name.Name
	if def := f.Package.ElementSpecByFullName(name); def != nil {
		ref.Spec = def
		logger.Debug("Found element definition within package")
		return
	}

	var ignoreError bool
	for _, imp := range f.Imports {
		if imp.Namespace() != "." {
			continue
		} else if imp.Package == nil {
			ignoreError = true
			continue
		}

		if def := imp.Package.ElementSpecByFullName(name); def != nil {
			logger.Debug("Found component within dot import",
				slog.String("import", imp.Package.PathInModule))
			return
		}
	}
	if ignoreError {
		logger.Warn("Could not resolve reference, but at least one dot import was not loaded: not reporting error (not searching in builtin because of this)")
		return
	}

	if l.builtin != nil {
		if def := l.builtin.ElementSpecByFullName(name); def != nil {
			ref.Spec = def
			logger.Debug("Found element definition within builtin package")
			return
		}
	}

	logger.Error("Could not resolve reference")
	l.report(&diagnostic.Diagnostic{
		Message: "element: unresolved reference",
		Primary: []diagnostic.Annotation{
			anno.Node(f, ref.AST, "could not resolve reference"),
		},
		Explanation: "As part of corgi's security model, all elements must be defined beforehand. " +
			"The linker could not find a definition for the element you tried to reference. " +
			"If you're using a HTML5 element, you probably have a typo. " +
			"If you're using a custom element, define it beforehand.",
		Docs: "element-definition",
	})
}

func (l *linker) linkQualifiedElementReference(_ context.Context, logger *slog.Logger, f *file.File, ref *file.ElementReference) {
	slog.Debug("Qualified element reference: external element definition")

	imp := f.ImportByNamespace(ref.AST.Package.Name)
	if imp == nil {
		logger.Error("Could not find import for package")

		if l.reportedMissingImports[f].Contains(ref.AST.Package.Name) {
			l.reportedMissingImports[f].Add(ref.AST.Package.Name)
			l.report(&diagnostic.Diagnostic{
				Message: "element: unresolved reference to package",
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

	ref.Spec = imp.Package.ElementSpecByQualifiedName(ref.AST.Name.Name)
	if ref.Spec == nil {
		logger.Error("Could not resolve reference")
		l.report(&diagnostic.Diagnostic{
			Message: "element: unresolved reference",
			Primary: []diagnostic.Annotation{
				anno.Node(f, ref.AST, "element is not defined in package"),
			},
		})
		return
	}
}
