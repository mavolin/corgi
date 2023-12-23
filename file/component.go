package file

// ============================================================================
// Component
// ======================================================================================

type (
	Component struct {
		// Name is the name of the Component.
		Name Ident

		LParen Position
		// Params is a list of the parameters of the Component.
		Params []ComponentParam
		RParen Position

		// Body is the scope of the Component.
		Body Body

		*ComponentInfo

		Position
	}

	ComponentInfo struct {
		// WritesBody indicates whether the Component writes to the body of an
		// element.
		// Blocks including block defaults are ignored.
		WritesBody bool
		// WritesElements indicates whether the Component writes elements.
		//
		// Only true, if WritesBody is as well.
		WritesElements bool
		// WritesTopLevelAttributes indicates whether the Component writes any
		// top-level attributes, except &-placeholders.
		WritesTopLevelAttributes bool
		// AndPlaceholder indicates whether the Component has any
		// &-placeholders.
		AndPlaceholders bool
		// TopLevelAndPlaceholder indicates whether the Component has any
		// top-level &-placeholders.
		//
		// Only true, if AndPlaceholders is as well.
		TopLevelAndPlaceholder bool
		// Blocks is are the blocks used in the Component in the order they
		// appear in, and in the order they appear in the functions' signature.
		Blocks []ComponentBlockInfo
	}

	ComponentBlockInfo struct {
		// Name is the name of the block.
		Name string
		// TopLevel is true, if at least one block with Name is placed at the
		// top-level of the Component, so that it writes to the element it is
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

func (Component) _scopeItem() {}

// ==================================== Component Param =====================================

// ComponentParam represents a parameter of a Component.
type ComponentParam struct {
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

	Colon *Position
	// Default is the optional default value of the parameter.
	Default *Expression // never a chain expression

	Position
}

// ============================================================================
// Component Call
// ======================================================================================

// ComponentCall represents the call to a Component.
type ComponentCall struct {
	// Namespace is the namespace of the Component, if any.
	Namespace *Ident
	// Name is the name of the Component.
	Name Ident

	// Component is a pointer to the called Component.
	//
	// It is set by the linker.
	Component *LinkedComponent

	LParen Position
	// Args is a list of the arguments of given to the Component.
	Args   []ComponentArg
	RParen Position

	// Body is the body of the Component call.
	//
	// It will only consist of If, Switch, And, and Block items.
	Body Scope

	Position
}

func (ComponentCall) _scopeItem() {}

type LinkedComponent struct {
	// File is the file the Component was declared in.
	//
	// Note that the file's scope may be empty, if this Component was precompiled.
	File *File
	// Component is the Component itself.
	//
	// Note that the Component's body may be empty, if this Component was precompiled.
	Component *Component
}

// =================================== Component Call Arg ===================================

// ComponentArg represents a single argument given to a Component.
type ComponentArg struct {
	// Name is the name of the argument.
	Name Ident

	Colon Position

	// Value is the expression that yields the value of the argument.
	Value Expression

	Position
}

// ============================================================================
// Block
// ======================================================================================

// Block represents a block with optional content.
type Block struct {
	// Name is the name of the block.
	Name Ident
	Body Body

	Position
}

func (Block) _scopeItem() {}

// ============================================================================
// UnderscoreBlockShorthand
// ======================================================================================

type UnderscoreBlockShorthand struct {
	Body Body
	Position
}

func (UnderscoreBlockShorthand) _scopeItem() {}
