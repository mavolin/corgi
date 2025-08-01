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

	// CouldAcceptAttributes indicates whether the component could accept
	// attributes passed to it.
	CouldAcceptAttributes Analysis[bool]
	// CouldForwardAttributes indicates whether the component could forward
	// attributes it receives to the element containing a component call
	// to it.
	//
	// CouldForwardAttributes implies CouldAcceptAttributes.
	CouldForwardAttributes Analysis[bool]

	// FirstPermanentAndPlaceholder is the first &-placeholder that is not
	// part of a block default.
	FirstPermanentAndPlaceholder Analysis[*ast.AndPlaceholder]
	// FirstPermanentTopLevelAndPlaceholder is the first &-placeholder that is
	// at the top-level of the component, i.e. not nested inside an element or
	// part of a block default.
	FirstPermanentTopLevelAndPlaceholder Analysis[*ast.AndPlaceholder]

	// FirstPermanentTopLevelAttributeWriter is the first attribute writer that is
	// not part of a block default.
	FirstPermanentTopLevelAttributeWriter Analysis[ast.AttributeWriter]
	// FirstPermanentContentWriter is the first content writer that is not part
	// of a block default.
	FirstPermanentContentWriter Analysis[ast.ContentWriter]
	// FirstPermanentElementWriter is the first element writer that is not part
	// of a block default.
	FirstPermanentElementWriter Analysis[ast.ElementWriter]
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
func (c *Component) FirstIncludedAndPlaceholder(cc *ComponentCall) Analysis[*ast.AndPlaceholder] {
	ap := c.FirstPermanentTopLevelAndPlaceholder
	if ap.NotZero() {
		return ap
	}

	failed := ap.Failed
	for _, block := range c.Blocks {
		instance := block.FirstIncludedAndPlaceholder(cc)
		if instance.Failed {
			failed = true
		} else if instance.Result != nil {
			return instance.Result.Default.FirstAndPlaceholder
		}
	}

	return ResultIf[*ast.AndPlaceholder](nil, !failed)
}

type ComponentParameter struct {
	//
	// BUILD SYMBOLS

	AST *ast.ComponentParameter

	//
	// ANALYZER

	// The InferredType of this value, if there is no explicit type or if using
	// a special type, like an attribute type.
	InferredType  Analysis[string]
	AttributeType Analysis[attrtype.Type] // if type is a safe.*
	AttributeName Analysis[string]        // if type is safe.Unsafe*
}

func (p *ComponentParameter) ResolvedType() Analysis[string] {
	if p.AST.Type != nil {
		return Result(p.AST.Type.Type)
	}
	return p.InferredType
}

func (p *ComponentParameter) Required() bool {
	return p.AST.Colon != nil || p.AST.Default != nil
}
