package file

import "github.com/mavolin/corgi/file/packageinfo"

// ============================================================================
// Component
// ======================================================================================

type Component struct {
	Name Ident

	LParen Position
	Params []ComponentParam
	RParen Position

	Body Body

	Info *packageinfo.Component // set by linker

	Position
}

func (Component) _scopeItem() {}

// ==================================== Component Param =====================================

// ComponentParam represents a parameter of a Component.
type ComponentParam struct {
	Name Ident

	// Type is the name of the type of the parameter, or nil if the type is
	// inferred from the default.
	Type *Type
	// InferredType is the type inferred from the Default, if Type is nil.
	//
	// It will be set by package typeinfer before linking.
	//
	// An empty string indicates the type could not be inferred.
	InferredType string

	Colon *Position
	// Default is the optional default value of the parameter.
	Default *GoCode

	Position
}

// ============================================================================
// Component Call
// ======================================================================================

type ComponentCall struct {
	Namespace *Ident
	Name      Ident

	Link *LinkedComponent // set by linker

	LParen Position
	Args   []ComponentArg
	RParen Position

	// Body is the body of the Component call.
	//
	// It will only consist of If, Switch, For, And, and Block items.
	Body Body

	Position
}

func (ComponentCall) _scopeItem() {}

type LinkedComponent struct {
	Package *packageinfo.Package
	Info    *packageinfo.Component

	// File is the file the Component is defined in.
	//
	// Users should not rely on this field being set.
	// Usually it's only set when a package is recompiled, or was compiled for
	// the first time.
	File *File
	// Component is the Component this LinkedComponent links to.
	//
	// Users should not rely on this field being set.
	// Usually it's only set when a package is recompiled, or was compiled for
	// the first time.
	Component *Component
}

// =================================== Component Call Arg ===================================

type ComponentArg struct {
	Name  Ident
	Colon Position
	Value Expression

	Position
}

// ============================================================================
// Block
// ======================================================================================

type Block struct {
	Name Ident
	Body Body // may be nil

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
