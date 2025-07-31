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

	// Component is the Component being called.
	Component *Component

	//
	// ANALYZER

	AnalyzedWithErrors bool

	// FirstAnd is the first & that fills the components placeholder.
	// Nil, if no such & exists.
	FirstAnd *ast.And
}

func (cc *ComponentCall) External() bool {
	return cc.File.Package != cc.Component.File.Package
}

func (cc *ComponentCall) Local() bool {
	return !cc.External()
}

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
	// ANALYZER

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
