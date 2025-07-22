package link

import (
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
)

func (l *linker) LinkComponentCalls() {
	logger := l.logger.WithGroup("link.component_calls")
	logger.Debug("Linking component calls")

	for _, f := range l.p.Files {
		logger := logger.With(slog.String("file", f.Name))

		for _, cc := range f.ComponentCalls {
			if cc == nil || cc.AST.Header == nil || cc.AST.Header.Name == nil {
				continue
			}

			logger := logger.With(
				slog.String("pos", cc.AST.Start().String()),
				slog.String("component", cc.AST.Header.Name.Full()))

			switch ident := cc.AST.Header.Name.(type) {
			case *ast.Identifier:
				l.linkUnqualifiedComponentCall(logger, f, cc, ident)
			case *ast.QualifiedIdentifier:
				l.linkQualifiedComponentCall(logger, f, cc, ident)
			}
		}
	}
}

func (l *linker) linkUnqualifiedComponentCall(logger *slog.Logger, f *file.File, cc *file.ComponentCall, ident *ast.Identifier) {
	if ident == nil {
		return
	}

	logger.Debug("Unqualified call: local, builtin, or dot import component")

	// search in current package
	if c := l.p.ComponentByName(ident.Name); c != nil {
		logger.Debug("Found component within package")
		cc.Component = c
		return
	}

	// if this is unexported: check if this is a builtin component
	if !file.IsExported(ident.Name) {
		if builtinImp := f.BuiltinImport(); builtinImp != nil {
			if c := builtinImp.Package.ComponentByName(ident.Name); c != nil {
				logger.Debug("Found component within builtin package")
				cc.Component = c
				return
			}
		}
		return
	}

	// exported: check dot imports
	var ignoreError bool
	for _, imp := range f.Imports {
		switch {
		case !imp.Explicit() || imp.Namespace != ".":
			continue
		case imp.LoadedWithErrors:
			ignoreError = true
		case imp.Package == nil:
			continue
		}

		if c := imp.Package.ComponentByName(ident.Name); c != nil {
			logger.Debug("Found component within dot import",
				slog.String("import", imp.Path))
			cc.Component = c
			return
		}
	}

	if ignoreError {
		logger.Warn("Could not resolve reference, but least one dot import was not loaded: not reporting error")
		return
	}

	logger.Error("Could not resolve reference")
	l.report(&diagnostic.Diagnostic{
		Message: "component call: unresolved reference",
		Primary: []diagnostic.Annotation{
			anno.Node(f, cc.AST.Header.Name, "no component with that name found in the current package or dot imports"),
		},
	})
}

func (l *linker) linkQualifiedComponentCall(
	logger *slog.Logger, f *file.File, cc *file.ComponentCall, ident *ast.QualifiedIdentifier,
) {
	if ident == nil {
		return
	}

	switch {
	case ident.Package == nil:
		logger.Warn("Qualified call with nil package, skipping")
		return
	case ident.Name == nil:
		logger.Warn("Qualified call with nil name, skipping")
		return
	case !file.IsExported(ident.Name.Name):
		logger.Error("Qualified call to unexported component")
		l.report(&diagnostic.Diagnostic{
			Message: "component call: cannot call unexported component",
			Primary: []diagnostic.Annotation{
				anno.Node(f, ident.Name, "the component you are trying to call is unexported"),
			},
			Explanation: "You can only call components from other packages if they are exported.\n" +
				"If you control the source of the package, " +
				"you can export the component by changing its name to start with an uppercase letter.",
		})
		return
	}

	// find import for package
	imp := f.ImportByNamespace(ident.Package.Name)
	if imp == nil || !imp.Explicit() {
		logger.Error("Could not find import for package")
		l.reportMissingImport(f, ident.Package.Name, &diagnostic.Diagnostic{
			Message: "component call: unresolved reference to package",
			Primary: []diagnostic.Annotation{
				anno.Node(f, ident.Package, "missing import for package"),
			},
		})
		return
	}

	if imp.LoadedWithErrors {
		logger.Warn("Could not resolve reference, but import was loaded with error, not reporting subsequent error")
		return
	}

	if imp.Package != nil {
		cc.Component = imp.Package.ComponentByName(ident.Name.Name)
	}
	if cc.Component == nil {
		logger.Error("Could not resolve reference")
		l.report(&diagnostic.Diagnostic{
			Message: "component call: unresolved reference",
			Primary: []diagnostic.Annotation{
				anno.Node(f, ident.Name, "could not resolve reference"),
			},
			Secondary: []diagnostic.Annotation{
				anno.Node(f, imp.AST, "in this package"),
			},
			Explanation: "The component you are trying to call does not exist in the package.",
		})
		return
	}
}
