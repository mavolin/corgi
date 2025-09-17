package file

import "github.com/mavolin/corgi/v2/file/ast"

type BlockSetter struct {
	//
	// BUILD SYMBOLS

	Name      Identifier
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
