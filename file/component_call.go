package file

import "github.com/mavolin/corgi/v2/file/ast"

type ComponentCall struct {
	//
	// BUILD SYMBOLS

	AST *ast.ComponentCall

	// File is the file the Component is defined in.
	File *File

	// BlockSetters are the withs used in this component call.
	//
	// All withs and their instances are guaranteed to be correctly set after
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

	// FirstAnd is the first & that fills the components placeholder.
	// Nil, if no such & exists.
	FirstAnd Analysis[*ast.And]
}

func (cc *ComponentCall) External() bool { return cc.File.Package != cc.Component.File.Package }
func (cc *ComponentCall) Local() bool    { return !cc.External() }

func (cc *ComponentCall) BlockSetterByName(name string) *BlockSetter {
	for _, with := range cc.BlockSetters {
		if with.Name == name {
			return with
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

func (w *BlockSetter) InstanceByNode(n ast.BlockSetter) *BlockSetterInstance {
	for _, instance := range w.Instances {
		if instance.AST == n {
			return instance
		}
	}
	return nil
}

type BlockSetterInstance struct {
	Group *BlockSetter
	AST   ast.BlockSetter
}
