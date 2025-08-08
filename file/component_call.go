package file

import "github.com/mavolin/corgi/v2/file/ast"

type ComponentCall struct {
	//
	// BUILD SYMBOLS

	AST *ast.ComponentCall

	// File is the file the Component is defined in.
	File *File

	// BlockSetters are the block setters used in this component call.
	//
	// All block setters and their instances are guaranteed to be correctly set after
	// analyzing, even if [AnalyzedWithErrors] is true.
	BlockSetters []*BlockSetter

	//
	// LINKER

	// Linked indicates whether this ComponentCall has been seen by the linker,
	// and it attempted to link it.
	//
	// If this is true, but [Component] is nil, the linker encountered an error
	// while linking the ComponentCall.
	Linked bool

	// Component is the Component being called.
	Component *Component

	//
	// ANALYZER

	// Analyzed indicates whether the ComponentCall has been analyzed,
	// albeit with errors.
	Analyzed bool

	// Circular indicates this call's component calls itself.
	// In other words, this component call is part of a recursion.
	Circular bool

	// ReceivesAttributes indicates that the component's &-placeholder gets
	// filled.
	//
	// The reason is either directly set to an attribute, or set to a component
	// call.
	// In case of the latter, the attribute writer causing the
	// delegation is the reason of why the component call ForwardsAttributes.
	//
	// Note that if the component call and the component call's
	// reason for ForwardsAttributes are the same, the component call
	// itself is the one writing attributes, as opposed to a block setter.
	// Refer to the documentation of ForwardsAttributes for more information.
	ReceivesAttributes AnalysisWithReason[ast.AttributeWriter]
	// ReceivesAndPlaceholder indicates that the component's &-placeholder gets
	// filled with the &-placeholder of the calling component.
	//
	// The reason is either directly set to an &-placeholder, or set to a
	// component call.
	// In case of the latter, the &-placeholder writer causing the delegation
	// is the reason why that call ForwardsAndPlaceholder.
	//
	// Note that if the component call and the component call's
	// ForwardsAndPlaceholder are the same, the component call
	// itself is the one filling the &-placeholder, as opposed to an &.
	// Refer to the documentation of ForwardsAndPlaceholder for more
	// information.
	ReceivesAndPlaceholder AnalysisWithReason[ast.AndPlaceholderWriter]

	// ForwardsReceivedAttributes indicates whether the component call
	// forwards attributes the attributes it receives through a top-level
	// &-placeholder to the element containing the component call again.
	//
	// In other words, whether the component call would output the attributes
	// it receives (that are delegated to it) at its top-level.
	//
	// The most simple example of that is:
	//    comp Foo() { &(&) }
	// Where Foo is called with a delegated attribute.
	//
	// The reason is the &-placeholder writer of the _component_ (not the call)
	// that forwards the attributes.
	ForwardsReceivedAttributes AnalysisWithReason[ast.AndPlaceholderWriter]
	// AcceptsAttributes indicates whether the call's component accepts
	// attributes.
	//
	// ForwardsReceivedAttributes implies AcceptsAttributes.
	//
	// The reason is the &-placeholder writer of the _component_ (not the call)
	// that accepts the attributes.
	AcceptsAttributes AnalysisWithReason[ast.AndPlaceholderWriter]

	// ForwardsAttributes indicates whether the component call forwards
	// attributes.
	//
	// For attribute writers in the body of this component call, i.e.
	// those passed to the component's top-level &-placeholder, or those in
	// top-level block setters, it is set to that attribute writer directly.
	//
	// For attribute writers in the component's body, it is set to the
	// component call itself.
	//
	// Values referencing the component call itself are preferred.
	//
	// This also means, if this is not zero, but not set to the component call
	// itself, the only place adding top-level attributes is the
	// component call itself, through block setters or the component's
	// top-level &-placeholder.
	ForwardsAttributes AnalysisWithReason[ast.AttributeWriter]
	// ForwardsAndPlaceholder indicates whether the component call forwards
	// the &-placeholder it receives.
	//
	// This fields considers &-placeholder writers in forwarded block setters
	// and &-placeholders given directly to the component.
	ForwardsAndPlaceholder AnalysisWithReason[ast.AndPlaceholderWriter]

	// WritesContent indicates whether the component call writes content.
	//
	// The reason is the content writer inside the component call, or the
	// component call itself if the called component writes content.
	// The component call itself is preferred.
	WritesContent AnalysisWithReason[ast.ContentWriter]
}

func (cc *ComponentCall) External() bool { return cc.File.Package != cc.Component.File.Package }
func (cc *ComponentCall) Local() bool    { return !cc.External() }

func (cc *ComponentCall) BlockSetterByName(name string) *BlockSetter {
	for _, s := range cc.BlockSetters {
		if s.Name == name {
			return s
		}
	}
	return nil
}

func (cc *ComponentCall) BlockSetterByNode(n ast.BlockSetter) *BlockSetter {
	w := cc.BlockSetterByName(n.Name())
	if w == nil {
		return nil
	}

	if w.InstanceByNode(n) != nil {
		return w
	}
	return nil
}

// ReceivedAttributeChain returns the chain of component calls
// that are responsible for the attributes this component receives.
func (cc *ComponentCall) ReceivedAttributeChain() []ast.AttributeWriter {
	if cc.ReceivesAttributes.False() {
		return nil
	}

	curAST, _ := cc.ReceivesAttributes.Reason().(*ast.ComponentCall)
	if curAST == nil {
		return []ast.AttributeWriter{cc.ReceivesAttributes.Reason()}
	}

	var chain []ast.AttributeWriter
	for {
		chain = append(chain, curAST)

		prevAST := curAST
		prev := cc.File.ComponentCallByNode(curAST)
		curAST, _ = prev.ForwardsAttributes.Reason().(*ast.ComponentCall)
		if curAST == nil || curAST == prevAST {
			return chain
		}
	}
}

// ReceivedAndPlaceholderChain returns the chain of component calls
// that are responsible for the &-placeholder this component receives.
func (cc *ComponentCall) ReceivedAndPlaceholderChain() []ast.AndPlaceholderWriter {
	if cc.ReceivesAndPlaceholder.False() {
		return nil
	}

	curAST, _ := cc.ReceivesAndPlaceholder.Reason().(*ast.ComponentCall)
	if curAST == nil {
		return []ast.AndPlaceholderWriter{cc.ReceivesAndPlaceholder.Reason()}
	}

	var chain []ast.AndPlaceholderWriter
	for {
		chain = append(chain, curAST)

		prevAST := curAST
		prev := cc.File.ComponentCallByNode(curAST)
		curAST, _ = prev.ForwardsAndPlaceholder.Reason().(*ast.ComponentCall)
		if curAST == nil || curAST == prevAST {
			return chain
		}
	}
}

type BlockSetter struct {
	//
	// BUILD SYMBOLS

	Name      string
	Instances []*BlockSetterInstance

	//
	// LINKER

	// Linked indicates whether this BlockSetter has been seen by the
	// linker, and it attempted to link it.
	//
	// If this is true, but [Block] is nil, the linker encountered an error
	// while linking the BlockSetter.
	Linked bool

	Block *Block
}

func (s *BlockSetter) InstanceByNode(n ast.BlockSetter) *BlockSetterInstance {
	for _, instance := range s.Instances {
		if instance.AST == n {
			return instance
		}
	}
	return nil
}

// WritesAndPlaceholder indicates that any of the block setter instances
// write an &-placeholder.
//
// The reason is that BlockSetterInstance.
func (s *BlockSetter) WritesAndPlaceholder() (a AnalysisWithReason[*BlockSetterInstance]) {
	a.SetReason(nil)
	for _, instance := range s.Instances {
		ap := instance.WritesAndPlaceholder
		if ap.Failed() {
			a.SetFailed()
		} else if ap.Reason() != nil {
			a.SetReason(instance)
			return a
		}
	}
	return a
}

// ForwardsAndPlaceholder indicates that any of the block setter instances
// forward an &-placeholder.
//
// The reason is that BlockSetterInstance.
func (s *BlockSetter) ForwardsAndPlaceholder() (a AnalysisWithReason[*BlockSetterInstance]) {
	a.SetReason(nil)
	for _, instance := range s.Instances {
		ap := instance.ForwardsAndPlaceholder
		if ap.Failed() {
			a.SetFailed()
		} else if ap.Reason() != nil {
			a.SetReason(instance)
			return a
		}
	}
	return a
}

// ForwardsAttributes indicates that any of the block setter instances
// forward attributes.
//
// The reason is that BlockSetterInstance.
func (s *BlockSetter) ForwardsAttributes() (a AnalysisWithReason[*BlockSetterInstance]) {
	a.SetReason(nil)
	for _, instance := range s.Instances {
		aw := instance.ForwardsAttributes
		if aw.Failed() {
			a.SetFailed()
		} else if aw.Reason() != nil {
			a.SetReason(instance)
			return a
		}
	}
	return a
}

// WritesContent indicates that any of the block setter instances write content.
//
// The reason is that BlockSetterInstance.
func (s *BlockSetter) WritesContent() (a AnalysisWithReason[*BlockSetterInstance]) {
	a.SetReason(nil)
	for _, instance := range s.Instances {
		cw := instance.WritesContent
		if cw.Failed() {
			a.SetFailed()
		} else if cw.Reason() != nil {
			a.SetReason(instance)
			return a
		}
	}
	return a
}

// WritesElements indicates that any of the block setter instances write
// elements.
//
// The reason is that BlockSetterInstance.
func (s *BlockSetter) WritesElements() (a AnalysisWithReason[*BlockSetterInstance]) {
	a.SetReason(nil)
	for _, instance := range s.Instances {
		ew := instance.WritesElements
		if ew.Failed() {
			a.SetFailed()
		} else if ew.Reason() != nil {
			a.SetReason(instance)
		}
	}
	return a
}

type BlockSetterInstance struct {
	Group *BlockSetter
	AST   ast.BlockSetter

	WritesAndPlaceholder   AnalysisWithReason[ast.AndPlaceholderWriter]
	ForwardsAndPlaceholder AnalysisWithReason[ast.AndPlaceholderWriter]
	ForwardsAttributes     AnalysisWithReason[ast.AttributeWriter]
	WritesContent          AnalysisWithReason[ast.ContentWriter]
	WritesElements         AnalysisWithReason[ast.ElementWriter]
}
