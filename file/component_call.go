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

	// FirstDelegatedAttributeWriter is the first attribute writer filling
	// the &-placeholder of the called component.
	//
	// It is either directly set to an attribute, or set to a component call.
	// In case of the latter, the attribute writer causing the
	// delegation is the FirstForwardedAttributeWriter of that component call.
	//
	// Note that if the component call and the component call's
	// FirstForwardedAttributeWriter are the same, the component call
	// itself is the one writing attributes, as opposed to a block setter.
	// Refer to the documentation of FirstForwardedAttributeWriter for more
	// information.
	FirstDelegatedAttributeWriter Analysis[ast.AttributeWriter]
	// FirstDelegatedAndPlaceholderWriter is the first &-placeholder writer
	// filling the &-placeholder of the called component.
	//
	// It is either directly set to an &-placeholder, or set to a component call.
	// In case of the latter, the &-placeholder writer causing the delegation
	// is the FirstForwardedAndPlaceholderWriter of that component call.
	//
	// Note that if the component call and the component call's
	// FirstForwardedAndPlaceholderWriter are the same, the component call
	// itself is the one filling the &-placeholder, as opposed to an &.
	// Refer to the documentation of FirstForwardedAndPlaceholderWriter for more
	// information.
	FirstDelegatedAndPlaceholderWriter Analysis[ast.AndPlaceholderWriter]
	// ForwardsDelegatedAttributes indicates whether the component call
	// forwards attributes the attributes it receives through a top-level
	// &-placeholder to the element containing the component call again.
	//
	// In other words, whether the component call would output the attributes
	// it receives (that are delegated to it) at its top-level.
	//
	// The most simple example of that is:
	//    comp Foo() { &(&) }
	// Where Foo is called with a delegated attribute.
	ForwardsDelegatedAttributes Analysis[bool]
	// AcceptsAttributes indicates whether the call's component accepts
	// attributes.
	//
	// ForwardsDelegatedAttributes implies AcceptsAttributes.
	AcceptsAttributes Analysis[bool]

	// FirstForwardedAttributeWriter is the first attribute writer producing
	// top-level attributes.
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
	FirstForwardedAttributeWriter Analysis[ast.AttributeWriter]
	// FirstForwardedAndPlaceholderWriter is the first &-placeholder writer producing
	// top-level attributes if filled.
	//
	// This fields considers &-placeholder writers in top-level block setters
	// and &-placeholders delegated to the component.
	FirstForwardedAndPlaceholderWriter Analysis[ast.AndPlaceholderWriter]
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

// FirstDelegatedAttributeWriterChain returns the chain of component calls
// that are responsible for the first delegated attribute writer of this
// component call.
func (cc *ComponentCall) FirstDelegatedAttributeWriterChain() []ast.AttributeWriter {
	if cc.FirstDelegatedAttributeWriter.Equal(nil) {
		return nil
	}

	curAST, _ := cc.FirstDelegatedAttributeWriter.Result.(*ast.ComponentCall)
	if curAST == nil {
		return []ast.AttributeWriter{cc.FirstDelegatedAttributeWriter.Result}
	}

	var chain []ast.AttributeWriter
	for {
		chain = append(chain, curAST)

		prevAST := curAST
		prev := cc.File.ComponentCallByNode(curAST)
		curAST, _ = prev.FirstForwardedAttributeWriter.Result.(*ast.ComponentCall)
		if curAST == nil || curAST == prevAST {
			return chain
		}
	}
}

// FirstDelegatedAndPlaceholderWriterChain returns the chain of component calls
// that are responsible for the first delegated &-placeholder writer of this
// component call.
func (cc *ComponentCall) FirstDelegatedAndPlaceholderWriterChain() []ast.AndPlaceholderWriter {
	if cc.FirstDelegatedAndPlaceholderWriter.Equal(nil) {
		return nil
	}

	curAST, _ := cc.FirstDelegatedAndPlaceholderWriter.Result.(*ast.ComponentCall)
	if curAST == nil {
		return []ast.AndPlaceholderWriter{cc.FirstDelegatedAndPlaceholderWriter.Result}
	}

	var chain []ast.AndPlaceholderWriter
	for {
		chain = append(chain, curAST)

		prevAST := curAST
		prev := cc.File.ComponentCallByNode(curAST)
		curAST, _ = prev.FirstForwardedAndPlaceholderWriter.Result.(*ast.ComponentCall)
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

// FirstAndPlaceholderWriter returns the first &-placeholder in any of the block
// setter's instances.
func (s *BlockSetter) FirstAndPlaceholderWriter() Analysis[*BlockSetterInstance] {
	var failed bool
	for _, instance := range s.Instances {
		ap := instance.FirstAndPlaceholderWriter
		if ap.Failed {
			failed = true
		} else if ap.Result != nil {
			return Result(instance)
		}
	}
	return ResultIf[*BlockSetterInstance](nil, !failed)
}

// FirstForwardedAndPlaceholderWriter returns the first top-level &-placeholder in any
// of the block setter's instances.
func (s *BlockSetter) FirstForwardedAndPlaceholderWriter() Analysis[*BlockSetterInstance] {
	var failed bool
	for _, instance := range s.Instances {
		ap := instance.FirstForwardedAndPlaceholderWriter
		if ap.Failed {
			failed = true
		} else if ap.Result != nil {
			return Result(instance)
		}
	}
	return ResultIf[*BlockSetterInstance](nil, !failed)
}

// FirstForwardedAttributeWriter returns the first top-level attribute writer in
// any of the block setter's instances.
func (s *BlockSetter) FirstForwardedAttributeWriter() Analysis[*BlockSetterInstance] {
	var failed bool
	for _, instance := range s.Instances {
		aw := instance.FirstForwardedAttributeWriter
		if aw.Failed {
			failed = true
		} else if aw.Result != nil {
			return Result(instance)
		}
	}
	return ResultIf[*BlockSetterInstance](nil, !failed)
}

// FirstContentWriter returns the first content writer in any of the block
// setter's instances.
func (s *BlockSetter) FirstContentWriter() Analysis[*BlockSetterInstance] {
	var failed bool
	for _, instance := range s.Instances {
		cw := instance.FirstContentWriter
		if cw.Failed {
			failed = true
		} else if cw.Result != nil {
			return Result(instance)
		}
	}
	return ResultIf[*BlockSetterInstance](nil, !failed)
}

// FirstElementWriter returns the first element writer in any of the block
// setter's instances.
func (s *BlockSetter) FirstElementWriter() Analysis[*BlockSetterInstance] {
	var failed bool
	for _, instance := range s.Instances {
		ew := instance.FirstElementWriter
		if ew.Failed {
			failed = true
		} else if ew.Result != nil {
			return Result(instance)
		}
	}
	return ResultIf[*BlockSetterInstance](nil, !failed)
}

type BlockSetterInstance struct {
	Group *BlockSetter
	AST   ast.BlockSetter

	FirstAndPlaceholderWriter          Analysis[ast.AndPlaceholderWriter]
	FirstForwardedAndPlaceholderWriter Analysis[ast.AndPlaceholderWriter]

	FirstForwardedAttributeWriter Analysis[ast.AttributeWriter]
	FirstContentWriter            Analysis[ast.ContentWriter]
	FirstElementWriter            Analysis[ast.ElementWriter]
}
