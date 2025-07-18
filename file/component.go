package file

import (
	"github.com/mavolin/corgi/v2/escape/attrtype"
	"github.com/mavolin/corgi/v2/file/ast"
)

// ============================================================================
// Component
// ======================================================================================

type Component struct {
	// BUILD SYMBOLS
	//

	// DefinedAST is the AST of a defined component, i.e. the AST of a
	// non-alias component.
	//
	// Either this or AliasAST is set, but not both.
	DefinedAST *ast.Component
	// AliasAST is the AST of the alias statement defining this Component.
	//
	// Either this or DefinedAST is set, but not both.
	AliasAST *ast.Alias

	// File is the file the Component is defined in.
	File *File

	// ComponentCalls are the component calls this component calls.
	ComponentCalls []*ComponentCall

	// ANALYZER
	//

	// AnalyzedWithErrors indicates whether this component could not be fully
	// analyzed without errors.
	// It might also be set by the linker, if there
	// is a circular alias, which needs to be checked in the linker, not the
	// analyzer to allow successful parameter linking.
	AnalyzedWithErrors bool

	// Parameters are the parameters of the Component.
	//
	// If the component is an alias, this list also includes the remaining
	// (unset by the aliased component call) parameters of the component
	// being aliased.
	Parameters []*ComponentParameter

	// Blocks are the blocks used in the Component in the order they
	// appear in.
	//
	// If the component is an alias, this list also includes the
	// remaining (unset by the aliased component call) blocks of the
	// component being aliased.
	Blocks []*Block

	// FirstPermanentAndPlaceholder is the first &-placeholder that is not
	// part of a block default.
	FirstPermanentAndPlaceholder *ast.AndPlaceholder
	// FirstPermanentTopLevelAndPlaceholder is the first &-placeholder that is
	// at the top-level of the component, i.e. not nested inside an element or
	// part of a block default.
	FirstPermanentTopLevelAndPlaceholder *ast.AndPlaceholder
}

func (c *Component) Header() *ast.ComponentHeader {
	if c.DefinedAST != nil {
		return c.DefinedAST.Header
	}
	return c.AliasAST.Header
}

func (c *Component) Start() ast.Position {
	if c.DefinedAST != nil {
		return c.DefinedAST.Start()
	}
	return c.AliasAST.Start()
}

func (c *Component) End() ast.Position {
	if c.DefinedAST != nil {
		return c.DefinedAST.End()
	}
	return c.AliasAST.End()
}

func (c *Component) QualifiedName() string {
	return c.File.Package.Name + "." + c.Header().Name.Ident
}

func (c *Component) ParameterByName(name string) *ComponentParameter {
	for _, p := range c.Parameters {
		if p.AST.Name.Ident == name {
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
	block := c.BlockByName(b.Name.Ident)
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
	block := c.BlockByName(b.Name.Ident)
	if block == nil {
		return nil
	}
	return block.InstanceByNode(b)
}

func (c *Component) Exported() bool {
	return IsExported(c.Header().Name.Ident)
}

// HasAndPlaceholder returns whether the component has an &-placeholder that is
// included in the output of the component for the given component call.
//
// Passing nil checks the general case, in which all block defaults are
// included.
func (c *Component) HasAndPlaceholder(cc *ComponentCall) bool {
	return c.FirstIncludedAndPlaceholder(cc) != nil
}

// FirstIncludedAndPlaceholder returns the first &-placeholder that is included in the
// output of the component for the given component call.
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
// &-placeholder that is included in the output of the component for the given
// component call.
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
	// ANALYZER
	//

	AST *ast.ComponentParameter

	// Component is the component this parameter belongs to.
	//
	// This might be different from the component containing this parameter, if
	// that component is an alias and this parameter belongs to the component
	// being aliased.
	Component *Component

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
	// ANALYZER
	//

	// Component is the component this block belongs to.
	//
	// This might be different from the component containing this block, if
	// that component is an alias and this block belongs to the component
	// being aliased.
	Component *Component

	// Name is the name of the block.
	Name string

	Instances []*BlockInstance

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
		Group *Block
		AST   *ast.Block
		// ChildOf is the instance of another block that contains this block.
		ChildOf *BlockInstance

		// TopLevel indicates whether this block instance is placed outside
		// any element.
		TopLevel bool

		Default BlockInstanceDefault // nil if no default
	}

	BlockInstanceDefault struct {
		AST ast.Body

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
	// BUILD SYMBOLS
	//

	AST *ast.ComponentCall

	// AliasFor is the component that aliases this call.
	//
	// If true, be mindful that not all required parameters/blocks might be set.
	AliasFor *Component

	// File is the file the Component is defined in.
	File *File

	// LINKER
	//

	// Component is the Component being called.
	Component *Component

	// ANALYZER
	//

	AnalyzedWithErrors bool

	// Withs are the withs used in this component call.
	//
	// All withs and their instances are guaranteed to be correctly set after
	// analyzing, even if [AnalyzedWithErrors] is true.
	Withs []*With
	// FirstPlaceholderAnd is the first & that fills the components placeholder.
	// Nil, if no such & exists.
	FirstPlaceholderAnd *ast.And
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
	w := cc.WithByName(n.Name.Ident)
	if w == nil {
		return nil
	}

	if w.InstanceByNode(n) != nil {
		return w
	}
	return nil
}

type With struct {
	Name  string
	Block *Block

	Instances []*WithInstance
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
