package file

import (
	"github.com/mavolin/corgi/v2/escape/attrtype"
	"github.com/mavolin/corgi/v2/file/ast"
)

// ============================================================================
// Component
// ======================================================================================

type Component struct {
	//
	// BUILD SYMBOLS

	AST *ast.Component

	// File is the file the Component is defined in.
	File *File

	// ComponentCalls are the component calls this component calls.
	ComponentCalls []*ComponentCall

	// Parameters are the parameters of the Component.
	Parameters []*ComponentParameter

	// Blocks are the blocks used in the Component in the order they
	// appear in.
	Blocks []*Block

	//
	// ANALYZER

	// Analyzed indicates whether the Component has been analyzed,
	// albeit with errors.
	Analyzed bool

	// FirstPermanentAndPlaceholder is the first &-placeholder that is not
	// part of a block default.
	FirstPermanentAndPlaceholder *ast.AndPlaceholder
	// FirstPermanentTopLevelAndPlaceholder is the first &-placeholder that is
	// at the top-level of the component, i.e. not nested inside an element or
	// part of a block default.
	FirstPermanentTopLevelAndPlaceholder *ast.AndPlaceholder
}

func (c *Component) ParameterByName(name string) *ComponentParameter {
	for _, p := range c.Parameters {
		if p.AST.Name.Name == name {
			return p
		}
	}
	return nil
}

func (c *Component) ParameterByNode(p *ast.ComponentParameter) *ComponentParameter {
	for _, param := range c.Parameters {
		if param.AST == p {
			return param
		}
	}
	return nil
}

func (c *Component) BlockByName(name string) *Block {
	for _, block := range c.Blocks {
		if block.Name == name {
			return block
		}
	}
	return nil
}

func (c *Component) BlockByNode(b *ast.Block) *Block {
	block := c.BlockByName(b.Name())
	if block == nil {
		return nil
	}

	for _, instance := range block.Instances {
		if instance.AST == b {
			return block
		}
	}
	return nil
}

func (c *Component) BlockInstanceByNode(b *ast.Block) *BlockInstance {
	block := c.BlockByName(b.Name())
	if block == nil {
		return nil
	}
	return block.InstanceByNode(b)
}

func (c *Component) Exported() bool {
	return IsExported(c.AST.Header.Name.Name)
}

// FirstIncludedAndPlaceholder returns the first &-placeholder that is included
// in the output of the component for the given component call.
//
// Passing nil checks the general case, in which all block defaults are
// included.
func (c *Component) FirstIncludedAndPlaceholder(cc *ComponentCall) *ast.AndPlaceholder {
	if c.FirstPermanentAndPlaceholder != nil {
		return c.FirstPermanentAndPlaceholder
	}

	instance := c.FirstBlockIncludedAndPlaceholder(cc)
	if instance != nil {
		return instance.Default.FirstAndPlaceholder
	}
	return nil
}

// FirstBlockIncludedAndPlaceholder returns the first block instance with an
// &-placeholder in its default that is included in the output of the component
// for the given component call.
//
// Passing nil checks the general case, in which all block defaults are
// included.
func (c *Component) FirstBlockIncludedAndPlaceholder(cc *ComponentCall) *BlockInstance {
	for _, block := range c.Blocks {
		if instance := block.FirstIncludedAndPlaceholder(cc); instance != nil {
			return instance
		}
	}
	return nil
}

type ComponentParameter struct {
	//
	// BUILD SYMBOLS

	AST *ast.ComponentParameter

	//
	// ANALYZER

	// The InferredType of this value, if there is no explicit type or if using
	// a special type, like an attribute type.
	InferredType  string
	AttributeType attrtype.Type // if type is a safe.*
	AttributeName string        // if type is safe.Unsafe*
}

func (p *ComponentParameter) ResolvedType() string {
	if p.AST.Type != nil {
		return p.AST.Type.Type
	}
	return p.InferredType
}

func (p *ComponentParameter) Required() bool {
	return p.AST.Colon != nil || p.AST.Default != nil
}

// Block provides information about a block used in a
// Component.
//
// If a block B (with or without default) is nested inside another block A's
// default, we automatically, for the sake of simplicity, set
// DefaultWritesBody, DefaultWritesElements, and
// DefaultWritesTopLevelAttributes of block A to true.
// The DefaultTopLevelAndPlaceholder, if not true regardless, will be set
// to true, if that exact placement of block B is top-level and has a
// top-level and placeholder.
type Block struct {
	//
	// BUILD SYMBOLS

	// Name is the name of the block.
	Name string

	Instances []*BlockInstance

	//
	// ANALYZER

	Required bool
}

func (b *Block) InstanceByNode(n *ast.Block) *BlockInstance {
	for _, instance := range b.Instances {
		if instance.AST == n {
			return instance
		}
	}
	return nil
}

// FirstIncludedAndPlaceholder returns the first instance of an &-placeholder
// in a block default that is included in the output of the component for the
// given component call.
//
// Passing nil checks the general case, in which all block defaults are
// included.
func (b *Block) FirstIncludedAndPlaceholder(cc *ComponentCall) *BlockInstance {
	for _, instance := range b.Instances {
		if instance.Default.FirstAndPlaceholder != nil && (cc == nil || !instance.DefaultOverwritten(cc)) {
			return instance
		}
	}

	return nil
}

// TopLevel reports whether this block is top-level.
func (b *Block) TopLevel(s AnalysisStrategy) bool {
	s.assertValid()

	for _, instance := range b.Instances {
		if instance.TopLevel && s == AtLeastOne {
			return true
		} else if !instance.TopLevel && s == All {
			return false
		}
	}

	return s == All
}

type (
	BlockInstance struct {
		//
		// BUILD SYMBOLS

		Group *Block
		AST   *ast.Block
		// ChildOf is the instance of another block that contains this block.
		ChildOf *BlockInstance

		Default *BlockInstanceDefault // nil if no default

		//
		// ANALYZER

		// TopLevel indicates whether this block instance is placed outside
		// any element.
		TopLevel bool
	}

	BlockInstanceDefault struct {
		//
		// BUILD SYMBOLS

		AST ast.Body

		//
		// ANALYZER

		FirstAndPlaceholder *ast.AndPlaceholder // nil if no &-placeholder
	}
)

func (cbi *BlockInstance) Required() bool {
	return cbi.AST.Default == nil
}

// DefaultOverwritten indicates whether the default of this block instance
// is overwritten in the given component call.
// This is the case if the component call sets this block or one of this
// block's parent blocks.
func (cbi *BlockInstance) DefaultOverwritten(cc *ComponentCall) bool {
	if cc.BlockSetterByName(cbi.Group.Name) != nil {
		return true
	}
	if cbi.ChildOf != nil {
		return cbi.ChildOf.DefaultOverwritten(cc)
	}
	return false
}
