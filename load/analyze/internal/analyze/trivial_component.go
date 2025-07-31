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
			slog.String("comp", c.AST.Header.Name.Name),
			slog.String("comp_pos", c.AST.Start().String()))

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
