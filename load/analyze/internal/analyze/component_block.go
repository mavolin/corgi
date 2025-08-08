package analyze

import (
	"context"
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	"github.com/mavolin/corgi/v2/file/walk"
)

// AnalyzeBlocks runs trivial analyses on the blocks of the given
// component.
//
// An analysis is trivial, if it does not depend on the analysis of a component
// call or the analysis of another component.
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
		// todo: instance forwards attributes
		z.AnalyzeBlockForwarded(block)
		z.AnalyzeBlockForwardsAttributes(block)
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
// Depends on Fields: None
func (z *analyzer) AnalyzeBlockRequired(logger *slog.Logger, c *file.Component, b *file.Block) {
	logger = logger.WithGroup("required")

	b.Required.Set(b.Instances[0].AST.Default == nil)
	for _, instance := range b.Instances[1:] {
		if b.Required.Result && instance.AST.Default != nil {
			goto Erroneous
		} else if !b.Required.Result && instance.AST.Default == nil {
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

// AnalyzeBlockForwarded determines whether the given component block is top-level,
// i.e. it contains at least one instance that is top-level.
//
// Depends on Checks: None
//
// Sets Fields:
//   - Components.Blocks.Forwarded
//
// Depends on Fields:
//   - Components.Blocks.Instances.Forwarded
func (z *analyzer) AnalyzeBlockForwarded(b *file.Block) {
	var failed bool
	for _, instance := range b.Instances {
		if instance.Forwarded.Equal(true) {
			b.Forwarded.Set(true)
		}
		failed = failed || instance.Forwarded.Failed
	}

	b.Forwarded.SetIf(false, !failed)
}

// AnalyzeBlockInstanceForwarded determines whether the given block is
// forwarded.
//
// Depends on Checks: None
//
// Sets Fields:
//   - Components.Blocks.Instances.Forwarded
//
// Depends on Fields: None
func (z *analyzer) AnalyzeBlockInstanceForwarded(ctx context.Context, c *file.Component, parents []*walk.Context, biAST *ast.Block) {
	bi := c.BlockInstanceByNode(biAST)
	if bi == nil {
		return
	}

	bi.Forwarded = z.isForwarded(ctx, c.File, parents)
}

// ============================================================================
// Forwards Attributes
// ======================================================================================

// AnalyzeBlockForwardsAttributes determines whether the given component block
// forwards attributes to the element containing it.
//
// Depends on Checks: None
//
// Sets Fields:
//   - Components.Blocks.ForwardsAttributes
//
// Depends on Fields:
//   - Components.Blocks.Instances.ForwardsAttributes
func (z *analyzer) AnalyzeBlockForwardsAttributes(b *file.Block) {
	for _, instance := range b.Instances {
		if instance.ForwardsAttributes.Failed {
			b.ForwardsAttributes.SetFailed()
			return
		} else if instance.ForwardsAttributes.Equal(false) {
			b.ForwardsAttributes.Set(false)
			return
		}
	}
	b.ForwardsAttributes.Set(true)
}
