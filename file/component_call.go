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
	// If the reason is a [ast.ComponentCall], then because that call
	// also receives attributes and forwards them, because that call's
	// component forwards attributes, or because one of the call's block
	// setters forwards attributes
	// ([ComponentCall.BlockSetterForwardsAttributes]).
	ReceivesAttributes AnalysisWithReason[ast.AttributeWriter]
	// ReceivesAndPlaceholder indicates that the call's &-placeholder gets
	// filled with the &-placeholder of the component containing this call.
	//
	// If the reason is a [ast.ComponentCall], then because that call
	// receives an &-placeholder and forwards its received attributes.
	ReceivesAndPlaceholder AnalysisWithReason[ast.AndPlaceholderWriter]

	// ForwardsReceivedAttributes indicates whether the component call
	// forwards attributes the attributes it receives through a top-level
	// &-placeholder to the element containing the component call again.
	//
	// The reason is a node from the body of the _component_ (not the call).
	//
	// If the reason is a [ast.ComponentCall], then because that call also
	// forwards received attributes and receives an &-placeholder.
	ForwardsReceivedAttributes AnalysisWithReason[ast.AndPlaceholderWriter]
	// AcceptsAttributes indicates whether the call's component accepts
	// attributes.
	//
	// ForwardsReceivedAttributes implies AcceptsAttributes.
	//
	// The reason is a node from the body of the _component_ (not the call).
	//
	// If the reason is a [ast.ComponentCall], then because that call also
	// accepts attributes and receives an &-placeholder.
	AcceptsAttributes AnalysisWithReason[ast.AndPlaceholderWriter]

	// ComponentForwardsAttributes indicates whether the component being called
	// forwards attributes, disregarding block setters, but considering block
	// defaults.
	//
	// The reason is a node from the body of the _component_ (not the call).
	//
	// If the reason is a [ast.ComponentCall], then because either that call's
	// component also forwards attributes, because one of the call's block
	// setters forwards attributes
	// ([ComponentCall.BlockSetterForwardsAttributes]), or because the call
	// receives attributes and forwards them.
	ComponentForwardsAttributes AnalysisWithReason[ast.AttributeWriter]

	// ComponentWritesContent indicates whether the component being called
	// writes content, disregarding block setters, but considering block
	// defaults.
	//
	// The reason is a node from the body of the _component_ (not the call).
	//
	// If the reason is a [ast.ComponentCall], then because either that call's
	// component also writes content, or because one of the call's block
	// setters writes content ([ComponentCall.BlockSetterWritesContent]).
	ComponentWritesContent AnalysisWithReason[ast.ContentWriter]
	// ComponentWritesElements indicates whether the component being called
	// writes elements, disregarding block setters, but considering block
	// defaults.
	//
	// If the reason is a [ast.ComponentCall], then because either that call's
	// component also writes elements, or because one of the call's block
	// setters writes elements ([ComponentCall.BlockSetterWritesElements]).
	ComponentWritesElements AnalysisWithReason[ast.ElementWriter]
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

// BlockSetterForwardsAndPlaceholder indicates that at least one block setter
// forwards its &-placeholder out of the component call.
func (cc *ComponentCall) BlockSetterForwardsAndPlaceholder() (a AnalysisWithReason[*BlockSetter]) {
	a.SetFalse()

	for _, s := range cc.BlockSetters {
		fap := s.ForwardsAndPlaceholder()
		if s.Block == nil {
			if fap.Failed() || fap.True() {
				a.SetFailed()
			}
			continue
		}

		fap = ConditionalAnalysis(s.Block.Forwarded, fap)
		if fap.True() {
			a.SetReason(s)
			return a
		} else if fap.Failed() {
			a.SetFailed()
		}
	}
	return a
}

// BlockSetterForwardsAttributes indicates that at least one block setter
// forwards attributes out of the component call.
func (cc *ComponentCall) BlockSetterForwardsAttributes() (a AnalysisWithReason[*BlockSetter]) {
	a.SetFalse()

	for _, s := range cc.BlockSetters {
		fa := s.ForwardsAttributes()
		if s.Block == nil {
			if fa.Failed() || fa.True() {
				a.SetFailed()
			}
			continue
		}

		fa = ConditionalAnalysis(s.Block.Forwarded, fa)
		if fa.True() {
			a.SetReason(s)
			return a
		} else if fa.Failed() {
			a.SetFailed()
		}
	}
	return a
}

func (cc *ComponentCall) BlockSetterWritesContent() (a AnalysisWithReason[*BlockSetter]) {
	a.SetFalse()

	for _, s := range cc.BlockSetters {
		wc := s.WritesContent()
		if wc.True() {
			a.SetReason(s)
			return a
		} else if wc.Failed() {
			a.SetFailed()
		}
	}
	return a
}

func (cc *ComponentCall) BlockSetterWritesElements() (a AnalysisWithReason[*BlockSetter]) {
	a.SetFalse()

	for _, s := range cc.BlockSetters {
		we := s.WritesElements()
		if we.True() {
			a.SetReason(s)
			return a
		} else if we.Failed() {
			a.SetFailed()
		}
	}

	return a
}

func (cc *ComponentCall) ForwardsAndPlaceholder() (a Analysis[bool]) {
	a.SetResult(false)

	forwardsReceivedAndPlaceholder := ConditionalAnalysis(cc.ForwardsReceivedAttributes, cc.ReceivesAndPlaceholder)
	if forwardsReceivedAndPlaceholder.True() {
		a.SetResult(true)
		return a
	} else if forwardsReceivedAndPlaceholder.Failed() {
		a.SetFailed()
	}

	bsfa := cc.BlockSetterForwardsAndPlaceholder()
	if bsfa.True() {
		a.SetResult(true)
		return a
	} else if bsfa.Failed() {
		a.SetFailed()
	}

	return a
}

func (cc *ComponentCall) ForwardsAttributes() (a Analysis[bool]) {
	a.SetResult(false)
	if cc.ComponentForwardsAttributes.True() {
		a.SetResult(true)
		return a
	} else if cc.ComponentForwardsAttributes.Failed() {
		a.SetFailed()
	}

	forwardsReceivedAttrs := ConditionalAnalysis(cc.ForwardsReceivedAttributes, cc.ReceivesAttributes)
	if forwardsReceivedAttrs.True() {
		a.SetResult(true)
		return a
	} else if forwardsReceivedAttrs.Failed() {
		a.SetFailed()
	}

	bsfa := cc.BlockSetterForwardsAttributes()
	if bsfa.True() {
		a.SetResult(true)
		return a
	} else if bsfa.Failed() {
		a.SetFailed()
	}

	return a
}

func (cc *ComponentCall) WritesContent() (a Analysis[bool]) {
	a.SetResult(false)

	if cc.ComponentWritesContent.True() {
		a.SetResult(true)
		return a
	} else if cc.ComponentWritesContent.Failed() {
		a.SetFailed()
	}

	bswc := cc.BlockSetterWritesContent()
	if bswc.True() {
		a.SetResult(true)
		return a
	} else if bswc.Failed() {
		a.SetFailed()
	}

	return a
}

func (cc *ComponentCall) WritesElements() (a Analysis[bool]) {
	a.SetResult(false)

	if cc.ComponentWritesElements.True() {
		a.SetResult(true)
		return a
	} else if cc.ComponentWritesElements.Failed() {
		a.SetFailed()
	}

	bswe := cc.BlockSetterWritesElements()
	if bswe.True() {
		a.SetResult(true)
		return a
	} else if bswe.Failed() {
		a.SetFailed()
	}

	return a
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
	a.SetFalse()
	for _, instance := range s.Instances {
		if instance.WritesAndPlaceholder.True() {
			a.SetReason(instance)
			return a
		} else if instance.WritesAndPlaceholder.Failed() {
			a.SetFailed()
		}
	}
	return a
}

// ForwardsAndPlaceholder indicates that any of the block setter instances
// forward an &-placeholder.
//
// The reason is that BlockSetterInstance.
func (s *BlockSetter) ForwardsAndPlaceholder() (a AnalysisWithReason[*BlockSetterInstance]) {
	a.SetFalse()
	for _, instance := range s.Instances {
		if instance.ForwardsAndPlaceholder.True() {
			a.SetReason(instance)
			return a
		} else if instance.ForwardsAndPlaceholder.Failed() {
			a.SetFailed()
		}
	}
	return a
}

// ForwardsAttributes indicates that any of the block setter instances
// forward attributes.
//
// The reason is that BlockSetterInstance.
func (s *BlockSetter) ForwardsAttributes() (a AnalysisWithReason[*BlockSetterInstance]) {
	a.SetFalse()
	for _, instance := range s.Instances {
		if instance.ForwardsAttributes.True() {
			a.SetReason(instance)
			return a
		} else if instance.ForwardsAttributes.Failed() {
			a.SetFailed()
		}
	}
	return a
}

// WritesContent indicates that any of the block setter instances write content.
//
// The reason is that BlockSetterInstance.
func (s *BlockSetter) WritesContent() (a AnalysisWithReason[*BlockSetterInstance]) {
	a.SetFalse()
	for _, instance := range s.Instances {
		if instance.WritesContent.True() {
			a.SetReason(instance)
			return a
		} else if instance.WritesContent.Failed() {
			a.SetFailed()
		}
	}
	return a
}

// WritesElements indicates that any of the block setter instances write
// elements.
//
// The reason is that BlockSetterInstance.
func (s *BlockSetter) WritesElements() (a AnalysisWithReason[*BlockSetterInstance]) {
	a.SetFalse()
	for _, instance := range s.Instances {
		if instance.WritesElements.True() {
			a.SetReason(instance)
		} else if instance.WritesElements.Failed() {
			a.SetFailed()
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
