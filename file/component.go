package file

import (
	"github.com/mavolin/corgi/v2/escape/attrtype"
	"github.com/mavolin/corgi/v2/file/ast"
)

type Component struct {
	//
	// BUILD SYMBOLS

	AST *ast.Component

	// File is the file the Component is defined in.
	File *File

	Name Identifier

	// Parameters are the parameters of the Component.
	//
	// Parameters that are nil in the AST's parameters list are omitted.
	// Therefore, there might be gaps in the indexes and fewer parameters than
	// in the AST.
	// Use ParameterByNode to get the parameter by its AST node, not it's
	// index.
	Parameters []*ComponentParameter

	// Blocks are the blocks used in the Component in the order they
	// appear in.
	Blocks []*Block

	//
	// ANALYZER

	// Analyzed indicates whether the Component has been analyzed,
	// albeit with errors.
	Analyzed bool

	// Circular indicates whether the Component is part of a cycle of
	// component calls.
	Circular bool

	// AlwaysAcceptsAttributes indicates whether the component has a
	// permanent &-placeholder writer, i.e. an &-placeholder writer that is not
	// part of a block default.
	//
	// The reason is that &-placeholder writer.
	AlwaysAcceptsAttributes AnalysisWithReason[ast.AndPlaceholderWriter]
	// AlwaysForwardsReceivedAttributes indicates whether the component has
	// a permanent &-placeholder writer that is forwarded and not part of a
	// block default.
	//
	// The reason is that &-placeholder writer.
	AlwaysForwardsReceivedAttributes AnalysisWithReason[ast.AndPlaceholderWriter]

	// AlwaysForwardsAttributes indicates whether the component has a
	// permanent attribute writer that is forwarded and not part of a block
	// default.
	//
	// The reason is that attribute writer.
	AlwaysForwardsAttributes AnalysisWithReason[ast.AttributeWriter]
	// AlwaysWritesContent indicates whether the component has a
	// permanent content writer that is not part of a block default.
	AlwaysWritesContent AnalysisWithReason[ast.ContentWriter]
	// AlwaysWritesElements indicates whether the component has a
	// permanent element writer that is not part of a block default.
	//
	// The reason is that element writer.
	AlwaysWritesElements AnalysisWithReason[ast.ElementWriter]

	// PermanentElementsWithAndPlaceholder are all elements outside any
	// block defaults that contain an &-placeholder.
	PermanentElementsWithAndPlaceholder Analysis[SliceRef[ast.AttributeReceiver]]
	// PermanentElementSpecsWithAndPlaceholder are the unique specs of all
	// elements outside any block defaults that contain an &-placeholder,
	// including those containing the elements indirectly.
	PermanentElementSpecsWithAndPlaceholder Analysis[SliceRef[*ElementSpec]]
}

func (c *Component) ParameterByName(name Identifier) *ComponentParameter {
	for _, p := range c.Parameters {
		if p.Name == name {
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

func (c *Component) BlockByName(name Identifier) *Block {
	for _, block := range c.Blocks {
		if block.Name == name {
			return block
		}
	}
	return nil
}

func (c *Component) BlockByNode(b *ast.Block) *Block {
	block := c.BlockByName(Identifier(b.Name()))
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
	block := c.BlockByName(Identifier(b.Name()))
	if block == nil {
		return nil
	}
	return block.InstanceByNode(b)
}

func (c *Component) Exported() bool {
	return c.Name.Exported()
}

// CouldAcceptAttributes indicates whether the component could accept
// attributes passed to it.
func (c *Component) CouldAcceptAttributes() (a AnalysisWithReason[ast.AndPlaceholderWriter]) {
	if c.AlwaysAcceptsAttributes.True() {
		a.SetReason(c.AlwaysAcceptsAttributes.Reason())
		return a
	}

	a.SetFalse()
	if c.AlwaysAcceptsAttributes.Failed() {
		a.SetFailed()
	}

	for _, block := range c.Blocks {
		for _, instance := range block.Instances {
			if instance.Default == nil {
				continue
			}
			if instance.Default.AcceptsAttributes.True() {
				a.SetReason(instance.Default.AcceptsAttributes.Reason())
				return a
			} else if instance.Default.AcceptsAttributes.Failed() {
				a.SetFailed()
			}
		}
	}

	return a
}

// CouldForwardReceivedAttributes indicates whether the component could forward
// attributes it receives to the element containing a component call
// to it.
//
// CouldForwardReceivedAttributes implies CouldAcceptAttributes.
func (c *Component) CouldForwardReceivedAttributes() (a AnalysisWithReason[ast.AndPlaceholderWriter]) {
	if c.AlwaysForwardsReceivedAttributes.True() {
		a.SetReason(c.AlwaysForwardsReceivedAttributes.Reason())
		return a
	}

	a.SetFalse()
	if c.AlwaysForwardsReceivedAttributes.Failed() {
		a.SetFailed()
	}

	for _, block := range c.Blocks {
		for _, instance := range block.Instances {
			if instance.Default == nil {
				continue
			}
			forwardsAndPlaceholder := ConditionalAnalysis(instance.Forwarded, instance.Default.ForwardsReceivedAttributes)
			if forwardsAndPlaceholder.True() {
				a.SetReason(forwardsAndPlaceholder.Reason())
				return a
			} else if forwardsAndPlaceholder.Failed() {
				a.SetFailed()
			}
		}
	}

	return a
}

// ============================================================================
// Parameter
// ======================================================================================

type ComponentParameter struct {
	//
	// BUILD SYMBOLS

	AST       *ast.ComponentParameter
	Name      Identifier
	Component *Component

	//
	// ANALYZER

	// The InferredType of this value, if there is no explicit type or if using
	// a special type, like an attribute type.
	InferredType  Analysis[Type]
	AttributeType Analysis[attrtype.Type] // if type is a safe.*
	AttributeName Analysis[string]        // if type is safe.Unsafe*
}

func (p *ComponentParameter) ResolvedType() Analysis[Type] {
	if p.AST.Type != nil {
		var a Analysis[Type]
		a.SetResult(Type(p.AST.Type.Type))
		return a
	}
	return p.InferredType
}

func (p *ComponentParameter) Required() bool {
	return p.AST.Colon != nil || p.AST.Default != nil
}
