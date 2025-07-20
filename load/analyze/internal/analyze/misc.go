package analyze

import (
	"log/slog"

	"github.com/mavolin/corgi/v2/escape/attrtype"
	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	"github.com/mavolin/corgi/v2/file/walk"
)

// ============================================================================
// Package Name
// ======================================================================================

// CheckPackageNamesMatch checks if all files in the package have the same
// package name.
//
// Depends on Checks: None
//
// Sets Fields: None
//
// Depends on Fields: None
func (z *analyzer) CheckPackageNamesMatch() (ok bool) {
	logger := z.Logger.WithGroup("check.package_names_match")
	logger.Info("Checking if package names match")

	if len(z.P.Files) <= 1 {
		logger.Debug("One or no files, skipping")
		return true
	}

	expect := z.P.Files[0].Name
	primaries := make([]diagnostic.Annotation, 1, len(z.P.Files))
	primaries[0] = anno.Node(z.P.Files[0], z.P.Files[0].AST.Package.Name, "found this name here")
	for _, f := range z.P.Files[1:] {
		if f.Name != expect {
			primaries = append(primaries, anno.Node(f, f.AST.Package.Name, "but found other name here"))
		}
	}

	if len(primaries) > 1 {
		logger.Error("Package names do not match")
		z.Report(&diagnostic.Diagnostic{
			Message: "package names do not match",
			Primary: primaries,
		})
	}

	return false
}

// SetPackageName sets the package name to the name of the first file in the
// package.
//
// Depends on Checks:
//   - CheckPackageNamesMatch - So that we don't assign an incorrect package
//     name.
//
// Sets Fields: None
//
// Depends on Fields: None
func (z *analyzer) SetPackageName() {
	logger := z.Logger.WithGroup("set_package_name")
	logger.Info("Setting package name")

	if len(z.P.Files) == 0 {
		logger.Warn("No files in package")
		return
	}

	z.P.Name = z.P.Files[0].Name
	logger.Debug("Set package name", "name", z.P.Name)
}

// ============================================================================
// Needs Escape Import
// ======================================================================================

// AnalyzeNeedsEscapeImport checks if the component's file needs an escape
// import base on the component's body and parameter defaults.
//
// Depends on Checks: None
//
// Sets Fields: None
//
// Depends on Fields:
//   - Components.Parameters.AttributeType
func (z *analyzer) AnalyzeNeedsEscapeImport() {
	logger := z.Logger.WithGroup("needs_escape_import")
	logger.Info("Checking if file needs escape import")

	for _, f := range z.P.Files {
		logger := logger.With(slog.String("file", f.Name))
		logger.Debug("Checking file")

		if f.NeedsEscapeImport {
			logger.Debug("File already marked as needing escape import, skipping")
			continue
		}

		for _, cc := range f.ComponentCalls {
			logger := logger.With(
				slog.String("call_name", cc.Component.Header().Name.Name),
				slog.String("call_pos", cc.AST.Start().String()))
			logger.Debug("Checking component call")

			z.analyzeComponentCallArgNeedsEscapeImport(logger, cc)
			if cc.File.NeedsEscapeImport {
				break
			}
		}
	}

	for _, c := range z.P.Components {
		logger := logger.With(
			slog.String("comp_name", c.Header().Name.Name),
			slog.String("comp_pos", c.Start().String()))
		logger.Debug("Checking component")

		if c.File.NeedsEscapeImport {
			logger.Debug("Component's file already marked as needing escape import, skipping")
			continue
		}

		z.analyzeNeedsEscapeImportThroughInterpolation(logger, c)
	}
}

func (z *analyzer) analyzeComponentCallArgNeedsEscapeImport(logger *slog.Logger, cc *file.ComponentCall) {
	logger = logger.WithGroup("component_call_arg")
	logger.Debug("Checking if one of the component call's arguments needs escape import")

	if cc.AST.Header.Arguments == nil || len(cc.AST.Header.Arguments.Args) == 0 {
		logger.Debug("Component call has no arguments, skipping")
		return
	}

	for _, arg := range cc.AST.Header.Arguments.Args {
		carg, _ := arg.(*ast.ComponentArgument)
		if carg == nil {
			continue
		}

		param := cc.Component.ParameterByName(carg.Name.Name)
		if param == nil {
			// this is reported by another analysis
			continue
		}

		if param.AttributeType == attrtype.Unknown || param.AttributeType == attrtype.Text {
			continue
		}

		cc.File.NeedsEscapeImport = true
		logger.Debug("Component call argument needs escape import",
			slog.String("arg", carg.Name.Name),
			slog.String("arg_pos", carg.Name.Start().String()))
		return
	}
}

func (z *analyzer) analyzeNeedsEscapeImportThroughInterpolation(logger *slog.Logger, c *file.Component) {
	logger = logger.WithGroup("expression_interpolation")
	logger.Debug("Checking if component needs escape import because of expression interpolation")

	var n ast.Node
	if c.DefinedAST != nil {
		n = c.DefinedAST
	} else if c.AliasAST != nil {
		n = c.AliasAST
	}

	walk.WalkT(n, func(ctx *walk.ContextT[*ast.ExpressionInterpolation]) error {
		logger.Debug("Found expression interpolation, marking component's file as needing escape import",
			slog.String("expr_pos", ctx.Node.Start().String()))
		c.File.NeedsEscapeImport = true
		return walk.Stop
	})
}
