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
		slog.String("file", c.File.Name),
		slog.String("comp", c.AST.Header.Name.Name),
		slog.String("comp_pos", c.AST.Start().String()))

	z.CheckComponentCallCycles(c)

	z.AnalyzeComponentParameters(logger, c)

	z.AnalyzeComponentAST(ctx, c)

	z.AnalyzeBlocks(logger, c)

	z.AnalyzeCouldForwardAttributes(c)
	z.AnalyzeCouldAcceptAttributes(c)
	z.FindFirstPermanentForwardedAndPlaceholderWriter(c)
	z.FindFirstPermanentAndPlaceholderWriter(c)

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
func (z *analyzer) AnalyzeComponentAST(ctx context.Context, c *file.Component) {
	walk.Walk(c.AST, func(wctx *walk.Context) walk.Action {
		switch n := wctx.Node.(type) {
		case *ast.Block:
			z.AnalyzeBlockInstanceForwarded(ctx, c, wctx.Parents, n)
		}
		return walk.Continue
	})
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
		if cc.Component == nil {
			continue
		} else if cc.Component.File.Package != root.File.Package {
			// The only way a component call causing a component call cycle can
			// be in a different package is if we have an import cycle, which
			// should've been caught elsewhere.
			continue
		}

		if cc.Component == root {
			chain = append(chain, cc)
			for _, call := range chain {
				call.Circular = true
			}

			secondaries := make([]diagnostic.Annotation, 1, 1+len(chain))
			secondaries[0] = anno.Node(root.File, root.AST, "in this component")

			prev := cc.Component
			for i, cc := range chain[1:] {
				annotation := fmt.Sprint(i+1, ": `"+prev.AST.Header.Name.Name+"` calls `"+cc.AST.Header.Name.Full()+"`")
				secondaries = append(secondaries, anno.Node(cc.File, cc.AST, annotation))
				prev = cc.Component
			}

			annotation := "call to itself"
			if len(chain) > 1 {
				annotation = "calls `" + root.AST.Header.Name.Name + "`"
			}

			z.Report(&diagnostic.Diagnostic{
				Message: "component call cycle",
				Primary: []diagnostic.Annotation{
					anno.Node(root.File, chain[0].AST, annotation),
				},
				Secondary: secondaries,
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

// AnalyzeCouldForwardAttributes attempts to see if the given component could
// forward the attributes it receives to the element containing it.
//
// Depends on Checks: None
//
// Sets Fields:
//   - Components.CouldForwardAttributes
//
// Depends on Fields:
//   - Components.Blocks.Instances.Forwarded
//   - Components.Blocks.Instances.Default.FirstForwardedAndPlaceholderWriter
func (z *analyzer) AnalyzeCouldForwardAttributes(c *file.Component) {
	ap := c.FirstPermanentForwardedAndPlaceholderWriter
	if ap.NotZero() {
		c.CouldForwardAttributes.Set(true)
		return
	}
	failed := ap.Failed

	for _, block := range c.Blocks {
		for _, instance := range block.Instances {
			firstForwardedAndPlaceholder := file.ConditionalAnalysis(instance.Forwarded, instance.Default.FirstForwardedAndPlaceholderWriter)
			if firstForwardedAndPlaceholder.NotZero() {
				c.CouldForwardAttributes.Set(true)
				return
			}
			failed = failed || firstForwardedAndPlaceholder.Failed
		}
	}

	c.CouldForwardAttributes.SetIf(false, !failed)
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
//   - Components.CouldForwardAttributes
//   - Components.Blocks.Instances.Default.FirstAndPlaceholderWriter
func (z *analyzer) AnalyzeCouldAcceptAttributes(c *file.Component) {
	if c.CouldForwardAttributes.NotZero() {
		c.CouldAcceptAttributes.Set(true)
		return
	}

	ap := c.FirstPermanentAndPlaceholderWriter
	if ap.NotZero() {
		c.CouldAcceptAttributes.Set(true)
		return
	}
	failed := ap.Failed

	for _, block := range c.Blocks {
		for _, instance := range block.Instances {
			firstAndPlaceholder := file.ConditionalAnalysis(instance.Forwarded, instance.Default.FirstAndPlaceholderWriter)
			if firstAndPlaceholder.NotZero() {
				c.CouldAcceptAttributes.Set(true)
				return
			}
			failed = failed || firstAndPlaceholder.Failed
		}
	}

	c.CouldAcceptAttributes.SetIf(false, !failed)
}

// ============================================================================
// First Permanent Top-Level &-Placeholder
// ======================================================================================

// FindFirstPermanentForwardedAndPlaceholderWriter attempts to find the first
// permanent top-level &-placeholder component in the given component.
//
// Depends on Checks: None
//
// Sets Fields:
//   - Components.FirstPermanentForwardedAndPlaceholderWriter
//
// Depends on Fields: None
func (z *analyzer) FindFirstPermanentForwardedAndPlaceholderWriter(c *file.Component) {
	z.AnalyzeComponentCall(context.Background(), c.ComponentCalls[0])
	// todo
}

// ============================================================================
// First Permanent &-Placeholder
// ======================================================================================

// FindFirstPermanentAndPlaceholderWriter attempts to find the first permanent
// &-placeholder component in the given component.
//
// Depends on Checks: None
//
// Sets Fields:
//   - Components.FirstPermanentAndPlaceholderWriter
//
// Depends on Fields:
//   - Components.FirstPermanentForwardedAndPlaceholderWriter
func (z *analyzer) FindFirstPermanentAndPlaceholderWriter(c *file.Component) {
	if c.FirstPermanentForwardedAndPlaceholderWriter.NotZero() {
		c.FirstPermanentAndPlaceholderWriter = c.FirstPermanentForwardedAndPlaceholderWriter
		return
	}

	// todo
}
