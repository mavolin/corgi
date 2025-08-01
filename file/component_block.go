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
	// TopLevel indicates at least one instance of this block is placed
	// outside any element.
	TopLevel Analysis[bool]
	// ForwardsAttributes indicates that all instances of this block can write
	// to the list of attributes of their containing element.
	ForwardsAttributes Analysis[bool]
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

		// TopLevel indicates whether this block instance is placed outside
		// any element.
		TopLevel Analysis[bool]
		// ForwardsAttributes indicates whether this block instance can write
		// to the list of attributes of it's containing element.
		//
		// TopLevel implies ForwardsAttributes.
		ForwardsAttributes Analysis[bool]
	}

	BlockInstanceDefault struct {
		//
		// BUILD SYMBOLS

		AST ast.Body

		//
		// ANALYZER

		FirstAndPlaceholderWriter          Analysis[ast.AndPlaceholderWriter]
		FirstForwardedAndPlaceholderWriter Analysis[ast.AndPlaceholderWriter]

		FirstTopLevelAttributeWriter Analysis[ast.AttributeWriter]
		FirstContentWriter           Analysis[ast.ContentWriter]
		FirstElementWriter           Analysis[ast.ElementWriter]
	}
)

// DefaultOverwritten indicates whether the default of this block instance
// is overwritten in the given component call.
// This is the case if the component call sets this block or one of this
// block's parent blocks.
func (cbi *BlockInstance) DefaultOverwritten(cc *ComponentCall) bool {
	if cc.BlockSetterByName(cbi.Group.Name) != nil {
		return true
	}
	if cbi.Parent != nil {
		return cbi.Parent.DefaultOverwritten(cc)
	}
	return false
}
