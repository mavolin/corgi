package file

type Library struct {
	// METADATA
	//

	// Module is the path/name of the Go module providing this library.
	Module string
	// ModulePath is the path to the library in the Go module, relative to the
	// module root.
	//
	// It is always specified as a forward slash separated path.
	ModulePath string
	// AbsolutePath is the resolved absolute path to the library.
	//
	// It is always specified as a forward slash separated path.
	AbsolutePath string

	// Precompiled indicates whether this library was precompiled.
	//
	// If true, the files in this library will only have their metadata set.
	Precompiled bool

	//
	// FILES
	//

	Files []File

	//
	// PRECOMPILATION DATA
	//
	// These fields are only set, if this library was precompiled.

	Dependencies []LibDependency

	GlobalCode []PrecompiledCode

	Mixins []PrecompiledMixin
}

type LibDependency struct {
	// Module is the path/name of the Go module providing this library.
	Module string
	// ModulePath is the path to the library in the Go module, relative to the
	// module root.
	//
	// It is always specified as a forward slash separated path.
	ModulePath string

	// Name is name of the mixin depended on.
	Name string

	RequiredBy []string
}

type MixinDependency struct {
	// Name is name of the mixin depended on.
	Name string
	// RequiredBy are the names of the depending mixins.
	RequiredBy []string
}

type PrecompiledCode struct {
	Comments []CorgiComment // only machine comments
	Code     Code
}

type PrecompiledMixin struct {
	// Mixin is the mixin itself.
	//
	// Its body is empty.
	Mixin Mixin

	// Precompiled is the precompiled function literal.
	// Its args start with the mixins args, followed by func()s for each of
	// the Blocks, and lastly, if HasAndPlaceholders is true, a final func()
	// called each time that the mixin's &s are supposed to be placed.
	//
	// It is only present, if this mixin was precompiled.
	Precompiled []byte

	// WritesBody indicates whether the mixin writes to the body of an element.
	// Blocks including block defaults are ignored.
	WritesBody bool
	// WritesElements indicates whether the mixin writes elements.
	//
	// Only true, if WritesBody is as well.
	WritesElements bool
	// WritesTopLevelAttributes indicates whether the mixin writes any top-level
	// attributes, except &-placeholders.
	WritesTopLevelAttributes bool
	// TopLevelAndPlaceholder indicates whether the mixin has any top-level
	// &-placeholders.
	TopLevelAndPlaceholder bool
	// Blocks is are the blocks used in the mixin in the order they appear in,
	// and in the order they appear in the functions' signature.
	Blocks             []PrecompiledMixinBlock
	HasAndPlaceholders bool
}

type PrecompiledMixinBlock struct {
	Name     string
	TopLevel bool // writes directly to the element it is called in
	// CanAttributes specifies whether &-directives can be used in this block.
	CanAttributes                   bool
	DefaultWritesBody               bool
	DefaultWritesTopLevelAttributes bool
	DefaultTopLevelAndPlaceholder   bool
}
