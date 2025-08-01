package file

import "github.com/mavolin/corgi/v2/file/ast"

// Block provides information about a block used in a
// Component.
type Block struct {
	//
	// BUILD SYMBOLS

	// Name is the name of the block.
	Name string

	Instances []*BlockInstance

	//
	// ANALYZER

	Required Analysis[bool]
}

func (b *Block) InstanceByNode(n *ast.Block) *BlockInstance {
	for _, instance := range b.Instances {
		if instance.AST == n {
			return instance
		}
	}
	return nil
}

// FirstIncludedAndPlaceholder returns the first instance of an &-placeholder
// in a block default that is included in the output of the component for the
// given component call.
//
// Passing nil checks the general case, in which all block defaults are
// included.
func (b *Block) FirstIncludedAndPlaceholder(cc *ComponentCall) Analysis[*BlockInstance] {
	if cc != nil && cc.BlockSetterByName(b.Name) != nil {
		return Result[*BlockInstance](nil)
	}

	var failed bool
	for _, instance := range b.Instances {
		ap := instance.Default.FirstAndPlaceholder
		if ap.Failed {
			failed = true
			continue
		}
		if ap.Result != nil && (cc == nil || !instance.DefaultOverwritten(cc)) {
			return Result(instance)
		}
	}

	return ResultIf[*BlockInstance](nil, !failed)
}

// FirstIncludedContentWriter returns the first instance of a content writer in
// a block default that is included in the output of the component for the given
// component call.
func (b *Block) FirstIncludedContentWriter(cc *ComponentCall) Analysis[*BlockInstance] {
	if cc != nil && cc.BlockSetterByName(b.Name) != nil {
		return Result[*BlockInstance](nil)
	}

	var failed bool
	for _, instance := range b.Instances {
		cw := instance.Default.FirstContentWriter
		if cw.Failed {
			failed = true
			continue
		}
		if cw.Result != nil && (cc == nil || !instance.DefaultOverwritten(cc)) {
			return Result(instance)
		}
	}

	return ResultIf[*BlockInstance](nil, !failed)
}

// FirstIncludedElementWriter returns the first instance of an element writer in
// a block default that is included in the output of the component for the given
// component call.
func (b *Block) FirstIncludedElementWriter(cc *ComponentCall) Analysis[*BlockInstance] {
	if cc != nil && cc.BlockSetterByName(b.Name) != nil {
		return Result[*BlockInstance](nil)
	}

	var failed bool
	for _, instance := range b.Instances {
		ew := instance.Default.FirstElementWriter
		if ew.Failed {
			failed = true
			continue
		}
		if ew.Result != nil && (cc == nil || !instance.DefaultOverwritten(cc)) {
			return Result(instance)
		}
	}

	return ResultIf[*BlockInstance](nil, !failed)
}

// TopLevel reports whether this block is top-level.
func (b *Block) TopLevel(s AnalysisStrategy) Analysis[bool] {
	s.assertValid()

	if s == All {
		for _, instance := range b.Instances {
			if instance.TopLevel.Failed {
				return FailedAnalysis[bool]()
			} else if !instance.TopLevel.Result {
				return Result(false)
			}
		}
		return Result(true)
	}

	var failed bool
	for _, instance := range b.Instances {
		if instance.TopLevel.Equal(true) {
			return Result(true)
		}
		failed = failed || instance.TopLevel.Failed
	}

	return ResultIf[bool](false, !failed)
}

type (
	BlockInstance struct {
		//
		// BUILD SYMBOLS

		Group *Block
		AST   *ast.Block
		// Parent is the instance of another block that contains this block.
		Parent *BlockInstance

		Default *BlockInstanceDefault // nil if no default

		//
		// ANALYZER

		// TopLevel indicates whether this block instance is placed outside
		// any element.
		TopLevel Analysis[bool]
		// CanWriteAttributes indicates whether this block instance can write
		// to the list of attributes of it's containing element.
		//
		// TopLevel implies CanWriteAttributes.
		CanWriteAttributes Analysis[bool]
	}

	BlockInstanceDefault struct {
		//
		// BUILD SYMBOLS

		AST ast.Body

		//
		// ANALYZER

		FirstAndPlaceholder         Analysis[*ast.AndPlaceholder]
		FirstTopLevelAndPlaceholder Analysis[*ast.AndPlaceholder]

		FirstTopLevelAttributeWriter Analysis[ast.AttributeWriter]
		FirstContentWriter           Analysis[ast.ContentWriter]
		FirstElementWriter           Analysis[ast.ElementWriter]
	}
)

// DefaultOverwritten indicates whether the default of this block instance
// is overwritten in the given component call.
// This is the case if the component call sets this block or one of this
// block's parent blocks.
func (cbi *BlockInstance) DefaultOverwritten(cc *ComponentCall) bool {
	if cc.BlockSetterByName(cbi.Group.Name) != nil {
		return true
	}
	if cbi.Parent != nil {
		return cbi.Parent.DefaultOverwritten(cc)
	}
	return false
}
