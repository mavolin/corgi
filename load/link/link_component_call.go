package link

import (
	"context"
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
)

func (l *linker) LinkComponentCalls(ctx context.Context) {
	logger := l.logger.WithGroup("component_calls")
	logger.Info("Linking component calls")
	if l.builtin == nil {
		logger.Warn("No builtin package set")
	}

	for _, f := range l.p.Files {
		logger := logger.With(slog.String("file", f.Name))
		logger.Debug("Linking file")

		for _, cc := range f.ComponentCalls {
			if cc == nil || cc.AST.Header == nil || cc.AST.Header.Name == nil {
				continue
			}
			logger := logger.With(
				slog.String("pos", cc.AST.Start().String()),
				slog.String("component", cc.AST.Header.Name.Full()))
			logger.Debug("Linking component call")

			switch ident := cc.AST.Header.Name.(type) {
			case *ast.Ident:
				if ident == nil {
					continue
				}
				l.linkUnqualifiedComponentCall(ctx, logger, f, cc)
			case *ast.QualifiedIdent:
				if ident == nil {
					continue
				}
				l.linkQualifiedComponentCall(ctx, logger, f, cc)
			}
		}
	}
}

func (l *linker) linkUnqualifiedComponentCall(_ context.Context, logger *slog.Logger, f *file.File, cc *file.ComponentCall) {
	logger.Debug("Unqualified call: local, builtin, or dot import component")

	ident := cc.AST.Header.Name.(*ast.Ident)
	if c := l.p.ComponentByName(ident.Ident); c != nil {
		logger.Debug("Found component within package")
		cc.Component = c
	}

	if !file.IsExported(ident.Ident) {
		return
	}

	var ignoreError bool
	for _, imp := range f.Symbols.Imports {
		if imp.Namespace() != "." {
			continue
		} else if imp.Package == nil {
			ignoreError = true
			continue
		}

		if c := imp.Package.ComponentByName(ident.Ident); c != nil {
			logger.Debug("Found component within dot import",
				slog.String("import", imp.ImportPath()))
			cc.Component = c
			return
		}
	}
	if ignoreError {
		logger.Warn("Could not resolve reference, but least one dot import was not loaded: not reporting error (not searching in builtin because of this)")
		return
	}

	if l.builtin != nil {
		if c := l.builtin.ComponentByName(ident.Ident); c != nil {
			logger.Debug("Found component within builtin package")
			cc.Component = c
		}
	}

	logger.Error("Could not resolve reference")
	l.report(&diagnostic.Diagnostic{
		Message: "component call: unresolved reference",
		Primary: []diagnostic.Annotation{
			anno.Node(f, cc.AST.Header.Name, "could not resolve reference"),
		},
	})
}

func (l *linker) linkQualifiedComponentCall(_ context.Context, logger *slog.Logger, f *file.File, cc *file.ComponentCall) {
	logger.Debug("Qualified call: external component")

	ident := cc.AST.Header.Name.(*ast.QualifiedIdent)
	if ident.Package == nil {
		logger.Warn("Qualified call with nil package, skipping")
		return
	} else if ident.Name == nil {
		logger.Warn("Qualified call with nil name, skipping")
		return
	} else if !file.IsExported(ident.Name.Ident) {
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

	imp := f.ImportByNamespace(ident.Package.Ident)
	if imp == nil {
		logger.Error("Could not find import for package")

		if l.reportedMissingImports[f].Add(ident.Package.Ident) {
			l.report(&diagnostic.Diagnostic{
				Message: "component call: unresolved reference to package",
				Primary: []diagnostic.Annotation{
					anno.Node(f, ident.Package, "missing import for this package"),
				},
			})
		}
		return
	}

	if imp.Package == nil {
		logger.Warn("Could not resolve reference, but import was not loaded, not reporting error")
		return
	}

	cc.Component = imp.Package.ComponentByName(ident.Name.Ident)
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
