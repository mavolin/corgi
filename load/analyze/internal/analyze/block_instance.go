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

type blockInstance_Forwarded struct{}

func (z *analyzer) AnalyzeBlockInstance_Forwarded(ctx context.Context, parents []*walk.Context, bi *file.BlockInstance) {
	defer z.Ran(bi, blockInstance_Forwarded{})

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
				} else if z.block_Forwarded(s.Block).False() {
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

type blockInstance_ContainingElements struct{}

func (z *analyzer) AnalyzeBlockInstance_ContainingElements(ctx context.Context, parents []*walk.Context, bi *file.BlockInstance) {
	defer z.Ran(bi, blockInstance_ContainingElements{})

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
				for _, bi2 := range s.Block.Instances {
					if z.blockInstance_ContainingElements(bi2).Failed() {
						bi.ContainingElements.SetFailed()
						return true
					}
				}

				containingElements = append(containingElements, &ast.BlockSetterContainingElement{
					ComponentCall: ccAST,
					BlockSetter:   parent,
				})
				if z.block_Forwarded(s.Block).False() {
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

type blockInstance_ContainingElementSpecs struct{}

func (z *analyzer) AnalyzeBlockInstance_ContainingElementSpecs(bi *file.BlockInstance) {
	defer z.Ran(bi, blockInstance_ContainingElementSpecs{})

	containingElements := z.blockInstance_ContainingElements(bi)
	if containingElements.Failed() {
		bi.ContainingElementSpecs.SetFailed()
		return
	}

	if containingElements.Result().Len() == 0 {
		bi.ContainingElementSpecs.SetResult(file.NilSliceRef[*file.ElementSpec]())
		return
	}

	f := bi.Group.Component.File

	specSet := make(map[*file.ElementSpec]struct{}, containingElements.Result().Len())
	for _, e := range containingElements.Result().Get() {
		switches.ContainingElement(e,
			func(e *ast.BlockSetterContainingElement) {
				cc := f.ComponentCallByNode(e.ComponentCall)
				s := cc.BlockSetterByNode(e.BlockSetter)
				if s == nil || s.Block == nil {
					bi.ContainingElementSpecs.SetFailed()
					return
				}

				for _, bi2 := range s.Block.Instances {
					containingElementSpecs := z.blockInstance_ContainingElementSpecs(bi2)
					if containingElementSpecs.Failed() {
						bi.ContainingElementSpecs.SetFailed()
						return
					}

					for _, spec := range containingElementSpecs.Result().Get() {
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

type blockInstance_CannotForwardAttributes struct{}

func (z *analyzer) AnalyzeBlockInstance_CannotForwardAttributes(
	bi *file.BlockInstance, cannotAttributes file.AnalysisWithReason[ast.AttributeInhibitor],
) {
	defer z.Ran(bi, blockInstance_CannotForwardAttributes{})
	bi.CannotForwardAttributes = cannotAttributes
}
