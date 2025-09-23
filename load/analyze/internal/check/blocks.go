package check

import (
	"log/slog"

	"github.com/mavolin/corgi/v2/escape/elemtype"
	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	"github.com/mavolin/corgi/v2/file/switches"
)

func (ch *checker) CheckBlocks(logger *slog.Logger, c *file.Component) {
	logger = logger.WithGroup("blocks")

	for _, block := range c.Blocks {
		logger := logger.With(slog.String("block", string(block.Name)))

		ch.CheckBlockNoConflictingRequired(logger, block)
	}
}

// ============================================================================
// No Conflicting Required
// ======================================================================================

func (ch *checker) CheckBlockNoConflictingRequired(logger *slog.Logger, b *file.Block) {
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

// ============================================================================
// Instances In Allowed Elements
// ======================================================================================

func (ch *checker) CheckBlockInstanceInAllowedElement(logger *slog.Logger, bi *file.BlockInstance) {
	if bi.ContainingElements.Failed() {
		return
	}

	f := bi.Group.Component.File
	for _, e := range bi.ContainingElements.Result().Get() {
		switches.ContainingElement(e,
			// Handled when CheckBlockInstanceInAllowedElement is called on the
			// component for that call.
			func(*ast.BlockSetterContainingElement) {},
			func(e *ast.Element) {
				ref := f.ElementReferenceByNode(e.Header.Name)
				if ref.Spec.Type.Failed() {
					return
				}

				switch ref.Spec.Type.Result() {
				case elemtype.JS:
					logger.
						WithGroup("checks.not_in_script").
						Error("Block instance in JS-typed element", slog.String("block", string(bi.Group.Name)))
					ch.Report(&diagnostic.Diagnostic{
						Message: "block placed in `js`-typed element",
						Primary: []diagnostic.Annotation{
							anno.Node(f, bi.AST, "cannot place `block` here"),
						},
						Explanation: "You cannot place blocks inside `js`-typed elements.",
					})
				case elemtype.CSS:
					logger.
						WithGroup("checks.not_in_script").
						Error("Block instance in CSS-typed element", slog.String("block", string(bi.Group.Name)))
					ch.Report(&diagnostic.Diagnostic{
						Message: "block placed in `css`-typed element",
						Primary: []diagnostic.Annotation{
							anno.Node(f, bi.AST, "cannot place `block` here"),
						},
						Explanation: "You cannot place blocks inside `css`-typed elements.",
					})
				case elemtype.Unknown, elemtype.Void, elemtype.Nothing, elemtype.Text, elemtype.Normal: // for linting
					// do nothing
				}
			})
	}
}
