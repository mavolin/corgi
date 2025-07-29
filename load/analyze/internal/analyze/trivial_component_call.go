package analyze

import (
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
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
				slog.String("call_name", cc.Component.AST.Header.Name.Name),
				slog.String("call_pos", cc.AST.Start().String()))

			z.LinkBlockSetterBlocks(logger, cc)
		}
	}
}

// ============================================================================
// Link Block Setter's Blocks Field
// ======================================================================================

// LinkBlockSetterBlocks links the Block field of the BlockSetters of the given
// component call.
//
// Depends on Checks: None
//
// Sets Fields:
//   - ComponentCalls.BlockSetters.Block
//
// Depends on Fields: None
func (z *analyzer) LinkBlockSetterBlocks(logger *slog.Logger, cc *file.ComponentCall) {
	logger = logger.WithGroup("link_block_setter_blocks")

	if !cc.Component.File.Package.Analyzed || cc.Component.AnalyzedWithErrors {
		cc.AnalyzedWithErrors = true
		return
	}

	for _, blockSetter := range cc.BlockSetters {
		logger := logger.With(slog.String("with_name", blockSetter.Name))

		blockSetter.Block = cc.Component.BlockByName(blockSetter.Name)
		if blockSetter.Block != nil {
			continue
		}

		cc.AnalyzedWithErrors = true
		logger.Error("Block Setter block not found")

		primaries := make([]diagnostic.Annotation, len(blockSetter.Instances))
		for i, instance := range blockSetter.Instances {
			var highlight anno.HighlightFunc
			switch instance := instance.AST.(type) {
			case *ast.With:
				if instance.Identifier == nil {
					highlight = anno.HighlightNRunes(*instance.With, len("with"))
				} else {
					highlight = anno.HighlightNode(instance.Identifier)
				}
			case *ast.DefaultBlockShorthand:
				highlight = anno.HighlightPosition(instance.Body.Start())
			}

			var annotation string
			if blockSetter.Name == "" {
				annotation = "`" + cc.AST.Header.Name.Full() + "` defines no default block"
			} else {
				annotation = "`" + cc.AST.Header.Name.Full() + "` defines no block with this name"
			}

			primaries[i] = anno.Anno(cc.File, anno.Annotation{
				Highlight:  highlight,
				Annotation: annotation,
			})
		}

		z.Report(&diagnostic.Diagnostic{
			Message: "component call: with references unknown block",
			Primary: primaries,
			Secondary: []diagnostic.Annotation{
				anno.Node(cc.File, cc.AST, "in this component call"),
			},
		})
	}
}
