package file

type Library struct {
	// METADATA
	//

	// Module is the path/name of the Go module providing this library.
	Module string
	// PathInModule is the path to the library in the Go module, relative to the
	// module root.
	//
	// It is always specified as a forward slash separated path.
	PathInModule string
	// AbsolutePath is the resolved absolute path to the library.
	//
	// It is specified using the system's separator.
	AbsolutePath string

	// Precompiled indicates whether this library was precompiled.
	//
	// If true, the files in this library will only have Type, Name, Module,
	// and PathInModule set.
	//
	// Additionally, Imports will be set,
	Precompiled bool

	//
	// FILES
	//

	// Files are the files this library consists of.
	//
	// If the library is precompiled, only Name, Module, PathInModule, and
	// Imports will be set.
	Files []*File

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

	Mixins []MixinDependency
}

type MixinDependency struct {
	// Name is name of the mixin depended on.
	Name string
	// Var is the variable used by the depending mixins to call this mixin.
	Var string
	// RequiredBy are the names of the depending mixins.
	RequiredBy []string
}

type PrecompiledCode struct {
	MachineComments []string
	Lines           []string
}

type PrecompiledMixin struct {
	// File is the file in which the mixin appears.
	File *File

	MachineComments []string
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
}
