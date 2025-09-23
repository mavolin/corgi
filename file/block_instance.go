package file

import (
	"github.com/mavolin/corgi/v2/escape/elemtype"
	"github.com/mavolin/corgi/v2/file/ast"
)

type (
	BlockInstance struct {
		//
		// BUILD SYMBOLS

		Group *Block
		AST   *ast.Block
		// ContainingInstance is the instance of another block that contains this block.
		ContainingInstance *BlockInstance

		Default *BlockInstanceDefault // nil if no default

		//
		// ANALYZER

		// Forwarded indicates that this block instance is forwarded somehow,
		// i.e. it is at the top-level of its component.
		Forwarded Analysis[bool]
		// ContainingElements are all elements containing this block instance.
		// If Forwarded is true, the list is not absolute: It would need to be
		// extended with the containing elements of the call to the component
		// containing this node.
		ContainingElements Analysis[SliceRef[ast.ContainingElement]]
		// ContainingElementSpecs are the unique specs of all containing
		// elements, including those containing the block instance indirectly.
		ContainingElementSpecs Analysis[SliceRef[*ElementSpec]]

		// CannotForwardAttributes indicates that this block instance can't
		// forward attributes to the element containing it.
		//
		// Forwarded might be true, but CannotForwardAttributes is also true:
		// Consider the following example:
		// 	comp Woof() {
		//	    br
		//      block
		//  }
		//
		// In the above example, the block is clearly forwarded, but it is
		// placed after the br element, which means it cannot forward
		// attributes.
		//
		// The reason is the first attribute inhibitor preventing the
		// forwarding of attributes.
		// If this block instance is at the top-level of a block setter that
		// cannot forward attributes, the reason is a
		// [ast.BlockSetterAttributeInhibitor] with the block setter field set
		// to that block setter.
		CannotForwardAttributes AnalysisWithReason[ast.AttributeInhibitor]
	}

	BlockInstanceDefault struct {
		//
		// BUILD SYMBOLS

		AST ast.Body

		//
		// ANALYZER

		// AcceptsAttributes indicates that this block instance default
		// writes the &-placeholder.
		//
		// If this is set to a component call, the component call's
		// [ReceivesAndPlaceholder] tells you where how sources the
		// &-placeholder.
		//
		// The reason is the first &-placeholder writer that writes the
		// &-placeholder.
		AcceptsAttributes AnalysisWithReason[ast.AndPlaceholderWriter]
		// ForwardsReceivedAttributes indicates that this block instance default
		// forwards the &-placeholder to the element containing the block
		// instance.
		//
		// The reason is the first &-placeholder writer that forwards the
		// &-placeholder.
		//
		// If this is set to a component call, the component call's
		// [ReceivesAndPlaceholder] tells you where how sources the
		// &-placeholder.
		//
		// ForwardsReceivedAttributes implies AcceptsAttributes.
		ForwardsReceivedAttributes AnalysisWithReason[ast.AndPlaceholderWriter]

		// ElementsWithAndPlaceholder are all elements in the default's body
		// that contain an &-placeholder.
		ElementsWithAndPlaceholder Analysis[SliceRef[ast.AttributeReceiver]]
		// ElementSpecsWithAndPlaceholder are the unique specs of all elements
		// containing an &-placeholder, including those containing the elements
		// indirectly.
		ElementSpecsWithAndPlaceholder Analysis[SliceRef[*ElementSpec]]

		ForwardsAttributes AnalysisWithReason[ast.AttributeWriter]
		WritesContent      AnalysisWithReason[ast.ContentWriter]
		WritesElements     AnalysisWithReason[ast.ElementWriter]
	}
)

// ElementType is the element type that this block assumes.
//
// If the block instance is fully forwarded, i.e. has no containing
// elements, ElementType is set to Normal.
//
// For all other element types it is the minimum of all containing
// elements.
//
// JS and CSS are not permitted.
//
// Void and Nothing are equivalent in this context, indicating the
// block only accepts attributes.
// For simplicity, Nothing is always used.
func (bi *BlockInstance) ElementType() (a Analysis[elemtype.Type]) {
	if bi.ContainingElementSpecs.Failed() {
		a.SetFailed()
		return a
	}

	if bi.ContainingElementSpecs.Result().Len() == 0 {
		a.SetResult(elemtype.Unknown)
		return a
	}

	t := elemtype.Normal
	for _, spec := range bi.ContainingElementSpecs.Result().Get() {
		if spec.Type.Failed() {
			a.SetFailed()
			return a
		}

		specType := spec.Type.Result()
		switch specType {
		case elemtype.JS, elemtype.CSS:
			a.SetFailed()
			return a
		case elemtype.Void:
			specType = elemtype.Nothing
		case elemtype.Unknown, elemtype.Nothing, elemtype.Normal, elemtype.Text:
		}
		t = min(t, specType)
	}
	a.SetResult(t)
	return a
}

// MostRestrictiveElement is one (of the possibly multiple) element with the
// most restrictive (smallest) element type.
//
// It follows the same rules as [ElementType].
func (bi *BlockInstance) MostRestrictiveElement() (a AnalysisWithReason[ast.ContainingElement]) {
	if bi.ContainingElements.Failed() {
		a.SetFailed()
		return a
	}

	f := bi.Group.Component.File

	var resultCEl ast.ContainingElement
	var resultTyp elemtype.Type
	for _, cEl := range bi.ContainingElements.Result().Get() {
		switch cEl := cEl.(type) {
		case *ast.BlockSetterContainingElement:
			cc := f.ComponentCallByNode(cEl.ComponentCall)
			s := cc.BlockSetterByNode(cEl.BlockSetter)
			if s == nil || s.Block == nil {
				a.SetFailed()
				return a
			}

			for _, instance := range s.Block.Instances {
				if instance.ContainingElementSpecs.Failed() {
					a.SetFailed()
					return a
				}
				for _, spec := range instance.ContainingElementSpecs.Result().Get() {
					if spec.Type.Failed() {
						a.SetFailed()
						return a
					}

					if resultCEl == nil || spec.Type.Result() < resultTyp {
						resultCEl = cEl
						resultTyp = spec.Type.Result()
					}
				}
			}
		case *ast.Element:
			ref := f.ElementReferenceByNode(cEl.Header.Name)
			if ref.Spec == nil {
				a.SetFailed()
				return a
			}

			// Always prefer elements to block setters, so use <=
			if resultCEl == nil || ref.Spec.Type.Result() <= resultTyp {
				resultCEl = cEl
				resultTyp = ref.Spec.Type.Result()
			}
		default:
			panic("unknown containing element type")
		}

		switch resultTyp {
		case elemtype.Unknown:
			a.SetFailed()
			return a
		case elemtype.JS, elemtype.CSS:
			a.SetFailed()
			return a
		case elemtype.Void:
			resultTyp = elemtype.Nothing
		case elemtype.Nothing, elemtype.Normal, elemtype.Text:
		}
	}

	a.SetReason(resultCEl)
	return a
}

func (bi *BlockInstance) ForwardsAttributes() (a Analysis[bool]) {
	if bi.CannotForwardAttributes.Failed() {
		a.SetFailed()
	} else {
		a.SetResult(bi.CannotForwardAttributes.False())
	}
	return a
}

// DefaultOverwritten indicates whether the default of this block instance
// is overwritten in the given component call.
// This is the case if the component call sets this block or one of this
// block's parent blocks.
func (bi *BlockInstance) DefaultOverwritten(cc *ComponentCall) bool {
	if cc.BlockSetterByName(bi.Group.Name) != nil {
		return true
	}
	if bi.ContainingInstance != nil {
		return bi.ContainingInstance.DefaultOverwritten(cc)
	}
	return false
}
