package file

// ============================================================================
// Block
// ======================================================================================

type BlockType uint8

const (
	BlockTypeBlock BlockType = iota + 1
	BlockTypePrepend
	BlockTypeAppend
)

// Block represents a block with content.
// It is used for blocks from extendable templates as well as blocks in
// MixinCalls.
type Block struct {
	// Type is the type of block.
	Type BlockType

	// Name is the name of the block.
	Name Ident

	Body Scope

	Position
}

var _ ScopeItem = Block{}

func (Block) _typeScopeItem() {}

// ============================================================================
// Block Expansion
// ======================================================================================

type BlockExpansion struct {
	Item ScopeItem // Either Block, Element, ArrowBlock, or MixinCall
	Position
}

var _ ScopeItem = BlockExpansion{}

func (BlockExpansion) _typeScopeItem() {}
