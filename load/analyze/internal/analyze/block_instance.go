package analyze

import (
	"context"
	"log/slog"
	"slices"

	"github.com/mavolin/corgi/v2/escape/elemtype"
	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/switches"
	"github.com/mavolin/corgi/v2/file/walk"
	"github.com/mavolin/corgi/v2/load/analyze/internal/candidate"
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
	z.AnalyzeBlockInstanceContainingElements(ctx, c, parents, bi)
	z.AnalyzeBlockInstanceContainingElementSpecs(c.File, bi)
	z.AnalyzeBlockInstanceElementType(bi)
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
		candidate.SwitchContainingElement(parents[i].Node,
			func(*ast.Element) {
				bi.Forwarded.SetResult(false)
			},
			func(*ast.ComponentCall) {
				// Continue checking: if the instance has another element as
				// parent, we can still be sure it's not forwarded.
				bi.Forwarded.SetFailed()
			},
			func(parent ast.BlockSetter) {
				ccI := walk.ClosestIndex[*ast.ComponentCall](parents[:i])
				if ccI < 0 {
					bi.Forwarded.SetFailed()
					return
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
				i = ccI // continue with the parent of the component call
			})
		if bi.Forwarded.Equal(false) {
			break
		}
		i--
	}
}

// ============================================================================
// Containing Elements
// ======================================================================================

// AnalyzeBlockInstanceContainingElements determines all elements containing
// the given block instance.
//
// Depends on Checks: None
//
// Sets Fields:
//   - Components.Blocks.Instances.ContainingElements
//
// Depends on Fields: None
func (z *analyzer) AnalyzeBlockInstanceContainingElements(ctx context.Context, c *file.Component, parents []*walk.Context, bi *file.BlockInstance) {
	bi.ContainingElements.SetZero()
	var containingElements []ast.ContainingElement

	i := len(parents) - 1
	for i >= 0 {
		done := candidate.SwitchContainingElementR(parents[i].Node,
			func(*ast.Element) bool {
				containingElements = append(containingElements)
				return true
			},
			func(*ast.ComponentCall) bool {
				bi.ContainingElements.SetFailed()
				return true
			},
			func(parent ast.BlockSetter) bool {
				ccI := walk.ClosestIndex[*ast.ComponentCall](parents[:i])
				if ccI < 0 {
					bi.ContainingElements.SetFailed()
					return true
				}

				ccAST := parents[ccI].Node.(*ast.ComponentCall) //nolint:errcheck
				cc := c.File.ComponentCallByNode(ccAST)
				z.AnalyzeComponentCall(ctx, cc)

				s := cc.BlockSetterByName(parent.Name())
				if s == nil || s.Block == nil || s.Block.ContainingElements.Failed() {
					bi.ContainingElements.SetFailed()
					return true
				}

				containingElements = append(containingElements, &ast.BlockSetterContainingElement{
					ComponentCall: ccAST,
					BlockSetter:   parent,
				})
				if s.Block.Forwarded.False() {
					return true
				}
				i = ccI // continue with the parent of the component call
				return false
			})
		if done {
			break
		}

		i--
	}

	if !bi.ContainingElements.Failed() {
		containingElements = slices.Clip(containingElements)
		bi.ContainingElements.SetResult(&containingElements)
	}
}

// ============================================================================
// Containing Element Specs
// ======================================================================================

// AnalyzeBlockInstanceContainingElementSpecs calculates the containing element
// specs of the given attribute.
//
// Depends on Checks: None
//
// Sets Fields:
//   - Components.Blocks.Instances.ContainingElementSpecs
//
// Depends on Fields:
//   - Components.Blocks.Instances.ContainingElements
//   - ComponentCalls.ElementSpecsWithAndPlaceholder
//   - Components.Blocks.ContainingElementSpecs
func (z *analyzer) AnalyzeBlockInstanceContainingElementSpecs(f *file.File, bi *file.BlockInstance) {
	if bi.ContainingElements.Failed() {
		bi.ContainingElementSpecs.SetFailed()
		return
	}

	containingElements := *bi.ContainingElements.Result()
	if len(containingElements) == 0 {
		var specs []*file.ElementSpec
		bi.ContainingElementSpecs.SetResult(&specs)
		return
	}

	bi.ContainingElementSpecs.SetZero()

	specsSet := make(map[*file.ElementSpec]struct{}, len(containingElements))
	for _, e := range containingElements {
		switches.ContainingElement(e,
			func(e *ast.AndPlaceholderContainingElement) {
				cc := f.ComponentCallByNode((*ast.ComponentCall)(e))
				if cc.ElementSpecsWithAndPlaceholder.Failed() {
					bi.ContainingElementSpecs.SetFailed()
					return
				}

				for _, spec := range *cc.ElementSpecsWithAndPlaceholder.Result() {
					specsSet[spec] = struct{}{}
				}
			},
			func(e *ast.BlockSetterContainingElement) {
				cc := f.ComponentCallByNode(e.ComponentCall)
				s := cc.BlockSetterByName(e.BlockSetter.Name())
				if s == nil || s.Block == nil || s.Block.ContainingElementSpecs.Failed() {
					bi.ContainingElementSpecs.SetFailed()
					return
				}

				for _, spec := range *s.Block.ContainingElementSpecs.Result() {
					specsSet[spec] = struct{}{}
				}
			},
			func(e *ast.Element) {
				ref := f.ElementReferenceByNode(e.Header.Name)
				if ref.Spec == nil {
					bi.ContainingElementSpecs.SetFailed()
					return
				}
				specsSet[ref.Spec] = struct{}{}
			})
		if bi.ContainingElementSpecs.Failed() {
			return
		}
	}

	specs := make([]*file.ElementSpec, 0, len(specsSet))
	for spec := range specsSet {
		specs = append(specs, spec)
	}

	bi.ContainingElementSpecs.SetResult(&specs)
}

// ============================================================================
// Element Type
// ======================================================================================

// AnalyzeBlockInstanceElementType determines the element type of the given
// block instance.
//
// Depends on Checks:
//   - CheckBlockInstanceInScript
//   - CheckBlockInstanceInStylesheet
//
// Sets Fields:
//   - Components.Blocks.Instances.ElementType
//
// Depends on Fields:
//   - Components.Blocks.Instances.ContainingElements
func (z *analyzer) AnalyzeBlockInstanceElementType(bi *file.BlockInstance) {
	if bi.ContainingElementSpecs.Failed() {
		bi.ElementType.SetFailed()
		return
	}

	containingElementSpecs := *bi.ContainingElementSpecs.Result()
	if len(containingElementSpecs) == 0 {
		bi.ElementType.SetResult(elemtype.Unknown)
		return
	}

	t := elemtype.Normal
	for _, spec := range containingElementSpecs[1:] {
		if spec.Type.Failed() {
			bi.ElementType.SetFailed()
			return
		}

		specType := spec.Type.Result()
		if specType == elemtype.Void {
			specType = elemtype.Nothing
		}
		t = min(t, specType)
	}
	bi.ElementType.SetResult(t)
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
func (z *analyzer) AnalyzeBlockInstanceCannotForwardAttributes(
	bi *file.BlockInstance, cannotAttributes file.AnalysisWithReason[ast.AttributeInhibitor],
) {
	bi.CannotForwardAttributes = cannotAttributes
}
