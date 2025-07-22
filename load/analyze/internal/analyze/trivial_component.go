package analyze

import (
	"fmt"
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/file/diagnostic/anno"
)

// TrivialAnalyzeComponents runs all trivial analyses on the components.
//
// An analysis is trivial, if it does not depend on the analysis of a component
// call or the analysis of another component.
//
// Depends on Checks: None
//
// Sets Fields: None
//
// Depends on Fields: None
func (z *analyzer) TrivialAnalyzeComponents() {
	logger := z.Logger.WithGroup("components")
	logger.Debug("Analyzing components")

	for _, c := range z.P.Components {
		logger := logger.With(
			slog.String("file", c.File.Name),
			slog.String("comp", c.Header().Name.Name),
			slog.String("comp_pos", c.Start().String()))

		z.CheckComponentCallCycles(c)

		z.AnalyzeComponentParameters(logger, c)
		z.TrivialAnalyzeBlocks(logger, c)
	}
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
	if c.AnalyzedWithErrors {
		// Possibly a circular alias
		return
	}

	z.checkComponentCallCycles(c, make([]*file.ComponentCall, 0, 24), c)
}

func (z *analyzer) checkComponentCallCycles(root *file.Component, ccs []*file.ComponentCall, c *file.Component) {
	for _, cc := range c.ComponentCalls {
		if cc.Component == root {
			ccs = append(ccs, cc)
			secondaries := make([]diagnostic.Annotation, len(ccs))
			prev := root
			for i, cc := range ccs {
				annotation := fmt.Sprint(i+1, ": `"+prev.DefinedAST.Header.Name.Name+"` calls `"+cc.AST.Header.Name.Full()+"`")
				secondaries[i] = anno.Node(cc.File, cc.AST.Header.Name, annotation)
				prev = cc.Component
			}

			z.Report(&diagnostic.Diagnostic{
				Message: "component call cycle",
				Primary: []diagnostic.Annotation{
					anno.Node(root.File, root.DefinedAST, "this component calls itself"),
				},
				Secondary: secondaries,
				Explanation: "This component recursively calls itself, which is not allowed.\n" +
					"To fix this error, you need to break the chain of recursion.",
			})
			root.AnalyzedWithErrors = true
			continue
		}

		if cc.Component.File.Package != root.File.Package {
			// The only way a component call causing a component call cycle can
			// be in a different package is if we have an import cycle, which
			// should've been caught elsewhere.
			continue
		}

		z.checkComponentCallCycles(root, append(ccs, cc), cc.Component)
	}
}
