package analyze

import (
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
)

// TrivialAnalyzeBlocks runs trivial analyses on the blocks of the given
// component.
//
// An analysis is trivial, if it does not depend on the analysis of a component
// call or the analysis of another component.
//
// Depends on Checks: None
//
// Sets Fields: None
//
// Depends on Fields:
//   - Components.Blocks
func (z *analyzer) TrivialAnalyzeBlocks(logger *slog.Logger, c *file.Component) {
	logger = logger.WithGroup("blocks")
	logger.Debug("Analyzing blocks")

	for _, block := range c.Blocks {
		if block.Component != c {
			continue // Only analyze blocks that belong to the component
		}

		logger := logger.With(slog.String("block", block.Name))
		logger.Debug("Analyzing block")

		z.AnalyzeBlockRequired(logger, block)
	}
}

// AnalyzeBlockRequired determines whether the given component block is
// required to be set by component calls.
//
// Depends on Checks: None
//
// Sets Fields:
//   - Components.Blocks.Required
//
// Depends on Fields:
//   - Components.Blocks.Instances.AST
func (z *analyzer) AnalyzeBlockRequired(logger *slog.Logger, b *file.Block) {
	logger = logger.WithGroup("required")
	logger.Debug("Determining if block is required")

	b.Required = b.Instances[0].AST.Default == nil
	for _, instance := range b.Instances[1:] {
		if b.Required && instance.AST.Default != nil {
			goto Erroneous
		}
	}

	return
Erroneous:
	primaries := make([]diagnostic.Annotation, len(b.Instances))
	for i, instance := range b.Instances {
		if instance.AST.Default == nil {
			primaries[i] = anno.Range(b.Component.File, instance.AST.Start(), instance.AST.Identifier.End(), "no default -> required") // todo
		} else {
			primaries[i] = anno.Range(b.Component.File, instance.AST.Start(), instance.AST.Identifier.End(), "default set -> not required") // todo
		}
	}

	z.Report(&diagnostic.Diagnostic{
		Message: "block: cannot determine if block is required: conflicting defaults",
		Primary: primaries,
		Explanation: "For every instance of a block with the same name, all instances must either " +
			"have a default value, or none of them must have a default value. " +
			"If all instances have a default value, the block is considered not required, " +
			"otherwise it is considered required.",
		Hints: []diagnostic.Hint{
			{
				Hint:    "To insert nothing, if the block is not set, use `{}` as default",
				Example: "`block " + b.Name + " {}`",
			},
		},
	})
}
