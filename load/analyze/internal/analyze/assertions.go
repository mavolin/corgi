package analyze

import (
	"github.com/mavolin/corgi/v2/escape/attrtype"
	"github.com/mavolin/corgi/v2/escape/elemtype"
	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/switches"
	"github.com/mavolin/corgi/v2/internal/assert"
)

// =============================================================================
// Attribute
// ======================================================================================

func (z *analyzer) attribute_Value(a *file.Attribute) file.ResolvedValue {
	z.requireAttribute_Value(a)
	return a.Value
}

func (z *analyzer) requireAttribute_Value(a *file.Attribute) {
	if !a.Analyzed {
		z.Require(a, attribute_Value{})
	}
}

func (z *analyzer) attribute_Forwarded(a *file.Attribute) file.Analysis[bool] {
	z.requireAttribute_Forwarded(a)
	return a.Forwarded
}

func (z *analyzer) requireAttribute_Forwarded(a *file.Attribute) {
	if !a.Analyzed {
		z.Require(a, attribute_Forwarded{})
	}
}

func (z *analyzer) attribute_Receivers(a *file.Attribute) file.Analysis[file.SliceRef[ast.AttributeReceiver]] {
	z.requireAttribute_Receivers(a)
	return a.Receivers
}

func (z *analyzer) requireAttribute_Receivers(a *file.Attribute) {
	if !a.Analyzed {
		z.Require(a, attribute_Receivers{})
	}
}

func (z *analyzer) attribute_ReceivingElementSpecs(a *file.Attribute) file.Analysis[file.SliceRef[*file.ElementSpec]] {
	z.requireAttribute_ReceivingElementSpecs(a)
	return a.ReceivingElementSpecs
}

func (z *analyzer) requireAttribute_ReceivingElementSpecs(a *file.Attribute) {
	if !a.Analyzed {
		z.Require(a, attribute_ReceivingElementSpecs{})
	}
}

func (z *analyzer) attribute_Type(a *file.Attribute) file.Analysis[attrtype.Type] {
	z.requireAttribute_Type(a)
	return a.Type
}

func (z *analyzer) requireAttribute_Type(a *file.Attribute) {
	if !a.Analyzed {
		z.Require(a, attribute_Type{})
	}
}

// ============================================================================
// Block
// ======================================================================================

func (z *analyzer) block_Required(b *file.Block) file.Analysis[bool] {
	z.requireBlock_Required(b)
	return b.Required()
}

func (z *analyzer) requireBlock_Required(*file.Block) {
	// no deps
}

func (z *analyzer) block_Forwarded(b *file.Block) file.AnalysisWithReason[*file.BlockInstance] {
	z.requireBlock_Forwarded(b)
	return b.Forwarded()
}

func (z *analyzer) requireBlock_Forwarded(b *file.Block) {
	if !assert.DebugEnabled || b.Component.Analyzed || b.Component.Circular {
		return
	}
	for _, bi := range b.Instances {
		z.requireBlockInstance_Forwarded(bi)
	}
}

func (z *analyzer) block_ForwardsAttributes(b *file.Block) file.Analysis[bool] {
	z.requireBlock_ForwardsAttributes(b)
	return b.ForwardsAttributes()
}

func (z *analyzer) requireBlock_ForwardsAttributes(b *file.Block) {
	if !b.Component.Analyzed && !b.Component.Circular {
		z.requireBlock_CannotForwardAttributes(b)
	}
}

func (z *analyzer) block_CannotForwardAttributes(b *file.Block) file.AnalysisWithReason[*file.BlockInstance] {
	z.requireBlock_CannotForwardAttributes(b)
	return b.CannotForwardAttributes()
}

func (z *analyzer) requireBlock_CannotForwardAttributes(b *file.Block) {
	if !assert.DebugEnabled || b.Component.Analyzed || b.Component.Circular {
		return
	}
	for _, bi := range b.Instances {
		z.requireBlockInstance_CannotForwardAttributes(bi)
	}
}

func (z *analyzer) block_ElementType(b *file.Block) file.Analysis[elemtype.Type] {
	z.requireBlock_ElementType(b)
	return b.ElementType()
}

func (z *analyzer) requireBlock_ElementType(b *file.Block) {
	if !assert.DebugEnabled || b.Component.Analyzed || b.Component.Circular {
		return
	}
	for _, bi := range b.Instances {
		z.requireBlockInstance_ContainingElementSpecs(bi)
		if bi.ContainingElementSpecs.Successful() {
			for _, es := range bi.ContainingElementSpecs.Result().Get() {
				z.requireElementSpec_Type(es)
			}
		}
	}
}

func (z *analyzer) block_MostRestrictiveElement(b *file.Block) file.AnalysisWithReason[ast.ContainingElement] {
	z.requireBlock_MostRestrictiveElement(b)
	return b.MostRestrictiveElement()
}

func (z *analyzer) requireBlock_MostRestrictiveElement(b *file.Block) {
	if !assert.DebugEnabled || b.Component.Analyzed || b.Component.Circular {
		return
	}
	for _, bi := range b.Instances {
		z.requireBlockInstance_ContainingElements(bi)
		if bi.ContainingElements.Successful() {
			for _, e := range bi.ContainingElements.Result().Get() {
				switches.ContainingElement(e,
					func(e *ast.BlockSetterContainingElement) {
						cc := b.Component.File.ComponentCallByNode(e.ComponentCall)
						s := cc.BlockSetterByNode(e.BlockSetter)
						if s == nil || s.Block == nil {
							return
						}
						for _, bi := range s.Block.Instances {
							z.requireBlockInstance_ContainingElementSpecs(bi)
							if bi.ContainingElementSpecs.Successful() {
								for _, es := range bi.ContainingElementSpecs.Result().Get() {
									z.requireElementSpec_Type(es)
								}
							}
						}
					},
					func(e *ast.Element) {
						ref := b.Component.File.ElementReferenceByNode(e.Header.Name)
						if ref.Spec == nil {
							return
						}
						z.requireElementSpec_Type(ref.Spec)
					})
			}
		}
	}
}

// ============================================================================
// Block Instance
// ======================================================================================

func (z *analyzer) blockInstance_Forwarded(bi *file.BlockInstance) file.Analysis[bool] {
	z.requireBlockInstance_Forwarded(bi)
	return bi.Forwarded
}

func (z *analyzer) requireBlockInstance_Forwarded(bi *file.BlockInstance) {
	if !bi.Group.Component.Analyzed && !bi.Group.Component.Circular {
		z.Require(bi, blockInstance_Forwarded{})
	}
}

func (z *analyzer) blockInstance_ContainingElements(bi *file.BlockInstance) file.Analysis[file.SliceRef[ast.ContainingElement]] {
	z.requireBlockInstance_ContainingElements(bi)
	return bi.ContainingElements
}

func (z *analyzer) requireBlockInstance_ContainingElements(bi *file.BlockInstance) {
	if !bi.Group.Component.Analyzed && !bi.Group.Component.Circular {
		z.Require(bi, blockInstance_ContainingElements{})
	}
}

func (z *analyzer) blockInstance_ContainingElementSpecs(bi *file.BlockInstance) file.Analysis[file.SliceRef[*file.ElementSpec]] {
	z.requireBlockInstance_ContainingElementSpecs(bi)
	return bi.ContainingElementSpecs
}

func (z *analyzer) requireBlockInstance_ContainingElementSpecs(bi *file.BlockInstance) {
	if !bi.Group.Component.Analyzed && !bi.Group.Component.Circular {
		z.Require(bi, blockInstance_ContainingElementSpecs{})
	}
}

func (z *analyzer) blockInstance_CannotForwardAttributes(bi *file.BlockInstance) file.AnalysisWithReason[ast.AttributeInhibitor] {
	z.requireBlockInstance_CannotForwardAttributes(bi)
	return bi.CannotForwardAttributes
}

func (z *analyzer) requireBlockInstance_CannotForwardAttributes(bi *file.BlockInstance) {
	if !bi.Group.Component.Analyzed && !bi.Group.Component.Circular {
		z.Require(bi, blockInstance_CannotForwardAttributes{})
	}
}

func (z *analyzer) blockInstance_ElementType(bi *file.BlockInstance) file.Analysis[elemtype.Type] {
	z.requireBlockInstance_ElementType(bi)
	return bi.ElementType()
}

func (z *analyzer) requireBlockInstance_ElementType(bi *file.BlockInstance) {
	if !assert.DebugEnabled || bi.Group.Component.Analyzed || bi.Group.Component.Circular {
		return
	}
	z.requireBlockInstance_ContainingElementSpecs(bi)
	if bi.ContainingElementSpecs.Successful() {
		for _, es := range bi.ContainingElementSpecs.Result().Get() {
			z.requireElementSpec_Type(es)
		}
	}
}

func (z *analyzer) blockInstance_MostRestrictiveElement(bi *file.BlockInstance) file.AnalysisWithReason[ast.ContainingElement] {
	z.requireBlockInstance_MostRestrictiveElement(bi)
	return bi.MostRestrictiveElement()
}

func (z *analyzer) requireBlockInstance_MostRestrictiveElement(bi *file.BlockInstance) {
	if !assert.DebugEnabled || bi.Group.Component.Analyzed || bi.Group.Component.Circular {
		return
	}
	z.requireBlockInstance_ContainingElements(bi)
	if bi.ContainingElements.Successful() {
		for _, e := range bi.ContainingElements.Result().Get() {
			switches.ContainingElement(e,
				func(e *ast.BlockSetterContainingElement) {
					cc := bi.Group.Component.File.ComponentCallByNode(e.ComponentCall)
					s := cc.BlockSetterByNode(e.BlockSetter)
					if s == nil || s.Block == nil {
						return
					}
					for _, bi := range s.Block.Instances {
						z.requireBlockInstance_ContainingElementSpecs(bi)
						if bi.ContainingElementSpecs.Successful() {
							for _, es := range bi.ContainingElementSpecs.Result().Get() {
								z.requireElementSpec_Type(es)
							}
						}
					}
				},
				func(e *ast.Element) {
					ref := bi.Group.Component.File.ElementReferenceByNode(e.Header.Name)
					if ref.Spec == nil {
						return
					}
					z.requireElementSpec_Type(ref.Spec)
				})
		}
	}
}

func (z *analyzer) blockInstance_ForwardsAttributes(bi *file.BlockInstance) file.Analysis[bool] {
	z.requireBlockInstance_ForwardsAttributes(bi)
	return bi.ForwardsAttributes()
}

func (z *analyzer) requireBlockInstance_ForwardsAttributes(bi *file.BlockInstance) {
	if !bi.Group.Component.Analyzed {
		z.Require(bi, blockInstance_CannotForwardAttributes{})
	}
}

func (z *analyzer) blockInstance_DefaultOverwritten(bi *file.BlockInstance, cc *file.ComponentCall) bool {
	z.requireBlockInstance_DefaultOverwritten(bi)
	return bi.DefaultOverwritten(cc)
}

func (z *analyzer) requireBlockInstance_DefaultOverwritten(*file.BlockInstance) {
	// no deps
}

// ============================================================================
// Block Instance Default
// ======================================================================================

func (z *analyzer) blockInstanceDefault_AcceptsAttributes(bi *file.BlockInstance) file.AnalysisWithReason[ast.AndPlaceholderWriter] {
	z.requireBlockInstanceDefault_AcceptsAttributes(bi)
	return bi.Default.AcceptsAttributes
}

func (z *analyzer) requireBlockInstanceDefault_AcceptsAttributes(bi *file.BlockInstance) {
	if !bi.Group.Component.Analyzed {
		z.Require(bi, blockInstanceDefault_AcceptsAttributes{})
	}
}

func (z *analyzer) blockInstanceDefault_ForwardsReceivedAttributes(bi *file.BlockInstance) file.AnalysisWithReason[ast.AndPlaceholderWriter] {
	z.requireBlockInstanceDefault_ForwardsReceivedAttributes(bi)
	return bi.Default.ForwardsReceivedAttributes
}

func (z *analyzer) requireBlockInstanceDefault_ForwardsReceivedAttributes(bi *file.BlockInstance) {
	if !bi.Group.Component.Analyzed {
		z.Require(bi, blockInstanceDefault_ForwardsReceivedAttributes{})
	}
}

func (z *analyzer) blockInstanceDefault_ElementsWithAndPlaceholder(bi *file.BlockInstance) file.Analysis[file.SliceRef[ast.AttributeReceiver]] {
	z.requireBlockInstanceDefault_ElementsWithAndPlaceholder(bi)
	return bi.Default.ElementsWithAndPlaceholder
}

func (z *analyzer) requireBlockInstanceDefault_ElementsWithAndPlaceholder(bi *file.BlockInstance) {
	if !bi.Group.Component.Analyzed {
		z.Require(bi, blockInstanceDefault_ElementsWithAndPlaceholder{})
	}
}

func (z *analyzer) blockInstanceDefault_ElementSpecsWithAndPlaceholder(bi *file.BlockInstance) file.Analysis[file.SliceRef[*file.ElementSpec]] {
	z.requireBlockInstanceDefault_ElementSpecsWithAndPlaceholder(bi)
	return bi.Default.ElementSpecsWithAndPlaceholder
}

func (z *analyzer) requireBlockInstanceDefault_ElementSpecsWithAndPlaceholder(bi *file.BlockInstance) {
	if !bi.Group.Component.Analyzed {
		z.Require(bi, blockInstanceDefault_ElementSpecsWithAndPlaceholder{})
	}
}

func (z *analyzer) blockInstanceDefault_ForwardsAttributes(bi *file.BlockInstance) file.AnalysisWithReason[ast.AttributeWriter] {
	z.requireBlockInstanceDefault_ForwardsAttributes(bi)
	return bi.Default.ForwardsAttributes
}

func (z *analyzer) requireBlockInstanceDefault_ForwardsAttributes(bi *file.BlockInstance) {
	if !bi.Group.Component.Analyzed {
		z.Require(bi, blockInstanceDefault_ForwardsAttributes{})
	}
}

func (z *analyzer) blockInstanceDefault_WritesContent(bi *file.BlockInstance) file.AnalysisWithReason[ast.ContentWriter] {
	z.requireBlockInstanceDefault_WritesContent(bi)
	return bi.Default.WritesContent
}

func (z *analyzer) requireBlockInstanceDefault_WritesContent(bi *file.BlockInstance) {
	if !bi.Group.Component.Analyzed {
		z.Require(bi, blockInstanceDefault_WritesContent{})
	}
}

func (z *analyzer) blockInstanceDefault_WritesElements(bi *file.BlockInstance) file.AnalysisWithReason[ast.ElementWriter] {
	z.requireBlockInstanceDefault_WritesElements(bi)
	return bi.Default.WritesElements
}

func (z *analyzer) requireBlockInstanceDefault_WritesElements(bi *file.BlockInstance) {
	if !bi.Group.Component.Analyzed {
		z.Require(bi, blockInstanceDefault_WritesElements{})
	}
}

// ============================================================================
// Block Setter
// ======================================================================================

func (z *analyzer) blockSetter_WritesAndPlaceholder(s *file.BlockSetter) file.AnalysisWithReason[*file.BlockSetterInstance] {
	z.requireBlockSetter_WritesAndPlaceholder(s)
	return s.WritesAndPlaceholder()
}

func (z *analyzer) requireBlockSetter_WritesAndPlaceholder(s *file.BlockSetter) {
	if !assert.DebugEnabled || s.ComponentCall.Analyzed {
		return
	}
	for _, i := range s.Instances {
		z.requireBlockSetterInstance_AcceptsAttributes(i)
	}
}

func (z *analyzer) blockSetter_ForwardsAndPlaceholder(s *file.BlockSetter) file.AnalysisWithReason[*file.BlockSetterInstance] {
	z.requireBlockSetter_ForwardsAndPlaceholder(s)
	return s.ForwardsAndPlaceholder()
}

func (z *analyzer) requireBlockSetter_ForwardsAndPlaceholder(s *file.BlockSetter) {
	if !assert.DebugEnabled || s.ComponentCall.Analyzed {
		return
	}
	for _, i := range s.Instances {
		z.requireBlockSetterInstance_ForwardsReceivedAttributes(i)
	}
}

func (z *analyzer) blockSetter_ForwardsAttributes(s *file.BlockSetter) file.AnalysisWithReason[*file.BlockSetterInstance] {
	z.requireBlockSetter_ForwardsAttributes(s)
	return s.ForwardsAttributes()
}

func (z *analyzer) requireBlockSetter_ForwardsAttributes(s *file.BlockSetter) {
	if !assert.DebugEnabled || s.ComponentCall.Analyzed {
		return
	}
	for _, i := range s.Instances {
		z.requireBlockSetterInstance_ForwardsAttributes(i)
	}
}

func (z *analyzer) blockSetter_WritesContent(s *file.BlockSetter) file.AnalysisWithReason[*file.BlockSetterInstance] {
	z.requireBlockSetter_WritesContent(s)
	return s.WritesContent()
}

func (z *analyzer) requireBlockSetter_WritesContent(s *file.BlockSetter) {
	if !assert.DebugEnabled || s.ComponentCall.Analyzed {
		return
	}
	for _, i := range s.Instances {
		z.requireBlockSetterInstance_WritesContent(i)
	}
}

func (z *analyzer) blockSetter_WritesElements(s *file.BlockSetter) file.AnalysisWithReason[*file.BlockSetterInstance] {
	z.requireBlockSetter_WritesElements(s)
	return s.WritesElements()
}

func (z *analyzer) requireBlockSetter_WritesElements(s *file.BlockSetter) {
	if !assert.DebugEnabled || s.ComponentCall.Analyzed {
		return
	}
	for _, i := range s.Instances {
		z.requireBlockSetterInstance_WritesElements(i)
	}
}

// ============================================================================
// Block Setter Instance
// ======================================================================================

func (z *analyzer) requireBlockSetterInstance_AcceptsAttributes(i *file.BlockSetterInstance) {
	if !i.Group.ComponentCall.Analyzed {
		z.Require(i, blockSetterInstance_AcceptsAttributes{})
	}
}

func (z *analyzer) requireBlockSetterInstance_ForwardsReceivedAttributes(i *file.BlockSetterInstance) {
	if !i.Group.ComponentCall.Analyzed {
		z.Require(i, blockSetterInstance_ForwardsReceivedAttributes{})
	}
}

func (z *analyzer) requireBlockSetterInstance_ForwardsAttributes(i *file.BlockSetterInstance) {
	if !i.Group.ComponentCall.Analyzed {
		z.Require(i, blockSetterInstance_ForwardsAttributes{})
	}
}

func (z *analyzer) requireBlockSetterInstance_WritesContent(i *file.BlockSetterInstance) {
	if !i.Group.ComponentCall.Analyzed {
		z.Require(i, blockSetterInstance_WritesContent{})
	}
}

func (z *analyzer) requireBlockSetterInstance_WritesElements(i *file.BlockSetterInstance) {
	if !i.Group.ComponentCall.Analyzed {
		z.Require(i, blockSetterInstance_WritesElements{})
	}
}

// ============================================================================
// Component
// ======================================================================================

func (z *analyzer) component_Circular(c *file.Component) bool {
	z.requireComponent_Circular(c)
	return c.Circular
}

func (z *analyzer) requireComponent_Circular(c *file.Component) {
	if !c.Analyzed {
		z.Require(c, component_Circular{})
	}
}

func (z *analyzer) component_AlwaysAcceptsAttributes(c *file.Component) file.AnalysisWithReason[ast.AndPlaceholderWriter] {
	z.requireComponent_AlwaysAcceptsAttributes(c)
	return c.AlwaysAcceptsAttributes
}

func (z *analyzer) requireComponent_AlwaysAcceptsAttributes(c *file.Component) {
	if !c.Analyzed {
		z.Require(c, component_AlwaysAcceptsAttributes{})
	}
}

func (z *analyzer) component_AlwaysForwardsReceivedAttributes(c *file.Component) file.AnalysisWithReason[ast.AndPlaceholderWriter] {
	z.requireComponent_AlwaysForwardsReceivedAttributes(c)
	return c.AlwaysForwardsReceivedAttributes
}

func (z *analyzer) requireComponent_AlwaysForwardsReceivedAttributes(c *file.Component) {
	if !c.Analyzed {
		z.Require(c, component_AlwaysForwardsReceivedAttributes{})
	}
}

func (z *analyzer) component_AlwaysForwardsAttributes(c *file.Component) file.AnalysisWithReason[ast.AttributeWriter] {
	z.requireComponent_AlwaysForwardsAttributes(c)
	return c.AlwaysForwardsAttributes
}

func (z *analyzer) requireComponent_AlwaysForwardsAttributes(c *file.Component) {
	if !c.Analyzed {
		z.Require(c, component_AlwaysForwardsAttributes{})
	}
}

func (z *analyzer) component_AlwaysWritesContent(c *file.Component) file.AnalysisWithReason[ast.ContentWriter] {
	z.requireComponent_AlwaysWritesContent(c)
	return c.AlwaysWritesContent
}

func (z *analyzer) requireComponent_AlwaysWritesContent(c *file.Component) {
	if !c.Analyzed {
		z.Require(c, component_AlwaysWritesContent{})
	}
}

func (z *analyzer) component_AlwaysWritesElements(c *file.Component) file.AnalysisWithReason[ast.ElementWriter] {
	z.requireComponent_AlwaysWritesElements(c)
	return c.AlwaysWritesElements
}

func (z *analyzer) requireComponent_AlwaysWritesElements(c *file.Component) {
	if !c.Analyzed {
		z.Require(c, component_AlwaysWritesElements{})
	}
}

func (z *analyzer) component_PermanentElementsWithAndPlaceholder(c *file.Component) file.Analysis[file.SliceRef[ast.AttributeReceiver]] {
	z.requireComponent_PermanentElementsWithAndPlaceholder(c)
	return c.PermanentElementsWithAndPlaceholder
}

func (z *analyzer) requireComponent_PermanentElementsWithAndPlaceholder(c *file.Component) {
	if !c.Analyzed {
		z.Require(c, component_PermanentElementsWithAndPlaceholder{})
	}
}

func (z *analyzer) component_PermanentElementSpecsWithAndPlaceholder(c *file.Component) file.Analysis[file.SliceRef[*file.ElementSpec]] {
	z.requireComponent_PermanentElementSpecsWithAndPlaceholder(c)
	return c.PermanentElementSpecsWithAndPlaceholder
}

func (z *analyzer) requireComponent_PermanentElementSpecsWithAndPlaceholder(c *file.Component) {
	if !c.Analyzed {
		z.Require(c, component_PermanentElementSpecsWithAndPlaceholder{})
	}
}

func (z *analyzer) component_CouldAcceptAttributes(c *file.Component) file.AnalysisWithReason[ast.AndPlaceholderWriter] {
	z.requireComponent_CouldAcceptAttributes(c)
	return c.CouldAcceptAttributes()
}

func (z *analyzer) requireComponent_CouldAcceptAttributes(c *file.Component) {
	if c.Analyzed || !assert.DebugEnabled {
		return
	}
	z.requireComponent_AlwaysAcceptsAttributes(c)
	for _, b := range c.Blocks {
		for _, bi := range b.Instances {
			if bi.Default == nil {
				continue
			}
			z.requireBlockInstanceDefault_AcceptsAttributes(bi)
		}
	}
}

func (z *analyzer) component_CouldForwardReceivedAttributes(c *file.Component) file.AnalysisWithReason[ast.AndPlaceholderWriter] {
	z.requireComponent_CouldForwardReceivedAttributes(c)
	return c.CouldForwardReceivedAttributes()
}

func (z *analyzer) requireComponent_CouldForwardReceivedAttributes(c *file.Component) {
	if c.Analyzed || !assert.DebugEnabled {
		return
	}
	z.requireComponent_AlwaysForwardsReceivedAttributes(c)
	for _, b := range c.Blocks {
		for _, bi := range b.Instances {
			if bi.Default == nil {
				continue
			}
			z.requireBlockInstance_Forwarded(bi)
			z.requireBlockInstanceDefault_ForwardsReceivedAttributes(bi)
		}
	}
}

// ============================================================================
// Component Argument
// ======================================================================================

func (z *analyzer) componentArgument_Value(arg *file.ComponentArgument) file.ResolvedValue {
	z.requireComponentArgument_Value(arg)
	return arg.Value
}

func (z *analyzer) requireComponentArgument_Value(arg *file.ComponentArgument) {
	if !arg.ComponentCall.Analyzed {
		z.Require(arg, componentArgument_Value{})
	}
}

// ============================================================================
// Component Parameter
// ======================================================================================

func (z *analyzer) componentParameter_InferredType(p *file.ComponentParameter) file.Analysis[file.Type] {
	z.requireComponentParameter_InferredType(p)
	return p.InferredType
}

func (z *analyzer) requireComponentParameter_InferredType(p *file.ComponentParameter) {
	if !p.Component.Analyzed {
		z.Require(p, componentParameter_InferredType{})
	}
}

func (z *analyzer) componentParameter_AttributeType(p *file.ComponentParameter) file.Analysis[attrtype.Type] {
	z.requireComponentParameter_AttributeType(p)
	return p.AttributeType
}

func (z *analyzer) requireComponentParameter_AttributeType(p *file.ComponentParameter) {
	if !p.Component.Analyzed {
		z.Require(p, componentParameter_AttributeType{})
	}
}

func (z *analyzer) componentParameter_AttributeName(p *file.ComponentParameter) file.Analysis[string] {
	z.requireComponentParameter_AttributeName(p)
	return p.AttributeName
}

func (z *analyzer) requireComponentParameter_AttributeName(p *file.ComponentParameter) {
	if !p.Component.Analyzed {
		z.Require(p, componentParameter_AttributeName{})
	}
}

func (z *analyzer) componentParameter_ResolvedType(p *file.ComponentParameter) file.Analysis[file.Type] {
	z.requireComponentParameter_ResolvedType(p)
	return p.ResolvedType()
}

func (z *analyzer) requireComponentParameter_ResolvedType(p *file.ComponentParameter) {
	if !assert.DebugEnabled || p.Component.Analyzed {
		return
	}
	// If explicit type is provided, ResolvedType does not depend on analyses.
	if p.AST.Type == nil {
		z.requireComponentParameter_InferredType(p)
	}
}

func (z *analyzer) componentParameter_Required(p *file.ComponentParameter) bool {
	z.requireComponentParameter_Required(p)
	return p.Required()
}

func (z *analyzer) requireComponentParameter_Required(*file.ComponentParameter) {
	// no deps
}

// ============================================================================
// Component Call
// ======================================================================================

func (z *analyzer) componentCall_ReceivesAttributes(cc *file.ComponentCall) file.AnalysisWithReason[ast.AttributeWriter] {
	z.requireComponentCall_ReceivesAttributes(cc)
	return cc.ReceivesAttributes
}

func (z *analyzer) requireComponentCall_ReceivesAttributes(cc *file.ComponentCall) {
	if !cc.Analyzed {
		z.Require(cc, componentCall_ReceivesAttributes{})
	}
}

func (z *analyzer) componentCall_ReceivesAndPlaceholder(cc *file.ComponentCall) file.AnalysisWithReason[ast.AndPlaceholderWriter] {
	z.requireComponentCall_ReceivesAndPlaceholder(cc)
	return cc.ReceivesAndPlaceholder
}

func (z *analyzer) requireComponentCall_ReceivesAndPlaceholder(cc *file.ComponentCall) {
	if !cc.Analyzed {
		z.Require(cc, componentCall_ReceivesAndPlaceholder{})
	}
}

func (z *analyzer) componentCall_ElementsWithAndPlaceholder(cc *file.ComponentCall) file.Analysis[file.SliceRef[ast.AttributeReceiver]] {
	z.requireComponentCall_ElementsWithAndPlaceholder(cc)
	return cc.ElementsWithAndPlaceholder
}

func (z *analyzer) requireComponentCall_ElementsWithAndPlaceholder(cc *file.ComponentCall) {
	if !cc.Analyzed {
		z.Require(cc, componentCall_ElementsWithAndPlaceholder{})
	}
}

func (z *analyzer) componentCall_ElementSpecsWithAndPlaceholder(cc *file.ComponentCall) file.Analysis[file.SliceRef[*file.ElementSpec]] {
	z.requireComponentCall_ElementSpecsWithAndPlaceholder(cc)
	return cc.ElementSpecsWithAndPlaceholder
}

func (z *analyzer) requireComponentCall_ElementSpecsWithAndPlaceholder(cc *file.ComponentCall) {
	if !cc.Analyzed {
		z.Require(cc, componentCall_ElementSpecsWithAndPlaceholder{})
	}
}

func (z *analyzer) componentCall_AcceptsAttributes(cc *file.ComponentCall) file.AnalysisWithReason[ast.AndPlaceholderWriter] {
	z.requireComponentCall_AcceptsAttributes(cc)
	return cc.AcceptsAttributes()
}

func (z *analyzer) requireComponentCall_AcceptsAttributes(cc *file.ComponentCall) {
	if cc.Analyzed || !assert.DebugEnabled {
		return
	}
	z.requireComponent_AlwaysAcceptsAttributes(cc.Component)
	for _, b := range cc.Component.Blocks {
		for _, bi := range b.Instances {
			if bi.Default == nil {
				continue
			}
			z.requireBlockInstance_DefaultOverwritten(bi)
			z.requireBlockInstanceDefault_AcceptsAttributes(bi)
		}
	}
}

func (z *analyzer) componentCall_ForwardsAcceptedAttributes(cc *file.ComponentCall) file.AnalysisWithReason[ast.AndPlaceholderWriter] {
	z.requireComponentCall_ForwardsAcceptedAttributes(cc)
	return cc.ForwardsAcceptedAttributes()
}

func (z *analyzer) requireComponentCall_ForwardsAcceptedAttributes(cc *file.ComponentCall) {
	if !assert.DebugEnabled || cc.Analyzed {
		return
	}
	z.requireComponent_AlwaysForwardsReceivedAttributes(cc.Component)
	for _, b := range cc.Component.Blocks {
		for _, bi := range b.Instances {
			if bi.Default == nil {
				continue
			}
			z.requireBlockInstance_DefaultOverwritten(bi)
			z.requireBlockInstance_Forwarded(bi)
			z.requireBlockInstanceDefault_ForwardsReceivedAttributes(bi)
		}
	}
}

func (z *analyzer) componentCall_ForwardsAndPlaceholder(cc *file.ComponentCall) file.Analysis[bool] {
	z.requireComponentCall_ForwardsAndPlaceholder(cc)
	return cc.ForwardsAndPlaceholder()
}

func (z *analyzer) requireComponentCall_ForwardsAndPlaceholder(cc *file.ComponentCall) {
	if !cc.Analyzed {
		z.requireComponentCall_ForwardsReceivedAndPlaceholder(cc)
		z.requireComponentCall_BlockSetterForwardsAndPlaceholder(cc)
	}
}

func (z *analyzer) componentCall_ForwardsReceivedAndPlaceholder(cc *file.ComponentCall) file.AnalysisWithReason[ast.AndPlaceholderWriter] {
	z.requireComponentCall_ForwardsReceivedAndPlaceholder(cc)
	return cc.ForwardsReceivedAndPlaceholder()
}

func (z *analyzer) requireComponentCall_ForwardsReceivedAndPlaceholder(cc *file.ComponentCall) {
	if !cc.Analyzed {
		z.requireComponentCall_ForwardsAcceptedAttributes(cc)
		z.requireComponentCall_ReceivesAndPlaceholder(cc)
	}
}

func (z *analyzer) componentCall_BlockSetterForwardsAndPlaceholder(cc *file.ComponentCall) file.AnalysisWithReason[*file.BlockSetter] {
	z.requireComponentCall_BlockSetterForwardsAndPlaceholder(cc)
	return cc.BlockSetterForwardsAndPlaceholder()
}

func (z *analyzer) requireComponentCall_BlockSetterForwardsAndPlaceholder(cc *file.ComponentCall) {
	if !assert.DebugEnabled || cc.Analyzed {
		return
	}
	for _, s := range cc.BlockSetters {
		z.requireBlockSetter_ForwardsAndPlaceholder(s)
		z.requireBlock_Forwarded(s.Block)
	}
}

func (z *analyzer) componentCall_ForwardsAttributes(cc *file.ComponentCall) file.Analysis[bool] {
	z.requireComponentCall_ForwardsAttributes(cc)
	return cc.ForwardsAttributes()
}

func (z *analyzer) requireComponentCall_ForwardsAttributes(cc *file.ComponentCall) {
	if !cc.Analyzed {
		z.requireComponentCall_ComponentForwardsAttributes(cc)
		z.requireComponentCall_ForwardsReceivedAttributes(cc)
		z.requireComponentCall_BlockSetterForwardsAttributes(cc)
	}
}

func (z *analyzer) componentCall_ComponentForwardsAttributes(cc *file.ComponentCall) file.AnalysisWithReason[ast.AttributeWriter] {
	z.requireComponentCall_ComponentForwardsAttributes(cc)
	return cc.ComponentForwardsAttributes()
}

func (z *analyzer) requireComponentCall_ComponentForwardsAttributes(cc *file.ComponentCall) {
	if !assert.DebugEnabled || cc.Analyzed {
		return
	}
	z.requireComponent_AlwaysForwardsAttributes(cc.Component)
	for _, b := range cc.Component.Blocks {
		for _, bi := range b.Instances {
			if bi.Default == nil {
				continue
			}
			z.requireBlockInstance_DefaultOverwritten(bi)
			z.requireBlockInstanceDefault_ForwardsAttributes(bi)
		}
	}
}

// ForwardsReceivedAttributes indicates that the component receives attributes
// and forwards it out of the component again.
func (z *analyzer) componentCall_ForwardsReceivedAttributes(cc *file.ComponentCall) file.AnalysisWithReason[ast.AttributeWriter] {
	z.requireComponentCall_ForwardsReceivedAttributes(cc)
	return cc.ForwardsReceivedAttributes()
}

func (z *analyzer) requireComponentCall_ForwardsReceivedAttributes(cc *file.ComponentCall) {
	if !cc.Analyzed {
		z.requireComponentCall_ForwardsAcceptedAttributes(cc)
		z.requireComponentCall_ReceivesAttributes(cc)
	}
}

func (z *analyzer) componentCall_BlockSetterForwardsAttributes(cc *file.ComponentCall) file.AnalysisWithReason[*file.BlockSetter] {
	z.requireComponentCall_BlockSetterForwardsAttributes(cc)
	return cc.BlockSetterForwardsAttributes()
}

func (z *analyzer) requireComponentCall_BlockSetterForwardsAttributes(cc *file.ComponentCall) {
	if !assert.DebugEnabled || cc.Analyzed {
		return
	}
	for _, s := range cc.BlockSetters {
		z.requireBlockSetter_ForwardsAttributes(s)
		z.requireBlock_Forwarded(s.Block)
	}
}

func (z *analyzer) componentCall_WritesContent(cc *file.ComponentCall) file.Analysis[bool] {
	z.requireComponentCall_WritesContent(cc)
	return cc.WritesContent()
}

func (z *analyzer) requireComponentCall_WritesContent(cc *file.ComponentCall) {
	if !cc.Analyzed {
		z.requireComponentCall_ComponentWritesContent(cc)
		z.requireComponentCall_BlockSetterWritesContent(cc)
	}
}

func (z *analyzer) componentCall_ComponentWritesContent(cc *file.ComponentCall) file.AnalysisWithReason[ast.ContentWriter] {
	z.requireComponentCall_ComponentWritesContent(cc)
	return cc.ComponentWritesContent()
}

func (z *analyzer) requireComponentCall_ComponentWritesContent(cc *file.ComponentCall) {
	if !assert.DebugEnabled || cc.Analyzed {
		return
	}
	z.requireComponent_AlwaysWritesContent(cc.Component)
	for _, b := range cc.Component.Blocks {
		for _, bi := range b.Instances {
			if bi.Default == nil {
				continue
			}
			z.requireBlockInstance_DefaultOverwritten(bi)
			z.requireBlockInstanceDefault_WritesContent(bi)
		}
	}
}

func (z *analyzer) componentCall_BlockSetterWritesContent(cc *file.ComponentCall) file.AnalysisWithReason[*file.BlockSetter] {
	z.requireComponentCall_BlockSetterWritesContent(cc)
	return cc.BlockSetterWritesContent()
}

func (z *analyzer) requireComponentCall_BlockSetterWritesContent(cc *file.ComponentCall) {
	if !assert.DebugEnabled || cc.Analyzed {
		return
	}
	for _, s := range cc.BlockSetters {
		z.requireBlockSetter_WritesContent(s)
	}
}

func (z *analyzer) componentCall_WritesElements(cc *file.ComponentCall) file.Analysis[bool] {
	z.requireComponentCall_WritesElements(cc)
	return cc.WritesElements()
}

func (z *analyzer) requireComponentCall_WritesElements(cc *file.ComponentCall) {
	if !cc.Analyzed {
		z.requireComponentCall_ComponentWritesElements(cc)
		z.requireComponentCall_BlockSetterWritesElements(cc)
	}
}

func (z *analyzer) componentCall_ComponentWritesElements(cc *file.ComponentCall) file.AnalysisWithReason[ast.ElementWriter] {
	z.requireComponentCall_ComponentWritesElements(cc)
	return cc.ComponentWritesElements()
}

func (z *analyzer) requireComponentCall_ComponentWritesElements(cc *file.ComponentCall) {
	if !assert.DebugEnabled || cc.Analyzed {
		return
	}
	z.requireComponent_AlwaysWritesElements(cc.Component)
	for _, b := range cc.Component.Blocks {
		for _, bi := range b.Instances {
			if bi.Default == nil {
				continue
			}
			z.requireBlockInstance_DefaultOverwritten(bi)
			z.requireBlockInstanceDefault_WritesElements(bi)
		}
	}
}

func (z *analyzer) componentCall_BlockSetterWritesElements(cc *file.ComponentCall) file.AnalysisWithReason[*file.BlockSetter] {
	z.requireComponentCall_BlockSetterWritesElements(cc)
	return cc.BlockSetterWritesElements()
}

func (z *analyzer) requireComponentCall_BlockSetterWritesElements(cc *file.ComponentCall) {
	if !assert.DebugEnabled || cc.Analyzed {
		return
	}
	for _, s := range cc.BlockSetters {
		z.requireBlockSetter_WritesElements(s)
	}
}

// ============================================================================
// Element Spec
// ======================================================================================

func (z *analyzer) elementSpec_Circular(es *file.ElementSpec) bool {
	z.requireElementSpec_Circular(es)
	return es.Circular
}

func (z *analyzer) requireElementSpec_Circular(es *file.ElementSpec) {
	if !es.Analyzed {
		z.Require(es, elementSpec_Circular{})
	}
}

func (z *analyzer) elementSpec_Type(es *file.ElementSpec) file.Analysis[elemtype.Type] {
	z.requireElementSpec_Type(es)
	return es.Type
}

func (z *analyzer) requireElementSpec_Type(es *file.ElementSpec) {
	if !es.Analyzed {
		z.Require(es, elementSpec_Type{})
	}
}
