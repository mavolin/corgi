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

	// search in current package
	if c := l.p.ComponentByName(ident.Name); c != nil {
		cc.Component = c
		return
	}

	// if this is unexported: check if this is a builtin component
	if !file.IsExported(ident.Name) {
		builtinImp := f.BuiltinImport()
		if builtinImp != nil && builtinImp.Package != nil && builtinImp.Package.PackageSymbols != nil {
			if c := builtinImp.Package.ComponentByName(ident.Name); c != nil {
				builtinImp.Forward = true
				cc.Component = c
				return
			}
		} else if l.builtinPath != "" {
			// The builtin import was not loaded, but should've been.
			logger.Debug("Couldn't resolve reference, but builtin import was loaded with errors: not reporting error")
			return
		}
	} else {
		// exported: check dot imports
		var ignoreError bool
		for _, imp := range f.Imports {
			if imp.LoadedWithErrors() || imp.Package.PackageSymbols == nil {
				ignoreError = true
			}
			switch {
			case !imp.Explicit() || imp.Namespace != "":
				continue
			case imp.Package == nil || imp.Package.PackageSymbols == nil:
				continue
			}

			cc.Component = imp.Package.ComponentByName(ident.Name)
			if cc.Component != nil {
				imp.Forward = true
				return
			}
		}

		if ignoreError {
			logger.Debug("Couldn't resolve reference, but at least one dot import was not loaded: not reporting error")
			return
		}
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
		return
	case ident.Name == nil:
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
	if imp == nil {
		logger.Error("Could not find import for package")
		l.reportMissingImport(f, ident.Package.Name, &diagnostic.Diagnostic{
			Message: "component call: unresolved reference to package",
			Primary: []diagnostic.Annotation{
				anno.Node(f, ident.Package, "missing import for package"),
			},
		})
		return
	}

	if !l.implicitImportCheck(logger, f, imp, ident.Package, "a", "component") {
		return
	}

	if imp.Package != nil && imp.Package.PackageSymbols != nil {
		cc.Component = imp.Package.ComponentByName(ident.Name.Name)
		if cc.Component != nil {
			imp.Forward = true
			return
		}
	}

	if imp.LoadedWithErrors() || imp.Package.PackageSymbols == nil {
		logger.Debug("Couldn't resolve reference, but package was loaded with errors: not reporting error")
		return
	}

	logger.Error("Could not resolve reference")
	l.report(&diagnostic.Diagnostic{
		Message: "component call: unresolved reference",
		Primary: []diagnostic.Annotation{
			anno.Node(f, ident.Name, "could not resolve reference"),
		},
		Secondary: []diagnostic.Annotation{
			anno.Node(f, imp.AST, "when searching in this package"),
		},
		Explanation: "The component you are trying to call does not exist in the package.",
	})
}
