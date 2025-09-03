package file

import (
	"github.com/mavolin/corgi/v2/escape/elemtype"
	"github.com/mavolin/corgi/v2/file/ast"
)

// Block provides information about a block used in a
// Component.
type Block struct {
	//
	// BUILD SYMBOLS

	// Name is the name of the block.
	Name string

	Instances []*BlockInstance

	//
	// ANALYZER

	Required Analysis[bool]

	// Forwarded indicates at least one instance of this block is placed
	// outside any element.
	//
	// The reason is that block instance.
	Forwarded AnalysisWithReason[*BlockInstance]
	// ContainingElements are all elements containing this block.
	// If Forwarded is true, the list is not absolute: It would need to be
	// extended with the containing elements of the call to the component
	// containing this block.
	//
	// The pointer to the slice has no significance and is just there to
	// satisfy the comparable constraint of Analysis.
	// It is never nil.
	ContainingElements Analysis[*[]ast.ContainingElement]
	// ContainingElementSpecs are the unique specs of all containing
	// elements, including those containing the block indirectly.
	//
	// The pointer to the slice has no significance and is just there to
	// satisfy the comparable constraint of Analysis.
	// It is never nil.
	ContainingElementSpecs Analysis[*[]*ElementSpec]

	// ElementType is the minimum element type of all containing elements.
	//
	// A type of Unknown indicates the block is fully forwarded and the
	// element type as such depends on the element containing the component
	// call.
	ElementType Analysis[elemtype.Type]

	// CannotForwardAttributes indicates that at least one instance of this
	// block cannot forward attributes.
	//
	// The reason is the first instance that cannot forward attributes.
	CannotForwardAttributes AnalysisWithReason[*BlockInstance]
}

func (b *Block) ForwardsAttributes() (a Analysis[bool]) {
	if b.CannotForwardAttributes.Failed() {
		a.SetFailed()
	} else {
		a.SetResult(b.CannotForwardAttributes.False())
	}
	return a
}

func (b *Block) InstanceByNode(n *ast.Block) *BlockInstance {
	for _, instance := range b.Instances {
		if instance.AST == n {
			return instance
		}
	}
	return nil
}

type (
	BlockInstance struct {
		//
		// BUILD SYMBOLS

		Group *Block
		AST   *ast.Block
		// Parent is the instance of another block that contains this block.
		Parent *BlockInstance

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
		//
		// The pointer to the slice has no significance and is just there to
		// satisfy the comparable constraint of Analysis.
		// It is never nil.
		ContainingElements Analysis[*[]ast.ContainingElement]
		// ContainingElementSpecs are the unique specs of all containing
		// elements, including those containing the block instance indirectly.
		//
		// The pointer to the slice has no significance and is just there to
		// satisfy the comparable constraint of Analysis.
		// It is never nil.
		ContainingElementSpecs Analysis[*[]*ElementSpec]

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
		ElementType Analysis[elemtype.Type]

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

		// WritesAndPlaceholder indicates that this block instance default
		// writes the &-placeholder.
		//
		// The reason is the first &-placeholder writer that writes the
		// &-placeholder.
		WritesAndPlaceholder AnalysisWithReason[ast.AndPlaceholderWriter]
		// ForwardsAndPlaceholder indicates that this block instance default
		// forwards the &-placeholder to the element containing the block
		// instance.
		//
		// The reason is the first &-placeholder writer that forwards the
		// &-placeholder.
		//
		// ForwardsAndPlaceholder implies WritesAndPlaceholder.
		ForwardsAndPlaceholder AnalysisWithReason[ast.AndPlaceholderWriter]

		ForwardsAttributes AnalysisWithReason[ast.AttributeWriter]
		WritesContent      AnalysisWithReason[ast.ContentWriter]
		WritesElements     AnalysisWithReason[ast.ElementWriter]
	}
)

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
	if bi.Parent != nil {
		return bi.Parent.DefaultOverwritten(cc)
	}
	return false
}
