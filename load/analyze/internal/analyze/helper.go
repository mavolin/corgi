package analyze

import (
	"context"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/walk"
)

// isNotForwarded returns whether the node with the passed parents is forwarded.
//
// Depends on Checks: None
//
// Sets Fields: None
//
// Depends on Fields:
//   - Components.ComponentCall.ForwardsReceivedAttributes
func (z *analyzer) isNotForwarded(ctx context.Context, f *file.File, parents []*walk.Context) (a file.AnalysisWithReason[ast.ElementWriter]) {
	a.SetFalse()

	i := len(parents) - 1
	for i >= 0 {
		parent := parents[i]
		switch parent := parent.Node.(type) {
		case *ast.Element:
			a.SetReason(parent)
			return a
		case *ast.ComponentCall:
			// the node we're analyzing must be an attribute, or something
			// yielding an attribute

			cc := f.ComponentCallByNode(parent)
			z.AnalyzeComponentCall(ctx, cc)
			if cc.ForwardsReceivedAttributes.False() {
				a.SetReason(cc.AST)
				return a
			}
		case ast.BlockSetter:
			ccI := walk.ClosestIndex[*ast.ComponentCall](parents[:i])
			if ccI < 0 {
				a.SetFailed()
				i = ccI - 1 // continue with the parent of the component call
				continue
			}

			ccAST := parents[ccI].Node.(*ast.ComponentCall)
			cc := f.ComponentCallByNode(ccAST)
			z.AnalyzeComponentCall(ctx, cc)

			s := cc.BlockSetterByName(parent.Name())
			if s.Block == nil {
				a.SetFailed()
			} else if s.Block.Forwarded.False() {
				a.SetReason(cc.AST)
				return a
			}
			i = ccI - 1 // continue with the parent of the component call
		default:
			i--
		}
	}

	return a
}
