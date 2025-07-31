package analyze

import (
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/walk"
)

// TrivialAnalyzeComponentCalls runs all trivial analyses on the components
// calls.
//
// An analysis is trivial, if it does not depend on the analysis of a component
// or the analysis of another component call.
//
// Depends on Checks: None
//
// Sets Fields: None
//
// Depends on Fields: None
func (z *analyzer) TrivialAnalyzeComponentCalls() {
	logger := z.Logger.WithGroup("component_calls")
	logger.Debug("Analyzing component calls")

	for _, f := range z.P.Files {
		logger := logger.With(slog.String("file", f.Name))

		for _, cc := range f.ComponentCalls {
			logger := logger.With(
				slog.String("call_package", cc.Component.File.Package.Module+"/"+cc.Component.File.Package.PathInModule),
				slog.String("call_name", cc.Component.AST.Header.Name.Name),
				slog.String("call_pos", cc.AST.Start().String()))

			z.ComponentCallFindFirstAnd(logger, cc)
		}
	}
}

// ============================================================================
// Find First And
// ======================================================================================

// ComponentCallFindFirstAnd finds the first & in the body of the given
// component call.
//
// Depends on Checks: None
//
// Sets Fields:
//   - ComponentCalls.FirstAnd
//
// Depends on Fields: None
func (z *analyzer) ComponentCallFindFirstAnd(logger *slog.Logger, cc *file.ComponentCall) {
	logger = logger.WithGroup("find_first_and")

	if cc.AST.Body == nil {
		return
	}
	scope, _ := cc.AST.Body.(*ast.Scope)
	if scope == nil {
		return
	}

	walk.WalkT(scope, func(ctx *walk.ContextT[*ast.And]) error {
		cc.FirstAnd = ctx.Node
		return walk.Stop
	}, walk.DontDive[ast.BlockSetter]())
}
