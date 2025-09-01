package analyze

import (
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
)

// AnalyzeBlocks analyzes all blocks of the given component.
//
// Depends on Checks: None
//
// Sets Fields: None
//
// Depends on Fields: None
func (z *analyzer) AnalyzeBlocks(logger *slog.Logger, c *file.Component) {
	logger = logger.WithGroup("blocks")

	for _, block := range c.Blocks {
		logger := logger.With(slog.String("block", block.Name))

		z.AnalyzeBlockRequired(logger, c, block)
		z.AnalyzeBlockForwarded(block)
		z.AnalyzeBlockCannotForwardAttributes(block)
	}
}

// ============================================================================
// Required
// ======================================================================================

// AnalyzeBlockRequired determines whether the given component block is
// required to be set by component calls.
//
// Depends on Checks: None
//
// Sets Fields:
//   - Components.Blocks.Required
//
// Depends on Fields: None
func (z *analyzer) AnalyzeBlockRequired(logger *slog.Logger, c *file.Component, b *file.Block) {
	logger = logger.WithGroup("required")

	b.Required.SetResult(b.Instances[0].AST.Default == nil)
	for _, instance := range b.Instances[1:] {
		if b.Required.Result() && instance.AST.Default != nil {
			goto Erroneous
		} else if !b.Required.Result() && instance.AST.Default == nil {
			goto Erroneous
		}
	}

	return
Erroneous:
	b.Required.SetFailed()
	primaries := make([]diagnostic.Annotation, len(b.Instances))
	for i, instance := range b.Instances {
		if instance.AST.Default == nil {
			primaries[i] = anno.Node(c.File, instance.AST, "no default -> required")
		} else {
			primaries[i] = anno.Node(c.File, instance.AST, "default set -> not required")
		}
	}

	logger.Error("Block has conflicting defaults, cannot determine if it is required")
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

// ============================================================================
// Forwarded
// ======================================================================================

// AnalyzeBlockForwarded determines whether the given component block is
// top-level, i.e. it contains at least one instance that is top-level.
//
// Depends on Checks: None
//
// Sets Fields:
//   - Components.Blocks.Forwarded
//
// Depends on Fields:
//   - Components.Blocks.Instances.Forwarded
func (z *analyzer) AnalyzeBlockForwarded(b *file.Block) {
	b.Forwarded.SetFalse()
	for _, instance := range b.Instances {
		if instance.Forwarded.Equal(true) {
			b.Forwarded.SetReason(instance)
		} else if instance.Forwarded.Failed() {
			b.Forwarded.SetFailed()
		}
	}
}

// ============================================================================
// Cannot Forward Attributes
// ======================================================================================

// AnalyzeBlockCannotForwardAttributes determines whether the given component
// block forwards attributes to the element containing it.
//
// Depends on Checks: None
//
// Sets Fields:
//   - Components.Blocks.CannotForwardAttributes
//
// Depends on Fields:
//   - Components.Blocks.Instances.CannotForwardAttributes
func (z *analyzer) AnalyzeBlockCannotForwardAttributes(b *file.Block) {
	b.CannotForwardAttributes.SetFalse()
	for _, instance := range b.Instances {
		if instance.CannotForwardAttributes.Failed() {
			b.CannotForwardAttributes.SetFailed()
			return
		} else if instance.CannotForwardAttributes.True() {
			b.CannotForwardAttributes.SetReason(instance)
			return
		}
	}
}
