package analyze

import (
	"context"
	"slices"

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
	ctx context.Context, parents []*walk.Context, bi *file.BlockInstance,
	cannotAttributes file.AnalysisWithReason[ast.AttributeInhibitor],
) {
	z.AnalyzeBlockInstance_Forwarded(ctx, parents, bi)
	z.AnalyzeBlockInstance_ContainingElements(ctx, parents, bi)
	z.AnalyzeBlockInstance_ContainingElementSpecs(bi)

	z.AnalyzeBlockInstance_CannotForwardAttributes(bi, cannotAttributes)

	z.AnalyzeBlockInstanceDefault(ctx, parents, bi)
}

// ============================================================================
// Forwarded
// ======================================================================================

// AnalyzeBlockInstance_Forwarded determines whether the given block instance
// is forwarded (i.e. placed outside any element).
//
// Depends on Checks: None
//
// Sets Fields:
//   - Components.Blocks.Instances.Forwarded
//
// Depends on Fields: None
func (z *analyzer) AnalyzeBlockInstance_Forwarded(ctx context.Context, parents []*walk.Context, bi *file.BlockInstance) {
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

				f := bi.Group.Component.File

				ccAST := parents[ccI].Node.(*ast.ComponentCall) //nolint:errcheck
				cc := f.ComponentCallByNode(ccAST)
				z.AnalyzeComponentCall(ctx, cc)

				s := cc.BlockSetterByNode(parent)
				if s == nil || s.Block == nil {
					// Continue checking: if the instance has another element as
					// parent, we can still be sure it's not forwarded.
					bi.Forwarded.SetFailed()
				} else if s.Block.Forwarded().False() {
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

// AnalyzeBlockInstance_ContainingElements determines all elements containing
// the given block instance.
//
// Depends on Checks: None
//
// Sets Fields:
//   - Components.Blocks.Instances.ContainingElements
//
// Depends on Fields: None
func (z *analyzer) AnalyzeBlockInstance_ContainingElements(ctx context.Context, parents []*walk.Context, bi *file.BlockInstance) {
	bi.ContainingElements.SetResult(file.NilSliceRef[ast.ContainingElement]())
	var containingElements []ast.ContainingElement

	i := len(parents) - 1
	for i >= 0 {
		done := candidate.SwitchContainingElementR(parents[i].Node,
			func(e *ast.Element) bool {
				containingElements = append(containingElements, e)
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

				f := bi.Group.Component.File

				ccAST := parents[ccI].Node.(*ast.ComponentCall) //nolint:errcheck
				cc := f.ComponentCallByNode(ccAST)
				z.AnalyzeComponentCall(ctx, cc)

				s := cc.BlockSetterByNode(parent)
				if s == nil || s.Block == nil {
					bi.ContainingElements.SetFailed()
					return true
				}
				for _, instance := range s.Block.Instances {
					if instance.ContainingElements.Failed() {
						bi.ContainingElements.SetFailed()
						return true
					}
				}

				containingElements = append(containingElements, &ast.BlockSetterContainingElement{
					ComponentCall: ccAST,
					BlockSetter:   parent,
				})
				if s.Block.Forwarded().False() {
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
		bi.ContainingElements.SetResult(file.SliceRefFrom(slices.Clip(containingElements)))
	}
}

// ============================================================================
// Containing Element Specs
// ======================================================================================

// AnalyzeBlockInstance_ContainingElementSpecs calculates the containing element
// specs of the given attribute.
//
// Depends on Checks: None
//
// Sets Fields:
//   - Components.Blocks.Instances.ContainingElementSpecs
//
// Depends on Fields:
//   - Components.Blocks.Instances.ContainingElements
func (z *analyzer) AnalyzeBlockInstance_ContainingElementSpecs(bi *file.BlockInstance) {
	if bi.ContainingElements.Failed() {
		bi.ContainingElementSpecs.SetFailed()
		return
	}

	if bi.ContainingElements.Result().Len() == 0 {
		bi.ContainingElementSpecs.SetResult(file.NilSliceRef[*file.ElementSpec]())
		return
	}

	f := bi.Group.Component.File

	specSet := make(map[*file.ElementSpec]struct{}, bi.ContainingElements.Result().Len())
	for _, e := range bi.ContainingElements.Result().Get() {
		switches.ContainingElement(e,
			func(e *ast.BlockSetterContainingElement) {
				cc := f.ComponentCallByNode(e.ComponentCall)
				s := cc.BlockSetterByNode(e.BlockSetter)
				if s == nil || s.Block == nil {
					bi.ContainingElementSpecs.SetFailed()
					return
				}

				for _, instance := range s.Block.Instances {
					if instance.ContainingElementSpecs.Failed() {
						bi.ContainingElementSpecs.SetFailed()
						return
					}

					for _, spec := range instance.ContainingElementSpecs.Result().Get() {
						specSet[spec] = struct{}{}
					}
				}
			},
			func(e *ast.Element) {
				ref := f.ElementReferenceByNode(e.Header.Name)
				if ref.Spec == nil {
					bi.ContainingElementSpecs.SetFailed()
					return
				}
				specSet[ref.Spec] = struct{}{}
			})
		if bi.ContainingElementSpecs.Failed() {
			return
		}
	}

	specs := make([]*file.ElementSpec, 0, len(specSet))
	for spec := range specSet {
		specs = append(specs, spec)
	}
	bi.ContainingElementSpecs.SetResult(file.SliceRefFrom(specs))
}

// ============================================================================
// Cannot Forward Attributes
// ======================================================================================

// AnalyzeBlockInstance_CannotForwardAttributes determines whether the given
// block instance could not forward attributes to the element containing it.
//
// Depends on Checks: None
//
// Sets Fields:
//   - Components.Blocks.Instances.CannotForwardAttributes
//
// Depends on Fields: None
func (z *analyzer) AnalyzeBlockInstance_CannotForwardAttributes(
	bi *file.BlockInstance, cannotAttributes file.AnalysisWithReason[ast.AttributeInhibitor],
) {
	bi.CannotForwardAttributes = cannotAttributes
}
