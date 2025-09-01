package analyze

import (
	"context"
	"log/slog"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/walk"
)

// AnalyzeBlockInstance analyzes the given block.
//
// Depends on Checks: None
//
// Sets Fields: None
//
// Depends on Fields: None
func (z *analyzer) AnalyzeBlockInstance(
	ctx context.Context, logger *slog.Logger, c *file.Component, parents []*walk.Context, bi *file.BlockInstance,
) {
	logger = logger.
		WithGroup("blocks").
		With(slog.String("block", bi.Group.Name))

	z.AnalyzeBlockInstanceForwarded(ctx, c, parents, bi)

}

// ============================================================================
// Forwarded
// ======================================================================================

// AnalyzeBlockInstanceForwarded determines whether the given block instance
// is forwarded (i.e. placed outside any element).
//
// Depends on Checks: None
//
// Sets Fields:
//   - Components.Blocks.Instances.Forwarded
//
// Depends on Fields: None
func (z *analyzer) AnalyzeBlockInstanceForwarded(ctx context.Context, c *file.Component, parents []*walk.Context, bi *file.BlockInstance) {
	bi.Forwarded.SetResult(true)

	i := len(parents) - 1
	for i >= 0 {
		parent := parents[i]
		switch parent := parent.Node.(type) {
		case *ast.ComponentCall:
			// Continue checking: if the instance has another element as
			// parent, we can still be sure it's not forwarded.
			bi.Forwarded.SetFailed()
		case ast.BlockSetter:
			ccI := walk.ClosestIndex[*ast.ComponentCall](parents[:i])
			if ccI < 0 {
				bi.Forwarded.SetFailed()
				continue
			}

			ccAST := parents[ccI].Node.(*ast.ComponentCall) //nolint:errcheck
			cc := c.File.ComponentCallByNode(ccAST)
			z.AnalyzeComponentCall(ctx, cc)

			s := cc.BlockSetterByName(parent.Name())
			if s == nil || s.Block == nil {
				// Continue checking: if the instance has another element as
				// parent, we can still be sure it's not forwarded.
				bi.Forwarded.SetFailed()
			} else if s.Block.Forwarded.False() {
				bi.Forwarded.SetResult(false)
				return
			}
			i = ccI - 1 // continue with the parent of the component call
			continue
		case *ast.Element:
			bi.Forwarded.SetResult(false)
			return
		}
		i--
	}
}

// ============================================================================
// Cannot Forward Attributes
// ======================================================================================

// AnalyzeBlockInstanceCannotForwardAttributes determines whether the given
// block instance could not forward attributes to the element containing it.
//
// Depends on Checks: None
//
// Sets Fields:
//   - Components.Blocks.Instances.CannotForwardAttributes
//
// Depends on Fields: None
func (z *analyzer) AnalyzeBlockInstanceCannotForwardAttributes(bi *file.BlockInstance, cannotAttributes file.AnalysisWithReason[ast.ContentWriter]) {
	bi.CannotForwardAttributes = cannotAttributes
}
