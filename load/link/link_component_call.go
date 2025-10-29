package link

import (
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
)

type componentCallsLinked struct{}

func (l *linker) LinkComponentCalls() {
	defer l.Ran(l.Pkg, componentCallsLinked{})
	l.Require(l.Pkg, importsLoaded{})
	l.Require(l.Pkg, dotImportComponentCollisionCheck{})
	l.Require(l.Pkg, componentCollisionCheck{})

	logger := l.Logger.WithGroup("link.component_calls")
	logger.Debug("Linking component calls")

	for _, f := range l.Pkg.Files {
		logger := logger.With(slog.String("file", string(f.Name)))

		for _, cc := range f.ComponentCalls {
			cc.Linked = true
			if cc.AST.Header == nil || cc.AST.Header.Name == nil {
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
			if cc.Component != nil {
				l.linkBlockSetterBlocks(logger, cc)
				l.linkComponentArguments(logger, cc)
			}
		}
	}
}

func (l *linker) linkUnqualifiedComponentCall(logger *slog.Logger, f *file.File, cc *file.ComponentCall, ident *ast.Identifier) {
	if ident == nil {
		return
	}

	// search in current package
	if c := l.Pkg.ComponentByName(cc.Name); c != nil {
		cc.Component = c
		return
	}

	if cc.Name.Exported() { // check dot imports
		var ignoreError bool
		for _, imp := range f.Imports {
			if imp.Package == nil || imp.Package.PackageSymbols == nil {
				ignoreError = true
			}
			switch {
			case !imp.Explicit() || imp.Qualifier != "":
				continue
			case imp.Package == nil || imp.Package.PackageSymbols == nil:
				continue
			}

			cc.Component = imp.Package.ComponentByName(cc.Name)
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
	l.Report(&diagnostic.Diagnostic{
		Message: "component call: unresolved reference",
		Primary: []diagnostic.Annotation{
			anno.Node(f, cc.AST.Header.Name, "neither defined in the current package nor dot imports"),
		},
	})
}

func (l *linker) linkQualifiedComponentCall(logger *slog.Logger, f *file.File, cc *file.ComponentCall, ident *ast.QualifiedIdentifier) {
	switch {
	case cc.Qualifier == "":
		return
	case cc.Name == "":
		return
	case !cc.Name.Exported():
		logger.Error("Qualified call to unexported component")
		l.Report(&diagnostic.Diagnostic{
			Message: "component call: cannot call unexported component",
			Primary: []diagnostic.Annotation{
				anno.Node(f, cc.AST.Header.Name, "the component you are trying to call is unexported"),
			},
			Explanation: "You can only call components from other packages if they are exported.\n" +
				"If you control the source of the package, " +
				"you can export the component by changing its name to start with an uppercase letter.",
		})
		return
	}

	// find import for package
	imp := f.ImportByQualifier(cc.Qualifier)
	if imp == nil {
		logger.Error("Could not find import for package")
		l.ReportMissingImport(f, cc.Qualifier, &diagnostic.Diagnostic{
			Message: "component call: unresolved reference to package",
			Primary: []diagnostic.Annotation{
				anno.Node(f, cc.AST.Header.Name, "missing import for package"),
			},
		})
		return
	}

	if !l.implicitImportCheck(logger, f, imp, ident.Package, "a", "component") {
		return
	}

	if imp.Package != nil && imp.Package.PackageSymbols != nil {
		cc.Component = imp.Package.ComponentByName(cc.Name)
		if cc.Component != nil {
			imp.Forward = true
			return
		}
	}

	if imp.Package == nil || imp.Package.PackageSymbols == nil {
		logger.Debug("Couldn't resolve reference, but package was loaded with errors: not reporting error")
		return
	}

	logger.Error("Could not resolve reference")
	l.Report(&diagnostic.Diagnostic{
		Message: "component call: unresolved reference",
		Primary: []diagnostic.Annotation{
			anno.Node(f, ident.Name, "could not resolve reference"),
		},
		Secondary: []diagnostic.Annotation{
			anno.Node(f, imp.AST, "searching in this package"),
		},
		Explanation: "The component you are trying to call does not exist in the package.",
	})
}

// linkBlockSetterBlocks links the Block field of the BlockSetters of the given
// component call.
func (l *linker) linkBlockSetterBlocks(logger *slog.Logger, cc *file.ComponentCall) {
	logger = logger.WithGroup("block_setter_blocks")

	for _, blockSetter := range cc.BlockSetters {
		logger := logger.With(slog.String("with_name", string(blockSetter.Name)))

		blockSetter.Block = cc.Component.BlockByName(blockSetter.Name)
		if blockSetter.Block != nil {
			continue
		}

		logger.Error("Block setter block not found")

		primaries := make([]diagnostic.Annotation, len(blockSetter.Instances))
		for i, instance := range blockSetter.Instances {
			var annotation string
			if blockSetter.Name == "" {
				annotation = "`" + cc.AST.Header.Name.Full() + "` defines no default block"
			} else {
				annotation = "`" + cc.AST.Header.Name.Full() + "` defines no block with this name"
			}

			primaries[i] = anno.Node(cc.File, instance.AST, annotation)
		}

		l.Report(&diagnostic.Diagnostic{
			Message: "component call: block setter references unknown block",
			Primary: primaries,
			Secondary: []diagnostic.Annotation{
				anno.Node(cc.File, cc.AST, "in this component call"),
			},
		})
	}
}

func (l *linker) linkComponentArguments(logger *slog.Logger, cc *file.ComponentCall) {
	logger = logger.WithGroup("component_arguments")

	for _, arg := range cc.ComponentArguments {
		logger := logger.With(slog.String("name", string(arg.Name)))

		arg.Parameter = cc.Component.ParameterByName(arg.Name)
		if arg.Parameter != nil {
			continue
		}

		logger.Error("Could not find parameter for argument")
		l.Report(&diagnostic.Diagnostic{
			Message: "component call: argument: unresolved reference",
			Primary: []diagnostic.Annotation{
				anno.Node(cc.File, arg.AST, "`"+string(cc.Name)+"` defines no parameter `"+string(arg.Name)+"`"),
			},
		})
	}
}
