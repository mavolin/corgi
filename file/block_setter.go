package file

import "github.com/mavolin/corgi/v2/file/ast"

type (
	BlockSetter struct {
		//
		// BUILD SYMBOLS

		// ComponentCall is the component call this block setter belongs to.
		ComponentCall *ComponentCall

		Name      Identifier
		Instances []*BlockSetterInstance

		//
		// LINKER

		Block *Block
	}

	BlockSetterInstance struct {
		//
		// BUILD SYMBOLS

		Group *BlockSetter
		AST   ast.BlockSetter

		//
		// ANALYZER

		AcceptsAttributes          AnalysisWithReason[ast.AndPlaceholderWriter]
		ForwardsReceivedAttributes AnalysisWithReason[ast.AndPlaceholderWriter]
		ForwardsAttributes         AnalysisWithReason[ast.AttributeWriter]
		WritesContent              AnalysisWithReason[ast.ContentWriter]
		WritesElements             AnalysisWithReason[ast.ElementWriter]
	}
)

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
		if instance.AcceptsAttributes.True() {
			a.SetReason(instance)
			return a
		} else if instance.AcceptsAttributes.Failed() {
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
		if instance.ForwardsReceivedAttributes.True() {
			a.SetReason(instance)
			return a
		} else if instance.ForwardsReceivedAttributes.Failed() {
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
