package analyze

import (
	"context"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/walk"
)

// isForwarded returns whether the node with the passed parents is forwarded.
//
// Depends on Checks: None
//
// Sets Fields: None
//
// Depends on Fields:
//   - Components.ComponentCall.ForwardsDelegatedAttributes
func (z *analyzer) isForwarded(ctx context.Context, f *file.File, parents []*walk.Context) file.Analysis[bool] {
	i := len(parents) - 1
	for i >= 0 {
		parent := parents[i]
		switch parent := parent.Node.(type) {
		case *ast.Element:
			return file.Result(false)
		case *ast.Doctype:
			return file.Result(false)
		case *ast.ComponentCall:
			// the node we're analyzing must be an attribute, or something
			// yielding an attribute

			cc := f.ComponentCallByNode(parent)
			z.AnalyzeComponentCall(ctx, cc)
			if !cc.ForwardsDelegatedAttributes.Equal(true) {
				return cc.ForwardsDelegatedAttributes
			}
		case ast.BlockSetter:
			ccI := walk.ClosestIndex[*ast.ComponentCall](parents[:i])
			if ccI < 0 {
				return file.FailedAnalysis[bool]()
			}

			ccAST := parents[ccI].Node.(*ast.ComponentCall)
			cc := f.ComponentCallByNode(ccAST)
			z.AnalyzeComponentCall(ctx, cc)

			s := cc.BlockSetterByName(parent.Name())
			if s.Block == nil {
				return file.FailedAnalysis[bool]()
			} else if !s.Block.Forwarded.Equal(true) {
				return s.Block.Forwarded
			}
			i = ccI - 1 // continue with the parent of the component call
		default:
			i--
		}
	}
	return file.Result(true)
}
