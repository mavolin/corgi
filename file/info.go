package file

type (
	PackageInfo struct {
		// HasState indicates whether this package contains state variables.
		HasState bool

		Components []ComponentInfo
	}

	ComponentInfo struct {
		Name   string
		Params []ComponentParamInfo

		// WritesBody indicates whether the ComponentInfo writes to the body of an
		// element.
		// Blocks including block defaults are ignored.
		WritesBody bool
		// WritesElements indicates whether the ComponentInfo writes elements.
		//
		// Only true, if WritesBody is as well.
		WritesElements bool
		// WritesTopLevelAttributes indicates whether the ComponentInfo writes any
		// top-level attributes, except &-placeholders.
		WritesTopLevelAttributes bool
		// AndPlaceholder indicates whether the ComponentInfo has any
		// &-placeholders.
		AndPlaceholders bool
		// TopLevelAndPlaceholder indicates whether the ComponentInfo has any
		// top-level &-placeholders.
		//
		// Only true, if AndPlaceholders is as well.
		TopLevelAndPlaceholder bool
		// Blocks is are the blocks used in the ComponentInfo in the order they
		// appear in, and in the order they appear in the functions' signature.
		Blocks []ComponentBlockInfo
	}

	ComponentParamInfo struct {
		Name       string
		Type       string
		IsSafeType bool // type from package safe
		HasDefault bool
	}

	ComponentBlockInfo struct {
		// Name is the name of the block.
		Name string
		// TopLevel is true, if at least one block with Name is placed at the
		// top-level of the ComponentInfo, so that it writes to the element it is
		// called in.
		TopLevel bool // writes directly to the element it is called in
		// CanAttributes specifies whether &-directives can be used in this
		// block.
		CanAttributes bool
		// DefaultWritesBody indicates whether the block writes to the body of
		// the element.
		DefaultWritesBody bool
		// DefaultWritesElements indicates whether the block writes any
		// elements.
		//
		// Only true, if DefaultWritesBody is as well.
		DefaultWritesElements bool
		// DefaultWritesTopLevelAttributes indicates whether the block writes
		// any top-level attributes, except &-placeholders.
		DefaultWritesTopLevelAttributes bool
		// DefaultAndPlaceholder indicates whether the block has any
		// &-placeholders at the top-level.
		DefaultTopLevelAndPlaceholder bool
	}
)

func AnalyzePackage(p *Package) *PackageInfo {
	// todo
	panic("todo")
}
