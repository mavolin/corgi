package analyze

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
	"github.com/mavolin/corgi/v2/file/walk"
)

// AnalyzeComponents analyzes all components in the package.
//
// Depends on Checks: None
//
// Sets Fields: None
//
// Depends on Fields: None
func (z *analyzer) AnalyzeComponents() {
	logger := z.Logger.WithGroup("components")
	logger.Debug("Analyzing components")

	ctx := context.Background()
	for _, c := range z.P.Components {
		z.AnalyzeComponent(ctx, logger, c)
	}
}

// AnalyzeComponent analyzes the passed component.
//
// Depends on Checks: None
//
// Sets Fields: None
//
// Depends on Fields: None
func (z *analyzer) AnalyzeComponent(ctx context.Context, logger *slog.Logger, c *file.Component) {
	logger = logger.With(
		slog.String("file", string(c.File.Name)),
		slog.String("comp", c.AST.Header.Name.Name),
		slog.String("comp_pos", c.AST.Start().String()))

	z.CheckComponentCallCycles(c)

	z.AnalyzeComponentParameters(logger, c)

	z.AnalyzeComponentAST(ctx, logger, c)

	z.AnalyzeCouldForwardReceivedAttributes(c)
	z.AnalyzeCouldAcceptAttributes(c)
	z.AnalyzeAlwaysForwardsAndPlaceholder(c)
	z.AnalyzedAlwaysWritesAndPlaceholder(c)

	c.Analyzed = true
}

// AnalyzeCallComponent analyzes the component of the passed component call.
//
// It does nothing if the component has already been analyzed or the component
// call is part of a cycle.
//
// Depends on Checks: None
//
// Sets Fields: None
//
// Depends on Fields: None
func (z *analyzer) AnalyzeCallComponent(ctx context.Context, cc *file.ComponentCall) {
	if cc.Component == nil || cc.Component.Analyzed || cc.Circular {
		return
	}

	logger := z.Logger.WithGroup("components")
	z.AnalyzeComponent(ctx, logger, cc.Component)
}

// ============================================================================
// AST-related Analyses
// ======================================================================================

// AnalyzeComponentAST runs all analyses that require knowledge of their
// position in the AST.
//
// Depends on Checks: None
//
// Sets Fields: None
//
// Depends on Fields: None
func (z *analyzer) AnalyzeComponentAST(ctx context.Context, logger *slog.Logger, c *file.Component) {
	var cannotAttributes file.AnalysisWithReason[ast.AttributeInhibitor]
	cannotAttributes.SetFalse()

	walk.Walk(c.AST, func(w *walk.Context) walk.Action {
		switch n := w.Node.(type) {
		case *ast.Block:
			bi := c.BlockInstanceByNode(n)
			z.AnalyzeBlockInstanceCannotForwardAttributes(bi, cannotAttributes)
			z.AnalyzeBlockInstance(ctx, logger, c, w.Parents, bi)
		}
		return walk.Continue
	}, z.cannotAttributes(ctx, c.File, &cannotAttributes))
}

// ============================================================================
// Component Call Cycles
// ======================================================================================

// CheckComponentCallCycles checks for component call cycles in the given
// component.
//
// Depends on Checks: None
//
// Sets Fields: None
//
// Depends on Fields: None
func (z *analyzer) CheckComponentCallCycles(c *file.Component) {
	z.checkComponentCallCycles(c, make([]*file.ComponentCall, 0, 24), c)
}

func (z *analyzer) checkComponentCallCycles(root *file.Component, chain []*file.ComponentCall, c *file.Component) {
	for _, cc := range c.ComponentCalls {
		switch {
		case cc.Component == nil:
			continue
		case cc.Component.File.Package != root.File.Package:
			continue
		case cc.Circular:
			continue
		}

		if cc.Component == root {
			chain = append(chain, cc)
			for _, call := range chain {
				call.Circular = true
			}

			primaries := make([]diagnostic.Annotation, 1, len(chain))

			prev := cc.Component
			for i, cc := range chain {
				annotation := fmt.Sprint(i+1, ": `"+prev.AST.Header.Name.Name+"` calls `"+cc.AST.Header.Name.Full()+"`")
				primaries = append(primaries, anno.Node(cc.File, cc.AST, annotation))
				prev = cc.Component
			}

			z.Report(&diagnostic.Diagnostic{
				Message: "component call cycle",
				Primary: primaries,
				Secondary: []diagnostic.Annotation{
					anno.Node(root.File, root.AST, "in this component"),
				},
				Explanation: "This component recursively calls itself, which is not allowed.\n" +
					"To fix this error, you need to break the chain of recursion.",
			})
			continue
		}

		z.checkComponentCallCycles(root, append(chain, cc), cc.Component)
	}
}

// ============================================================================
// Could Forward Attributes
// ======================================================================================

// AnalyzeCouldForwardReceivedAttributes attempts to see if the given component could
// forward the attributes it receives to the element containing it.
//
// Depends on Checks: None
//
// Sets Fields:
//   - Components.CouldForwardReceivedAttributes
//
// Depends on Fields:
//   - Components.AlwaysForwardsReceivedAttributes
//   - Components.Blocks.Instances.Forwarded
//   - Components.Blocks.Instances.Default.ForwardsReceivedAttributes
func (z *analyzer) AnalyzeCouldForwardReceivedAttributes(c *file.Component) {
	if c.AlwaysForwardsReceivedAttributes.True() {
		c.CouldForwardReceivedAttributes.SetReason(c.AlwaysForwardsReceivedAttributes.Reason())
		return
	}

	c.CouldForwardReceivedAttributes.SetFalse()
	if c.AlwaysForwardsReceivedAttributes.Failed() {
		c.CouldForwardReceivedAttributes.SetFailed()
	}

	for _, block := range c.Blocks {
		for _, instance := range block.Instances {
			forwardsAndPlaceholder := file.ConditionalAnalysis(instance.Forwarded, instance.Default.ForwardsReceivedAttributes)
			if forwardsAndPlaceholder.True() {
				c.CouldForwardReceivedAttributes.SetReason(forwardsAndPlaceholder.Reason())
				return
			} else if forwardsAndPlaceholder.Failed() {
				c.CouldForwardReceivedAttributes.SetFailed()
			}
		}
	}
}

// ============================================================================
// Could Accept Attributes
// ======================================================================================

// AnalyzeCouldAcceptAttributes attempts to see if the given component could
// accept attributes.
//
// Depends on Checks: None
//
// Sets Fields:
//   - Components.CouldAcceptAttributes
//
// Depends on Fields:
//   - Components.CouldForwardReceivedAttributes
//   - Components.Blocks.Instances.Default.AcceptsAttributes
func (z *analyzer) AnalyzeCouldAcceptAttributes(c *file.Component) {
	if c.CouldForwardReceivedAttributes.True() {
		c.CouldAcceptAttributes.SetReason(c.CouldForwardReceivedAttributes.Reason())
		return
	}

	if c.AlwaysAcceptsAttributes.True() {
		c.CouldAcceptAttributes.SetReason(c.AlwaysAcceptsAttributes.Reason())
		return
	}

	c.CouldAcceptAttributes.SetFalse()
	if c.AlwaysAcceptsAttributes.Failed() {
		c.CouldAcceptAttributes.SetFailed()
	}

	for _, block := range c.Blocks {
		for _, instance := range block.Instances {
			if instance.Default.AcceptsAttributes.True() {
				c.CouldAcceptAttributes.SetReason(instance.Default.AcceptsAttributes.Reason())
				return
			} else if instance.Default.AcceptsAttributes.Failed() {
				c.CouldAcceptAttributes.SetFailed()
			}
		}
	}
}

// ============================================================================
// First Permanent Top-Level &-Placeholder
// ======================================================================================

// AnalyzeAlwaysForwardsAndPlaceholder attempts to find the first
// permanent top-level &-placeholder component in the given component.
//
// Depends on Checks: None
//
// Sets Fields:
//   - Components.AlwaysForwardsReceivedAttributes
//
// Depends on Fields: None
func (z *analyzer) AnalyzeAlwaysForwardsAndPlaceholder(c *file.Component) {
	// todo
}

// ============================================================================
// First Permanent &-Placeholder
// ======================================================================================

// AnalyzedAlwaysWritesAndPlaceholder attempts to find the first permanent
// &-placeholder component in the given component.
//
// Depends on Checks: None
//
// Sets Fields:
//   - Components.AlwaysAcceptsAttributes
//
// Depends on Fields:
//   - Components.AlwaysForwardsReceivedAttributes
func (z *analyzer) AnalyzedAlwaysWritesAndPlaceholder(c *file.Component) {
	if c.AlwaysForwardsReceivedAttributes.True() {
		c.AlwaysAcceptsAttributes.SetReason(c.AlwaysForwardsReceivedAttributes.Reason())
		return
	}

	// todo
}
