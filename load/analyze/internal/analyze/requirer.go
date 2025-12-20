package analyze

import (
	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/ast"
	"github.com/mavolin/corgi/v2/file/switches"
	"github.com/mavolin/corgi/v2/internal/assert"
)

type requirer struct{}

// =============================================================================
// Attribute
// ======================================================================================

func (requirer) Attribute_Value(z *analyzer, a *file.Attribute) {
	if !a.Analyzed {
		requireField.Attribute.Value(z, a)
	}
}

func (requirer) Attribute_Forwarded(z *analyzer, a *file.Attribute) {
	if !a.Analyzed {
		requireField.Attribute.Forwarded(z, a)
	}
}

func (requirer) Attribute_Receivers(z *analyzer, a *file.Attribute) {
	if !a.Analyzed {
		requireField.Attribute.Receivers(z, a)
	}
}

func (requirer) Attribute_ReceivingElementSpecs(z *analyzer, a *file.Attribute) {
	if !a.Analyzed {
		requireField.Attribute.ReceivingElementSpecs(z, a)
	}
}

func (requirer) Attribute_Type(z *analyzer, a *file.Attribute) {
	if !a.Analyzed {
		requireField.Attribute.Type(z, a)
	}
}

// =========================================================================
// Block
// ======================================================================================

func (requirer) Block_Required(*analyzer, *file.Block) {
	// no deps
}

func (requirer) Block_Forwarded(z *analyzer, b *file.Block) {
	if !assert.DebugEnabled || b.Component.Analyzed || b.Component.Circular {
		return
	}
	for _, bi := range b.Instances {
		require.BlockInstance.Forwarded(z, bi)
	}
}

func (requirer) Block_ForwardsAttributes(z *analyzer, b *file.Block) {
	if !b.Component.Analyzed && !b.Component.Circular {
		require.Block.CannotForwardAttributes(z, b)
	}
}

func (requirer) Block_CannotForwardAttributes(z *analyzer, b *file.Block) {
	if !assert.DebugEnabled || b.Component.Analyzed || b.Component.Circular {
		return
	}
	for _, bi := range b.Instances {
		require.BlockInstance.CannotForwardAttributes(z, bi)
	}
}

func (requirer) Block_ElementType(z *analyzer, b *file.Block) {
	if !assert.DebugEnabled || b.Component.Analyzed || b.Component.Circular {
		return
	}
	for _, bi := range b.Instances {
		require.BlockInstance.ContainingElementSpecs(z, bi)
		if bi.ContainingElementSpecs.Successful() {
			for _, es := range bi.ContainingElementSpecs.Result().Get() {
				require.ElementSpec.Type(z, es)
			}
		}
	}
}

func (requirer) Block_MostRestrictiveElement(z *analyzer, b *file.Block) {
	if !assert.DebugEnabled || b.Component.Analyzed || b.Component.Circular {
		return
	}
	for _, bi := range b.Instances {
		require.BlockInstance.ContainingElements(z, bi)
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
							require.BlockInstance.ContainingElementSpecs(z, bi)
							if bi.ContainingElementSpecs.Successful() {
								for _, es := range bi.ContainingElementSpecs.Result().Get() {
									require.ElementSpec.Type(z, es)
								}
							}
						}
					},
					func(e *ast.Element) {
						ref := b.Component.File.ElementReferenceByNode(e.Header.Name)
						if ref.Spec == nil {
							return
						}
						require.ElementSpec.Type(z, ref.Spec)
					})
			}
		}
	}
}

// =========================================================================
// Block Instance
// ======================================================================================

func (requirer) BlockInstance_Forwarded(z *analyzer, bi *file.BlockInstance) {
	if !bi.Group.Component.Analyzed && !bi.Group.Component.Circular {
		requireField.BlockInstance.Forwarded(z, bi)
	}
}

func (requirer) BlockInstance_ContainingElements(z *analyzer, bi *file.BlockInstance) {
	if !bi.Group.Component.Analyzed && !bi.Group.Component.Circular {
		requireField.BlockInstance.ContainingElements(z, bi)
	}
}

func (requirer) BlockInstance_ContainingElementSpecs(z *analyzer, bi *file.BlockInstance) {
	if !bi.Group.Component.Analyzed && !bi.Group.Component.Circular {
		requireField.BlockInstance.ContainingElementSpecs(z, bi)
	}
}

func (requirer) BlockInstance_CannotForwardAttributes(z *analyzer, bi *file.BlockInstance) {
	if !bi.Group.Component.Analyzed && !bi.Group.Component.Circular {
		requireField.BlockInstance.CannotForwardAttributes(z, bi)
	}
}

func (requirer) BlockInstance_ElementType(z *analyzer, bi *file.BlockInstance) {
	if !assert.DebugEnabled || bi.Group.Component.Analyzed || bi.Group.Component.Circular {
		return
	}
	require.BlockInstance.ContainingElementSpecs(z, bi)
	if bi.ContainingElementSpecs.Successful() {
		for _, es := range bi.ContainingElementSpecs.Result().Get() {
			require.ElementSpec.Type(z, es)
		}
	}
}

func (requirer) BlockInstance_MostRestrictiveElement(z *analyzer, bi *file.BlockInstance) {
	if !assert.DebugEnabled || bi.Group.Component.Analyzed || bi.Group.Component.Circular {
		return
	}
	require.BlockInstance.ContainingElements(z, bi)
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
						require.BlockInstance.ContainingElementSpecs(z, bi)
						if bi.ContainingElementSpecs.Successful() {
							for _, es := range bi.ContainingElementSpecs.Result().Get() {
								require.ElementSpec.Type(z, es)
							}
						}
					}
				},
				func(e *ast.Element) {
					ref := bi.Group.Component.File.ElementReferenceByNode(e.Header.Name)
					if ref.Spec == nil {
						return
					}
					require.ElementSpec.Type(z, ref.Spec)
				})
		}
	}
}

func (requirer) BlockInstance_ForwardsAttributes(z *analyzer, bi *file.BlockInstance) {
	if !bi.Group.Component.Analyzed {
		requireField.BlockInstance.CannotForwardAttributes(z, bi)
	}
}

func (requirer) BlockInstance_DefaultOverwritten(*analyzer, *file.BlockInstance) {
	// no deps
}

// =========================================================================
// Block Instance Default
// ======================================================================================

func (requirer) BlockInstance_Default_AcceptsAttributes(z *analyzer, bi *file.BlockInstance) {
	if !bi.Group.Component.Analyzed {
		requireField.BlockInstance.Default.AcceptsAttributes(z, bi)
	}
}

func (requirer) BlockInstance_Default_ForwardsReceivedAttributes(z *analyzer, bi *file.BlockInstance) {
	if !bi.Group.Component.Analyzed {
		requireField.BlockInstance.Default.ForwardsReceivedAttributes(z, bi)
	}
}

func (requirer) BlockInstance_Default_ElementsWithAndPlaceholder(z *analyzer, bi *file.BlockInstance) {
	if !bi.Group.Component.Analyzed {
		requireField.BlockInstance.Default.ElementsWithAndPlaceholder(z, bi)
	}
}

func (requirer) BlockInstance_Default_ElementSpecsWithAndPlaceholder(z *analyzer, bi *file.BlockInstance) {
	if !bi.Group.Component.Analyzed {
		requireField.BlockInstance.Default.ElementSpecsWithAndPlaceholder(z, bi)
	}
}

func (requirer) BlockInstance_Default_ForwardsAttributes(z *analyzer, bi *file.BlockInstance) {
	if !bi.Group.Component.Analyzed {
		requireField.BlockInstance.Default.ForwardsAttributes(z, bi)
	}
}

func (requirer) BlockInstance_Default_WritesContent(z *analyzer, bi *file.BlockInstance) {
	if !bi.Group.Component.Analyzed {
		requireField.BlockInstance.Default.WritesContent(z, bi)
	}
}

func (requirer) BlockInstance_Default_WritesElements(z *analyzer, bi *file.BlockInstance) {
	if !bi.Group.Component.Analyzed {
		requireField.BlockInstance.Default.WritesElements(z, bi)
	}
}

// =========================================================================
// Block Setter
// ======================================================================================

func (requirer) BlockSetter_WritesAndPlaceholder(z *analyzer, s *file.BlockSetter) {
	if !assert.DebugEnabled || s.ComponentCall.Analyzed {
		return
	}
	for _, i := range s.Instances {
		require.BlockSetterInstance.AcceptsAttributes(z, i)
	}
}

func (requirer) BlockSetter_ForwardsAndPlaceholder(z *analyzer, s *file.BlockSetter) {
	if !assert.DebugEnabled || s.ComponentCall.Analyzed {
		return
	}
	for _, i := range s.Instances {
		require.BlockSetterInstance.ForwardsReceivedAttributes(z, i)
	}
}

func (requirer) BlockSetter_ForwardsAttributes(z *analyzer, s *file.BlockSetter) {
	if !assert.DebugEnabled || s.ComponentCall.Analyzed {
		return
	}
	for _, i := range s.Instances {
		require.BlockSetterInstance.ForwardsAttributes(z, i)
	}
}

func (requirer) BlockSetter_WritesContent(z *analyzer, s *file.BlockSetter) {
	if !assert.DebugEnabled || s.ComponentCall.Analyzed {
		return
	}
	for _, i := range s.Instances {
		require.BlockSetterInstance.WritesContent(z, i)
	}
}

func (requirer) BlockSetter_WritesElements(z *analyzer, s *file.BlockSetter) {
	if !assert.DebugEnabled || s.ComponentCall.Analyzed {
		return
	}
	for _, i := range s.Instances {
		require.BlockSetterInstance.WritesElements(z, i)
	}
}

// =========================================================================
// Block Setter Instance
// ======================================================================================

func (requirer) BlockSetterInstance_AcceptsAttributes(z *analyzer, bsi *file.BlockSetterInstance) {
	if !bsi.Group.ComponentCall.Analyzed {
		requireField.BlockSetterInstance.AcceptsAttributes(z, bsi)
	}
}

func (requirer) BlockSetterInstance_ForwardsReceivedAttributes(z *analyzer, bsi *file.BlockSetterInstance) {
	if !bsi.Group.ComponentCall.Analyzed {
		requireField.BlockSetterInstance.ForwardsReceivedAttributes(z, bsi)
	}
}

func (requirer) BlockSetterInstance_ForwardsAttributes(z *analyzer, bsi *file.BlockSetterInstance) {
	if !bsi.Group.ComponentCall.Analyzed {
		requireField.BlockSetterInstance.ForwardsAttributes(z, bsi)
	}
}

func (requirer) BlockSetterInstance_WritesContent(z *analyzer, bsi *file.BlockSetterInstance) {
	if !bsi.Group.ComponentCall.Analyzed {
		requireField.BlockSetterInstance.WritesContent(z, bsi)
	}
}

func (requirer) BlockSetterInstance_WritesElements(z *analyzer, bsi *file.BlockSetterInstance) {
	if !bsi.Group.ComponentCall.Analyzed {
		requireField.BlockSetterInstance.WritesElements(z, bsi)
	}
}

// =========================================================================
// Component
// ======================================================================================

func (requirer) Component_Circular(z *analyzer, c *file.Component) {
	if !c.Analyzed {
		requireField.Component.Circular(z, c)
	}
}

func (requirer) Component_AlwaysAcceptsAttributes(z *analyzer, c *file.Component) {
	if !c.Analyzed {
		requireField.Component.AlwaysAcceptsAttributes(z, c)
	}
}

func (requirer) Component_AlwaysForwardsReceivedAttributes(z *analyzer, c *file.Component) {
	if !c.Analyzed {
		requireField.Component.AlwaysForwardsReceivedAttributes(z, c)
	}
}

func (requirer) Component_AlwaysForwardsAttributes(z *analyzer, c *file.Component) {
	if !c.Analyzed {
		requireField.Component.AlwaysForwardsAttributes(z, c)
	}
}

func (requirer) Component_AlwaysWritesContent(z *analyzer, c *file.Component) {
	if !c.Analyzed {
		requireField.Component.AlwaysWritesContent(z, c)
	}
}

func (requirer) Component_AlwaysWritesElements(z *analyzer, c *file.Component) {
	if !c.Analyzed {
		requireField.Component.AlwaysWritesElements(z, c)
	}
}

func (requirer) Component_PermanentElementsWithAndPlaceholder(z *analyzer, c *file.Component) {
	if !c.Analyzed {
		requireField.Component.PermanentElementsWithAndPlaceholder(z, c)
	}
}

func (requirer) Component_PermanentElementSpecsWithAndPlaceholder(z *analyzer, c *file.Component) {
	if !c.Analyzed {
		requireField.Component.PermanentElementSpecsWithAndPlaceholder(z, c)
	}
}

func (requirer) Component_CouldAcceptAttributes(z *analyzer, c *file.Component) {
	if c.Analyzed || !assert.DebugEnabled {
		return
	}
	require.Component.AlwaysAcceptsAttributes(z, c)
	for _, b := range c.Blocks {
		for _, bi := range b.Instances {
			if bi.Default == nil {
				continue
			}
			require.BlockInstance.Default.AcceptsAttributes(z, bi)
		}
	}
}

func (requirer) Component_CouldForwardReceivedAttributes(z *analyzer, c *file.Component) {
	if c.Analyzed || !assert.DebugEnabled {
		return
	}
	require.Component.AlwaysForwardsReceivedAttributes(z, c)
	for _, b := range c.Blocks {
		for _, bi := range b.Instances {
			if bi.Default == nil {
				continue
			}
			require.BlockInstance.Forwarded(z, bi)
			require.BlockInstance.Default.ForwardsReceivedAttributes(z, bi)
		}
	}
}

// =========================================================================
// Component Argument
// ======================================================================================

func (requirer) ComponentArgument_Value(z *analyzer, arg *file.ComponentArgument) {
	if !arg.ComponentCall.Analyzed {
		requireField.ComponentArgument.Value(z, arg)
	}
}

// =========================================================================
// Component Parameter
// ======================================================================================

func (requirer) ComponentParameter_InferredType(z *analyzer, p *file.ComponentParameter) {
	if !p.Component.Analyzed {
		requireField.ComponentParameter.InferredType(z, p)
	}
}

func (requirer) ComponentParameter_AttributeType(z *analyzer, p *file.ComponentParameter) {
	if !p.Component.Analyzed {
		requireField.ComponentParameter.AttributeType(z, p)
	}
}

func (requirer) ComponentParameter_AttributeName(z *analyzer, p *file.ComponentParameter) {
	if !p.Component.Analyzed {
		requireField.ComponentParameter.AttributeName(z, p)
	}
}

func (requirer) ComponentParameter_ResolvedType(z *analyzer, p *file.ComponentParameter) {
	if !assert.DebugEnabled || p.Component.Analyzed {
		return
	}
	// If explicit type is provided, ResolvedType does not depend on analyses.
	if p.AST.Type == nil {
		require.ComponentParameter.InferredType(z, p)
	}
}

func (requirer) ComponentParameter_Required(*analyzer, *file.ComponentParameter) {
	// no deps
}

// =========================================================================
// Component Call
// ======================================================================================

func (requirer) ComponentCall_ReceivesAttributes(z *analyzer, cc *file.ComponentCall) {
	if !cc.Analyzed {
		requireField.ComponentCall.ReceivesAttributes(z, cc)
	}
}

func (requirer) ComponentCall_ReceivesAndPlaceholder(z *analyzer, cc *file.ComponentCall) {
	if !cc.Analyzed {
		requireField.ComponentCall.ReceivesAndPlaceholder(z, cc)
	}
}

func (requirer) ComponentCall_ElementsWithAndPlaceholder(z *analyzer, cc *file.ComponentCall) {
	if !cc.Analyzed {
		requireField.ComponentCall.ElementsWithAndPlaceholder(z, cc)
	}
}

func (requirer) ComponentCall_ElementSpecsWithAndPlaceholder(z *analyzer, cc *file.ComponentCall) {
	if !cc.Analyzed {
		requireField.ComponentCall.ElementSpecsWithAndPlaceholder(z, cc)
	}
}

func (requirer) ComponentCall_AcceptsAttributes(z *analyzer, cc *file.ComponentCall) {
	if cc.Analyzed || !assert.DebugEnabled {
		return
	}
	require.Component.AlwaysAcceptsAttributes(z, cc.Component)
	for _, b := range cc.Component.Blocks {
		for _, bi := range b.Instances {
			if bi.Default == nil {
				continue
			}
			require.BlockInstance.DefaultOverwritten(z, bi)
			require.BlockInstance.Default.AcceptsAttributes(z, bi)
		}
	}
}

func (requirer) ComponentCall_ForwardsAcceptedAttributes(z *analyzer, cc *file.ComponentCall) {
	if !assert.DebugEnabled || cc.Analyzed {
		return
	}
	require.Component.AlwaysForwardsReceivedAttributes(z, cc.Component)
	for _, b := range cc.Component.Blocks {
		for _, bi := range b.Instances {
			if bi.Default == nil {
				continue
			}
			require.BlockInstance.DefaultOverwritten(z, bi)
			require.BlockInstance.Forwarded(z, bi)
			require.BlockInstance.Default.ForwardsReceivedAttributes(z, bi)
		}
	}
}

func (requirer) ComponentCall_ForwardsAndPlaceholder(z *analyzer, cc *file.ComponentCall) {
	if !cc.Analyzed {
		require.ComponentCall.ForwardsReceivedAndPlaceholder(z, cc)
		require.ComponentCall.BlockSetterForwardsAndPlaceholder(z, cc)
	}
}

func (requirer) ComponentCall_ForwardsReceivedAndPlaceholder(z *analyzer, cc *file.ComponentCall) {
	if !cc.Analyzed {
		require.ComponentCall.ForwardsAcceptedAttributes(z, cc)
		require.ComponentCall.ReceivesAndPlaceholder(z, cc)
	}
}

func (requirer) ComponentCall_BlockSetterForwardsAndPlaceholder(z *analyzer, cc *file.ComponentCall) {
	if !assert.DebugEnabled || cc.Analyzed {
		return
	}
	for _, s := range cc.BlockSetters {
		require.BlockSetter.ForwardsAndPlaceholder(z, s)
		require.Block.Forwarded(z, s.Block)
	}
}

func (requirer) ComponentCall_ForwardsAttributes(z *analyzer, cc *file.ComponentCall) {
	if !cc.Analyzed {
		require.ComponentCall.ComponentForwardsAttributes(z, cc)
		require.ComponentCall.ForwardsReceivedAttributes(z, cc)
		require.ComponentCall.BlockSetterForwardsAttributes(z, cc)
	}
}

func (requirer) ComponentCall_ComponentForwardsAttributes(z *analyzer, cc *file.ComponentCall) {
	if !assert.DebugEnabled || cc.Analyzed {
		return
	}
	require.Component.AlwaysForwardsAttributes(z, cc.Component)
	for _, b := range cc.Component.Blocks {
		for _, bi := range b.Instances {
			if bi.Default == nil {
				continue
			}
			require.BlockInstance.DefaultOverwritten(z, bi)
			require.BlockInstance.Default.ForwardsAttributes(z, bi)
		}
	}
}

func (requirer) ComponentCall_ForwardsReceivedAttributes(z *analyzer, cc *file.ComponentCall) {
	if !cc.Analyzed {
		require.ComponentCall.ForwardsAcceptedAttributes(z, cc)
		require.ComponentCall.ReceivesAttributes(z, cc)
	}
}

func (requirer) ComponentCall_BlockSetterForwardsAttributes(z *analyzer, cc *file.ComponentCall) {
	if !assert.DebugEnabled || cc.Analyzed {
		return
	}
	for _, s := range cc.BlockSetters {
		require.BlockSetter.ForwardsAttributes(z, s)
		require.Block.Forwarded(z, s.Block)
	}
}

func (requirer) ComponentCall_WritesContent(z *analyzer, cc *file.ComponentCall) {
	if !cc.Analyzed {
		require.ComponentCall.ComponentWritesContent(z, cc)
		require.ComponentCall.BlockSetterWritesContent(z, cc)
	}
}

func (requirer) ComponentCall_ComponentWritesContent(z *analyzer, cc *file.ComponentCall) {
	if !assert.DebugEnabled || cc.Analyzed {
		return
	}
	require.Component.AlwaysWritesContent(z, cc.Component)
	for _, b := range cc.Component.Blocks {
		for _, bi := range b.Instances {
			if bi.Default == nil {
				continue
			}
			require.BlockInstance.DefaultOverwritten(z, bi)
			require.BlockInstance.Default.WritesContent(z, bi)
		}
	}
}

func (requirer) ComponentCall_BlockSetterWritesContent(z *analyzer, cc *file.ComponentCall) {
	if !assert.DebugEnabled || cc.Analyzed {
		return
	}
	for _, s := range cc.BlockSetters {
		require.BlockSetter.WritesContent(z, s)
	}
}

func (requirer) ComponentCall_WritesElements(z *analyzer, cc *file.ComponentCall) {
	if !cc.Analyzed {
		require.ComponentCall.ComponentWritesElements(z, cc)
		require.ComponentCall.BlockSetterWritesElements(z, cc)
	}
}

func (requirer) ComponentCall_ComponentWritesElements(z *analyzer, cc *file.ComponentCall) {
	if !assert.DebugEnabled || cc.Analyzed {
		return
	}
	require.Component.AlwaysWritesElements(z, cc.Component)
	for _, b := range cc.Component.Blocks {
		for _, bi := range b.Instances {
			if bi.Default == nil {
				continue
			}
			require.BlockInstance.DefaultOverwritten(z, bi)
			require.BlockInstance.Default.WritesElements(z, bi)
		}
	}
}

func (requirer) ComponentCall_BlockSetterWritesElements(z *analyzer, cc *file.ComponentCall) {
	if !assert.DebugEnabled || cc.Analyzed {
		return
	}
	for _, s := range cc.BlockSetters {
		require.BlockSetter.WritesElements(z, s)
	}
}

// =========================================================================
// Element Spec
// ======================================================================================

func (requirer) ElementSpec_Circular(z *analyzer, es *file.ElementSpec) {
	if !es.Analyzed {
		requireField.ElementSpec.Circular(z, es)
	}
}

func (requirer) ElementSpec_Type(z *analyzer, es *file.ElementSpec) {
	if !es.Analyzed {
		requireField.ElementSpec.Type(z, es)
	}
}
