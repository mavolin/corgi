package link

import (
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
)

func (l *linker) LinkElementReferences() {
	logger := l.logger.WithGroup("links.element_references")
	logger.Debug("Linking element references")

	for _, f := range l.p.Files {
		logger := logger.With(slog.String("file", f.Name))

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

			if ref.AST.Package != nil {
				l.linkQualifiedElementReference(logger, f, ref)
			} else {
				l.linkUnqualifiedElementReference(logger, f, ref)
			}
		}
	}
}

func (l *linker) linkUnqualifiedElementReference(logger *slog.Logger, f *file.File, ref *file.ElementReference) {
	name := ref.AST.Name.Name

	// search in current package
	if def := f.Package.ElementSpecByFullName(name); def != nil {
		ref.Spec = def
		return
	}

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

		if spec := imp.Package.ElementSpecByFullName(name); spec != nil {
			ref.Spec = spec
			return
		}
	}

	if builtinImp := f.BuiltinImport(); builtinImp != nil {
		if spec := builtinImp.Package.ElementSpecByFullName(name); spec != nil {
			ref.Spec = spec
			return
		}
	}

	if ignoreError {
		logger.Warn("Could not resolve reference, but at least one dot import was not loaded: not reporting error")
		return
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

func (l *linker) linkQualifiedElementReference(logger *slog.Logger, f *file.File, ref *file.ElementReference) {
	imp := f.ImportByNamespace(ref.AST.Package.Name)
	if imp == nil || !imp.Explicit() {
		logger.Error("Could not find import for package")
		l.reportMissingImport(f, ref.AST.Package.Name, &diagnostic.Diagnostic{
			Message: "element: unresolved reference to package",
			Primary: []diagnostic.Annotation{
				anno.Node(f, ref.AST.Package, "missing import for this package"),
			},
		})
		return
	} else if imp.LoadedWithErrors {
		logger.Debug("Could not resolve reference, but import was not loaded, not reporting error")
		return
	}

	if imp.Package != nil {
		ref.Spec = imp.Package.ElementSpecByQualifiedName(ref.AST.Name.Name)
	}
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
