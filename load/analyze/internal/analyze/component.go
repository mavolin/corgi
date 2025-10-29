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

func (z *analyzer) AnalyzeComponents() {
	logger := z.Logger.WithGroup("components")
	logger.Debug("Analyzing components")

	z.AnalyzeComponent_Circular(logger)

	ctx := context.Background()
	for _, c := range z.Pkg.Components {
		z.AnalyzeComponent(ctx, logger, c)
	}
}

func (z *analyzer) AnalyzeComponent(ctx context.Context, logger *slog.Logger, c *file.Component) {
	if c.Analyzed {
		return
	}

	logger = logger.With(
		slog.String("file", string(c.File.Name)),
		slog.String("comp", c.AST.Header.Name.Name),
		slog.String("comp_pos", c.AST.Start().String()))

	z.AnalyzeComponent_Parameters(logger, c)

	z.AnalyzeComponent_AlwaysForwardsReceivedAttributes(c)
	z.AnalyzeComponent_AlwaysAcceptsAttributes(c)

	z.AnalyzeComponent_AST(ctx, c)

	c.Analyzed = true
}

func (z *analyzer) AnalyzeComponentCall_Component(ctx context.Context, cc *file.ComponentCall) {
	if cc.Component == nil || cc.Component.Analyzed || z.component_Circular(cc.Component) {
		return
	}

	logger := z.Logger.WithGroup("components")
	z.AnalyzeComponent(ctx, logger, cc.Component)
}

// ============================================================================
// Circular
// ======================================================================================

func (z *analyzer) AnalyzeComponent_Circular(logger *slog.Logger) {
	for _, c := range z.Pkg.Components {
		if !c.Analyzed {
			c.Circular = false // reset
		}
	}

	chain := make([]*file.ComponentCall, 0, 24)
	for _, c := range z.Pkg.Components {
		if !c.Analyzed {
			chain = chain[:1]
			z.analyzeComponent_Circular(logger, c, chain, c)
			z.Ran(c, component_Circular{})
		}
	}
}

type component_Circular struct{}

func (z *analyzer) analyzeComponent_Circular(logger *slog.Logger, root *file.Component, chain []*file.ComponentCall, last *file.Component) {
	if last.Circular {
		return
	}

	walk.WalkT[*ast.ComponentCall](last.AST.Body, func(w *walk.ContextT[*ast.ComponentCall]) walk.Action {
		cc := last.File.ComponentCallByNode(w.Node)
		switch {
		case cc.Component == nil:
			return walk.Continue
		case cc.Component.File.Package != root.File.Package:
			return walk.Continue
		case cc.Component.Circular:
			return walk.Continue
		}

		if cc.Component != root {
			z.analyzeComponent_Circular(logger, root, append(chain, cc), cc.Component)
			return walk.Continue
		}

		chain = append(chain, cc)
		root.Circular = true
		for _, call := range chain {
			call.Component.Circular = true
		}

		primaries := make([]diagnostic.Annotation, 0, len(chain))

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
		return walk.Continue
	})
}

// ============================================================================
// AST-related Analyses
// ======================================================================================

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
// Always Forwards Received Attributes
// ======================================================================================

type component_AlwaysForwardsReceivedAttributes struct{}

func (z *analyzer) AnalyzeComponent_AlwaysForwardsReceivedAttributes(c *file.Component) {
	defer z.Ran(c, component_AlwaysForwardsReceivedAttributes{})
	// todo
}

// ============================================================================
// Always Accepts Attributes
// ======================================================================================

type component_AlwaysAcceptsAttributes struct{}

func (z *analyzer) AnalyzeComponent_AlwaysAcceptsAttributes(c *file.Component) {
	defer z.Ran(c, component_AlwaysAcceptsAttributes{})

	alwaysForwardsReceivedAttributes := z.component_AlwaysForwardsReceivedAttributes(c)
	if alwaysForwardsReceivedAttributes.True() {
		c.AlwaysAcceptsAttributes.SetReason(alwaysForwardsReceivedAttributes.Reason())
		return
	}

	// todo
}

// ============================================================================
// Always Forwards Attributes
// ======================================================================================

type component_AlwaysForwardsAttributes struct{}

func (z *analyzer) AnalyzeComponent_AlwaysForwardsAttributes(c *file.Component) {
	defer z.Ran(c, component_AlwaysForwardsAttributes{})
	// todo
}

// ============================================================================
// Always Writes Content
// ======================================================================================

type component_AlwaysWritesContent struct{}

func (z *analyzer) AnalyzeComponent_AlwaysWritesContent(c *file.Component) {
	defer z.Ran(c, component_AlwaysWritesContent{})
	// todo
}

// ============================================================================
// Always Writes Elements
// ======================================================================================

type component_AlwaysWritesElements struct{}

func (z *analyzer) AnalyzeComponent_AlwaysWritesElements(c *file.Component) {
	defer z.Ran(c, component_AlwaysWritesElements{})
	// todo
}

// ============================================================================
// Permanent Elements With &-Placeholder
// ======================================================================================

type component_PermanentElementsWithAndPlaceholder struct{}

func (z *analyzer) AnalyzeComponent_PermanentElementsWithAndPlaceholder(c *file.Component) {
	defer z.Ran(c, component_PermanentElementsWithAndPlaceholder{})
	// todo
}

// ============================================================================
// Permanent Element Specs With &-Placeholder
// ======================================================================================

type component_PermanentElementSpecsWithAndPlaceholder struct{}

func (z *analyzer) AnalyzeComponent_PermanentElementSpecsWithAndPlaceholder(c *file.Component) {
	defer z.Ran(c, component_PermanentElementSpecsWithAndPlaceholder{})
	// todo
}
