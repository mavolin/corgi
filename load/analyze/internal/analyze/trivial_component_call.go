package analyze

import (
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
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
				slog.String("call_name", cc.Component.Header().Name.Name),
				slog.String("call_pos", cc.AST.Start().String()))

			z.LinkWithBlocks(logger, cc)
		}
	}
}

// ============================================================================
// Link With's Blocks Field
// ======================================================================================

// LinkWithBlocks links the Block field of the Withs of the given
// component call.
//
// Depends on Checks:
//   - AggregateWiths - To iterate over the Withs
//   - AggregateBlocks - To find the blocks by name
//
// Sets Fields:
//   - ComponentCalls.Withs.Block
//
// Depends on Fields:
//   - ComponentCalls.Withs.Name
//   - Components.Blocks.Name
func (z *analyzer) LinkWithBlocks(logger *slog.Logger, cc *file.ComponentCall) {
	logger = logger.WithGroup("link_with_blocks")

	if !cc.Component.File.Package.Analyzed || cc.Component.AnalyzedWithErrors {
		cc.AnalyzedWithErrors = true
		return
	}

	for _, with := range cc.Withs {
		logger := logger.With(slog.String("with_name", with.Name))

		with.Block = cc.Component.BlockByName(with.Name)
		if with.Block == nil {
			cc.AnalyzedWithErrors = true
			logger.Error("With block not found")

			var primary diagnostic.Annotation
			if with.Name == "" {
				primary = anno.NRunes(cc.File, *with.Instances[0].AST.With, len("with"), "`"+cc.AST.Header.Name.Full()+"` defines no default block")
			} else {
				primary = anno.Node(cc.File, with.Instances[0].AST.Identifier, "`"+cc.AST.Header.Name.Full()+"` defines no block with this name")
			}

			z.Report(&diagnostic.Diagnostic{
				Message: "component call: with references unknown block",
				Primary: []diagnostic.Annotation{
					primary,
				},
				Secondary: []diagnostic.Annotation{
					anno.Node(cc.File, cc.AST, "in this component call"),
				},
			})
		}
	}
}
