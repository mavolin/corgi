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
	// the &-placeholder of the called component outside a block setter.
	FirstDelegatedAttributeWriter Analysis[ast.AttributeWriter]
	// FirstDelegatedContentWriter is the first content writer filling
	// the &-placeholder of the called component outside a block setter.
	FirstDelegatedAndPlaceholder Analysis[*ast.AndPlaceholder]
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

// FirstBlockAndPlaceholder returns the first block setter instance that has a
// &-placeholder in any of its instances.
func (cc *ComponentCall) FirstBlockAndPlaceholder() Analysis[*BlockSetterInstance] {
	var failed bool
	for _, s := range cc.BlockSetters {
		ap := s.FirstAndPlaceholder()
		if ap.Failed {
			failed = true
			continue
		} else if ap.Result != nil {
			return ap
		}
	}
	return ResultIf[*BlockSetterInstance](nil, !failed)
}

// FirstBlockTopLevelAndPlaceholder returns the first block setter instance that
// has a top-level &-placeholder in any of its instances.
func (cc *ComponentCall) FirstBlockTopLevelAndPlaceholder() Analysis[*BlockSetterInstance] {
	var failed bool
	for _, s := range cc.BlockSetters {
		ap := s.FirstTopLevelAndPlaceholder()
		if ap.Failed {
			failed = true
			continue
		} else if ap.Result != nil {
			return ap
		}
	}
	return ResultIf[*BlockSetterInstance](nil, !failed)
}

// FirstBlockTopLevelAttributeWriter returns the first block setter instance
// that has a top-level attribute writer in any of its instances.
func (cc *ComponentCall) FirstBlockTopLevelAttributeWriter() Analysis[*BlockSetterInstance] {
	var failed bool
	for _, s := range cc.BlockSetters {
		aw := s.FirstTopLevelAttributeWriter()
		if aw.Failed {
			failed = true
			continue
		} else if aw.Result != nil {
			return aw
		}
	}
	return ResultIf[*BlockSetterInstance](nil, !failed)
}

// FirstBlockContentWriter returns the first block setter instance that has a
// content writer in any of its instances.
func (cc *ComponentCall) FirstBlockContentWriter() Analysis[*BlockSetterInstance] {
	var failed bool
	for _, s := range cc.BlockSetters {
		cw := s.FirstContentWriter()
		if cw.Failed {
			failed = true
			continue
		} else if cw.Result != nil {
			return cw
		}
	}
	return ResultIf[*BlockSetterInstance](nil, !failed)
}

// FirstBlockElementWriter returns the first block setter instance that has an
// element writer in any of its instances.
func (cc *ComponentCall) FirstBlockElementWriter() Analysis[*BlockSetterInstance] {
	var failed bool
	for _, s := range cc.BlockSetters {
		ew := s.FirstElementWriter()
		if ew.Failed {
			failed = true
			continue
		} else if ew.Result != nil {
			return ew
		}
	}
	return ResultIf[*BlockSetterInstance](nil, !failed)
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

// FirstAndPlaceholder returns the first &-placeholder in any of the block
// setter's instances.
func (s *BlockSetter) FirstAndPlaceholder() Analysis[*BlockSetterInstance] {
	var failed bool
	for _, instance := range s.Instances {
		ap := instance.FirstAndPlaceholder
		if ap.Failed {
			failed = true
		} else if ap.Result != nil {
			return Result(instance)
		}
	}
	return ResultIf[*BlockSetterInstance](nil, !failed)
}

// FirstTopLevelAndPlaceholder returns the first top-level &-placeholder in any
// of the block setter's instances.
func (s *BlockSetter) FirstTopLevelAndPlaceholder() Analysis[*BlockSetterInstance] {
	var failed bool
	for _, instance := range s.Instances {
		ap := instance.FirstTopLevelAndPlaceholder
		if ap.Failed {
			failed = true
		} else if ap.Result != nil {
			return Result(instance)
		}
	}
	return ResultIf[*BlockSetterInstance](nil, !failed)
}

// FirstTopLevelAttributeWriter returns the first top-level attribute writer in
// any of the block setter's instances.
func (s *BlockSetter) FirstTopLevelAttributeWriter() Analysis[*BlockSetterInstance] {
	var failed bool
	for _, instance := range s.Instances {
		aw := instance.FirstTopLevelAttributeWriter
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

	FirstAndPlaceholder         Analysis[*ast.AndPlaceholder]
	FirstTopLevelAndPlaceholder Analysis[*ast.AndPlaceholder]

	FirstTopLevelAttributeWriter Analysis[ast.AttributeWriter]
	FirstContentWriter           Analysis[ast.ContentWriter]
	FirstElementWriter           Analysis[ast.ElementWriter]
}
