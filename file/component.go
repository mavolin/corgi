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

	// AnalyzedWithErrors indicates whether this component could not be fully
	// analyzed without errors.
	// It might also be set by the linker, if there
	// is a circular alias, which needs to be checked in the linker, not the
	// analyzer to allow successful parameter linking.
	AnalyzedWithErrors bool
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
	if cc.WithByName(cbi.Group.Name) != nil {
		return true
	}
	if cbi.ChildOf != nil {
		return cbi.ChildOf.DefaultOverwritten(cc)
	}
	return false
}

// ============================================================================
// Component Call
// ======================================================================================

type ComponentCall struct {
	//
	// BUILD SYMBOLS

	AST *ast.ComponentCall

	// File is the file the Component is defined in.
	File *File

	// Withs are the withs used in this component call.
	//
	// All withs and their instances are guaranteed to be correctly set after
	// analyzing, even if [AnalyzedWithErrors] is true.
	Withs []*With

	//
	// LINKER

	// Component is the Component being called.
	Component *Component

	//
	// ANALYZER

	AnalyzedWithErrors bool
}

func (cc *ComponentCall) External() bool {
	return cc.File.Package != cc.Component.File.Package
}

func (cc *ComponentCall) Local() bool {
	return !cc.External()
}

func (cc *ComponentCall) WithByName(name string) *With {
	for _, with := range cc.Withs {
		if with.Name == name {
			return with
		}
	}
	return nil
}

func (cc *ComponentCall) WithByNode(n *ast.With) *With {
	w := cc.WithByName(n.Name())
	if w == nil {
		return nil
	}

	if w.InstanceByNode(n) != nil {
		return w
	}
	return nil
}

type With struct {
	//
	// BUILD SYMBOLS

	Name      string
	Instances []*WithInstance

	//
	// ANALYZER

	Block *Block
}

func (w *With) InstanceByNode(n *ast.With) *WithInstance {
	for _, instance := range w.Instances {
		if instance.AST == n {
			return instance
		}
	}
	return nil
}

type WithInstance struct {
	Group *With
	AST   *ast.With
}
