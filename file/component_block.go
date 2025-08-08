package file

import "github.com/mavolin/corgi/v2/file/ast"

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

		// NotForwarded indicates that this block instance is not forwarded,
		// i.e. it is placed inside an element.
		//
		// The reason is the element writer containing this block instance.
		// If this block instance is placed inside a block setter, and that
		// block is not forwarded, the reason will be set to the component call.
		// Component call reasons are preferred over element reasons.
		NotForwarded AnalysisWithReason[ast.ElementWriter]
		// CannotForwardAttributes indicates that this block instance can't
		// forward attributes to the element containing it.
		//
		// NotForwarded might be false, but CannotForwardAttributes is true:
		// Consider the following example:
		// 	comp Woof() {
		//		br
		//      block
		//  }
		//
		// In the above example, the block is clearly forwarded, but it is
		// placed after the br element, which means it cannot forward
		// attributes.
		CannotForwardAttributes AnalysisWithReason[ast.ContentWriter]
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

func (bi *BlockInstance) Forwarded() (a Analysis[bool]) {
	if bi.NotForwarded.Failed() {
		a.SetFailed()
	} else {
		a.SetResult(bi.NotForwarded.False())
	}
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
	if bi.Parent != nil {
		return bi.Parent.DefaultOverwritten(cc)
	}
	return false
}
