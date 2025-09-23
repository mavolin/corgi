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

	z.CheckComponent_CallCycles(c)

	z.AnalyzeComponent_Parameters(logger, c)

	z.AnalyzeComponent_AlwaysForwardsReceivedAttributes(c)
	z.AnalyzeComponent_AlwaysAcceptsAttributes(c)

	z.AnalyzeComponent_AST(ctx, c)

	c.Analyzed = true
}

// AnalyzeComponentCall_Component analyzes the component of the passed component call.
//
// It does nothing if the component has already been analyzed or the component
// call is part of a cycle.
//
// Depends on Checks: None
//
// Sets Fields: None
//
// Depends on Fields: None
func (z *analyzer) AnalyzeComponentCall_Component(ctx context.Context, cc *file.ComponentCall) {
	if cc.Component == nil || cc.Component.Analyzed || cc.Circular {
		return
	}

	logger := z.Logger.WithGroup("components")
	z.AnalyzeComponent(ctx, logger, cc.Component)
}

// ============================================================================
// AST-related Analyses
// ======================================================================================

// AnalyzeComponent_AST runs all analyses that require knowledge of their
// position in the AST.
//
// Depends on Checks: None
//
// Sets Fields: None
//
// Depends on Fields: None
func (z *analyzer) AnalyzeComponent_AST(ctx context.Context, c *file.Component) {
	var cannotAttributes file.AnalysisWithReason[ast.AttributeInhibitor]
	cannotAttributes.SetFalse()

	walk.Walk(c.AST, func(w *walk.Context) walk.Action {
		switch n := w.Node.(type) {
		case *ast.Block:
			bi := c.BlockInstanceByNode(n)
			z.AnalyzeBlockInstance(ctx, w.Parents, bi, cannotAttributes)
		}
		return walk.Continue
	}, z.cannotAttributes(ctx, c.File, &cannotAttributes))
}

// ============================================================================
// Component Call Cycles
// ======================================================================================

// CheckComponent_CallCycles checks for component call cycles in the given
// component.
//
// Depends on Checks: None
//
// Sets Fields: None
//
// Depends on Fields: None
func (z *analyzer) CheckComponent_CallCycles(c *file.Component) {
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
// Always Forwards Received Attributes
// ======================================================================================

// AnalyzeComponent_AlwaysForwardsReceivedAttributes attempts to find the first
// permanent top-level &-placeholder component in the given component.
//
// Depends on Checks: None
//
// Sets Fields:
//   - Components.AlwaysForwardsReceivedAttributes
//
// Depends on Fields: None
func (z *analyzer) AnalyzeComponent_AlwaysForwardsReceivedAttributes(c *file.Component) {
	// todo
}

// ============================================================================
// Always Accepts Attributes
// ======================================================================================

// AnalyzeComponent_AlwaysAcceptsAttributes attempts to find the first permanent
// &-placeholder component in the given component.
//
// Depends on Checks: None
//
// Sets Fields:
//   - Components.AlwaysAcceptsAttributes
//
// Depends on Fields:
//   - Components.AlwaysForwardsReceivedAttributes
func (z *analyzer) AnalyzeComponent_AlwaysAcceptsAttributes(c *file.Component) {
	if c.AlwaysForwardsReceivedAttributes.True() {
		c.AlwaysAcceptsAttributes.SetReason(c.AlwaysForwardsReceivedAttributes.Reason())
		return
	}

	// todo
}

// ============================================================================
// Always Forwards Attributes
// ======================================================================================

// AnalyzeComponent_AlwaysForwardsAttributes attempts to find the first
// attribute writer not part of a block default.
func (z *analyzer) AnalyzeComponent_AlwaysForwardsAttributes(c *file.Component) {
	// todo
}

// ============================================================================
// Always Writes Content
// ======================================================================================

// AnalyzeComponent_AlwaysWritesContent attempts to find the first content
// writer not part of a block default.
func (z *analyzer) AnalyzeComponent_AlwaysWritesContent(c *file.Component) {
	// todo
}

// ============================================================================
// Always Writes Elements
// ======================================================================================

// AnalyzeComponent_AlwaysWritesElements attempts to find the first element
// writer not part of a block default.
func (z *analyzer) AnalyzeComponent_AlwaysWritesElements(c *file.Component) {
	// todo
}

// ============================================================================
// Permanent Elements With &-Placeholder
// ======================================================================================

// AnalyzeComponent_PermanentElementsWithAndPlaceholder finds all permanent
// elements with an &-placeholder.
func (z *analyzer) AnalyzeComponent_PermanentElementsWithAndPlaceholder(c *file.Component) {
	// todo
}

// ============================================================================
// Permanent Element Specs With &-Placeholder
// ======================================================================================

// AnalyzeComponent_PermanentElementSpecsWithAndPlaceholder finds all unique
// permanent element specs with an &-placeholder.
func (z *analyzer) AnalyzeComponent_PermanentElementSpecsWithAndPlaceholder(c *file.Component) {
	// todo
}
