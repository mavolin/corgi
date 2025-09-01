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
	ctx context.Context, logger *slog.Logger, c *file.Component, parents []*walk.Context, b *file.BlockInstance,
) {
	block := c.BlockByNode(biAST)

	logger = logger.
		WithGroup("blocks").
		With(slog.String("block", block.Name))

	z.AnalyzeBlockInstanceParentInformation(ctx, c, parents)
	z.AnalyzeBlockInstanceCannotForwardAttributes(c, cannotAttributes, n)

}

// ============================================================================
// Forwarded
// ======================================================================================

func (z *analyzer) AnalyzeBlockInstancesForwarded(b *file.Block) {

}

func (z *analyzer) analyzeBlockInstanceForwarded(b *file.Block) {

}

// AnalyzeBlockInstanceParentInformation determines whether the parent
// information of the given block instance.
//
// Depends on Checks: None
//
// Sets Fields:
//   - Components.Blocks.Instances.Forwarded
//
// Depends on Fields: None
func (z *analyzer) AnalyzeBlockInstanceParentInformation(ctx context.Context, c *file.Component, parents []*walk.Context, biAST *ast.Block) {
	bi := c.BlockInstanceByNode(biAST)
	if bi == nil {
		return
	}

	var (
		element *file.ElementReference
		cc      *file.ComponentCall
		block   *file.Block
	)
	bi.Forwarded.SetResult(true)

	i := len(parents) - 1
Loop:
	for i >= 0 {
		parent := parents[i]
		switch parent := parent.Node.(type) {
		case *ast.ComponentCall:
			bi.Forwarded.SetFailed()
			bi.ContainingElements.SetFailed()
			return
		case ast.BlockSetter:
			if cc != nil {
				continue
			}

			ccI := walk.ClosestIndex[*ast.ComponentCall](parents[:i])
			if ccI < 0 {
				bi.Forwarded.SetFailed()
				bi.ContainingElements.SetFailed()
				continue
			}

			ccAST := parents[ccI].Node.(*ast.ComponentCall) //nolint:errcheck
			cc := c.File.ComponentCallByNode(ccAST)
			z.AnalyzeComponentCall(ctx, cc)

			s := cc.BlockSetterByName(parent.Name())
			if s == nil || s.Block == nil {
				bi.Forwarded.SetFailed()
				bi.ContainingElements.SetFailed()
			} else if s.Block.Forwarded.False() {
				bi.Forwarded.SetResult(false)
				break Loop
			}
			i = ccI - 1 // continue with the parent of the component call
			continue
		case *ast.Element:
			bi.Forwarded.SetResult(false)
			element = c.File.ElementReferenceByNode(parent.Header.Name)
			break Loop
		}
		i--

	}

	if bi.ContainingElements.Failed() {
		return
	} else if cc == nil {
		var containingElements []*file.ElementReference
		if element != nil {
			containingElements = []*file.ElementReference{element}
		}
		bi.ContainingElements.SetResult(&containingElements)
		return
	}

	containingElements := make([]*file.ElementReference, 0, 1+len(block.Instances))
	if block.Forwarded.True() && element != nil {
		containingElements = append(containingElements, element)
	}

	added := make(map[*file.ElementReference]bool)
	for _, inst := range block.Instances {
		if inst.ContainingElements.Failed() {
			bi.ContainingElements.SetFailed()
			return
		}
		for _, el := range *inst.ContainingElements.Result() {
			if !added[el] {
				containingElements = append(containingElements, el)
				added[el] = true
			}
		}
	}
	bi.ContainingElements.SetResult(&containingElements)
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
