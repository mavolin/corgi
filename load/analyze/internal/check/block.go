package check

import (
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
)

func (ch *checker) CheckBlocks(logger *slog.Logger, c *file.Component) {
	logger = logger.WithGroup("blocks")

	for _, block := range c.Blocks {
		logger := logger.With(slog.String("block", string(block.Name)))

		ch.CheckBlock_NoConflictingRequired(logger, block)

		ch.CheckBlockInstances(logger, block)
	}
}

// ============================================================================
// No Conflicting Required
// ======================================================================================

func (ch *checker) CheckBlock_NoConflictingRequired(logger *slog.Logger, b *file.Block) {
	if !b.Required().Failed() {
		return
	}

	primaries := make([]diagnostic.Annotation, len(b.Instances))
	for i, instance := range b.Instances {
		if instance.AST.Default == nil {
			primaries[i] = anno.Node(b.Component.File, instance.AST, "no default -> required")
		} else {
			primaries[i] = anno.Node(b.Component.File, instance.AST, "default set -> not required")
		}
	}

	logger.
		WithGroup("required").
		Error("Block has conflicting defaults, cannot determine if it is required")
	ch.Report(&diagnostic.Diagnostic{
		Message: "block: cannot determine if block is required: conflicting defaults",
		Primary: primaries,
		Explanation: "For every instance of a block with the same name, all instances must either " +
			"have a default value, or none of them must have a default value. " +
			"If all instances have a default value, the block is considered not required, " +
			"otherwise it is considered required.",
		Hints: []diagnostic.Hint{
			{
				Hint:    "To insert nothing, if the block is not set, use `{}` as default",
				Example: "`block " + string(b.Name) + " {}`",
			},
		},
	})
}
