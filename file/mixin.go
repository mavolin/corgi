package file

// ============================================================================
// Mixin
// ======================================================================================

type Mixin struct {
	// Name is the name of the mixin.
	Name Ident

	LParenPos *Position // nil if params were omitted

	// Params is a list of the parameters of the mixin.
	Params []MixinParam

	RParenPos *Position

	// Body is the scope of the mixin.
	Body Scope

	// Precompiled is the precompiled function literal.
	// Its args start with the mixins args, followed by func()s for each of
	// the Blocks, and lastly, if HasAndPlaceholders is true, a final func()
	// called each time that the mixin's &s are supposed to be placed.
	//
	// It is only present, if this mixin was precompiled.
	Precompiled []byte

	*MixinInfo

	Position
}

type (
	MixinInfo struct {
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
		Blocks             []MixinBlockInfo
		HasAndPlaceholders bool
	}

	MixinBlockInfo struct {
		Name     string
		TopLevel bool // writes directly to the element it is called in
		// CanAttributes specifies whether &-directives can be used in this block.
		CanAttributes                   bool
		DefaultWritesBody               bool
		DefaultWritesElements           bool
		DefaultWritesTopLevelAttributes bool
		DefaultTopLevelAndPlaceholder   bool
	}
)

var _ ScopeItem = Mixin{}

func (Mixin) _typeScopeItem() {}

// ==================================== Mixin Param =====================================

// MixinParam represents a parameter of a mixin.
type MixinParam struct {
	// Name is the name of the parameter.
	Name Ident

	// Type is the name of the type of the parameter, or nil if the type is
	// inferred from the default.
	Type *GoType
	// InferredType is the type inferred from the Default, if Type is nil.
	//
	// It will be set by package typeinfer before linking.
	//
	// An empty string indicates the type could not be inferred.
	InferredType string

	AssignPos *Position
	// Default is the optional default value of the parameter.
	Default *Expression // never a chain expression

	Position
}

// ============================================================================
// Return
// ======================================================================================

type Return struct {
	Err *Expression
	Position
}

var _ ScopeItem = Return{}

func (Return) _typeScopeItem() {}

// ============================================================================
// Mixin Call
// ======================================================================================

// MixinCall represents the call to a mixin.
type MixinCall struct {
	// Namespace is the namespace of the mixin, if any.
	Namespace *Ident
	// Name is the name of the mixin.
	Name Ident

	// Mixin is a pointer to the called mixin.
	//
	// It is set by the linker.
	Mixin *LinkedMixin

	LParenPos *Position

	// Args is a list of the arguments of given to the mixin.
	Args []MixinArg

	RParenPos *Position

	// Body is the body of the mixin call.
	//
	// It will only consist of If, IfBlock, Switch, And, and Block items.
	Body Scope

	Position
}

var _ ScopeItem = MixinCall{}

func (MixinCall) _typeScopeItem() {}

type LinkedMixin struct {
	// File is the file the mixin was declared in.
	//
	// Note that the file's scope may be empty, if this mixin was precompiled.
	File *File
	// Mixin is the mixin itself.
	//
	// Note that the mixin's body may be empty, if this mixin was precompiled.
	Mixin *Mixin
}

// =================================== Mixin Call Arg ===================================

// MixinArg represents a single argument given to a mixin.
type MixinArg struct {
	// Name is the name of the argument.
	Name Ident
	// Value is the expression that yields the value of the argument.
	Value Expression

	Position
}

// ============================================================================
// Mixin Main Block Shorthand
// ======================================================================================

type MixinMainBlockShorthand struct {
	Body Scope
	Position
}

var _ ScopeItem = MixinMainBlockShorthand{}

func (MixinMainBlockShorthand) _typeScopeItem() {}
