package link

import "fmt"

// ============================================================================
// ImportNamespaceError
// ======================================================================================

type ImportNamespaceError struct {
	// Source is the source of the file that attempted to import the package,
	// but could not because of the conflict in the other file.
	Source string
	// File is the name of the file that attempted to import the package,
	// but could not because of the conflict in the other file.
	File string

	Line int
	Col  int

	OtherSource string
	OtherFile   string
	OtherLine   int
	OtherCol    int

	// Namespace is the conflicting checkNamespaceCollisions.
	Namespace string
}

var _ error = (*ImportNamespaceError)(nil)

func (e *ImportNamespaceError) Error() string {
	return fmt.Sprintf("%s/%s:%d:%d: import namespace `%s` already in use in %s/%s:%d/%d for different import",
		e.Source, e.File, e.Line, e.Col,
		e.Namespace,
		e.OtherSource, e.OtherFile, e.OtherLine, e.OtherCol)
}

// ============================================================================
// UseNamespaceError
// ======================================================================================

type UseNamespaceError struct {
	Source string
	File   string
	Line   int

	OtherLine int

	Namespace string
}

var _ error = (*UseNamespaceError)(nil)

func (e *UseNamespaceError) Error() string {
	return fmt.Sprintf("%s/%s:%d: namespace collision with `use` in line %d",
		e.Source, e.File, e.Line, e.Line)
}

// ============================================================================
// MixinNotFoundError
// ======================================================================================

type MixinNotFoundError struct {
	Source string
	File   string
	Line   int
	Col    int

	Namespace string
	Name      string
}

var _ error = (*MixinNotFoundError)(nil)

func (e *MixinNotFoundError) Error() string {
	return fmt.Sprintf("%s/%s:%d:%d: unknown mixin `%s.%s`", e.Source, e.File, e.Line, e.Col, e.Namespace, e.Name)
}

// ============================================================================
// ParamNotFoundError
// ======================================================================================

type UnknownParamError struct {
	Source string
	File   string
	Line   int
	Col    int

	Name string
}

var _ error = (*UnknownParamError)(nil)

func (e *UnknownParamError) Error() string {
	return fmt.Sprintf("%s/%s:%d:%d: unknown mixin parameter `%s`",
		e.Source, e.File, e.Line, e.Col, e.Name)
}

// ============================================================================
// MissingParamError
// ======================================================================================

type MissingParamError struct {
	Source string
	File   string
	Line   int
	Col    int

	Name string
}

var _ error = (*MissingParamError)(nil)

func (e *MissingParamError) Error() string {
	return fmt.Sprintf("%s/%s:%d:%d: missing required parameter `%s`",
		e.Source, e.File, e.Line, e.Col, e.Name)
}

// ============================================================================
// MissingNilCheckDefaultError
// ======================================================================================

type MissingNilCheckDefaultError struct {
	Source string
	File   string
	Line   int
	Col    int

	Name string
}

var _ error = (*MissingNilCheckDefaultError)(nil)

func (e *MissingNilCheckDefaultError) Error() string {
	return fmt.Sprintf("%s/%s:%d:%d: required parameter `%s` must have nil check default",
		e.Source, e.File, e.Line, e.Col, e.Name)
}

// ============================================================================
// DuplicateParamError
// ======================================================================================

type DuplicateParamError struct {
	Source string
	File   string
	Line   int
	Col    int

	OtherLine int
	OtherCol  int

	Name string
}

var _ error = (*DuplicateParamError)(nil)

func (e *DuplicateParamError) Error() string {
	return fmt.Sprintf("%s/%s:%d:%d: duplicate parameter `%s` at %s/%s:%d:%d",
		e.Source, e.File, e.Line, e.Col, e.Name, e.Source, e.File, e.OtherLine, e.OtherCol)
}

// ============================================================================
// DuplicateArgError
// ======================================================================================

type DuplicateArgError struct {
	Source string
	File   string
	Line   int
	Col    int

	OtherLine int
	OtherCol  int

	Name string
}

var _ error = (*DuplicateArgError)(nil)

func (e *DuplicateArgError) Error() string {
	return fmt.Sprintf("%s/%s:%d:%d: duplicate argument `%s` at %s/%s:%d:%d",
		e.Source, e.File, e.Line, e.Col, e.Name, e.Source, e.File, e.OtherLine, e.OtherCol)
}

// ============================================================================
// DuplicateBlock
// ======================================================================================

type DuplicateBlockError struct {
	Source string
	File   string
	Line   int
	Col    int

	OtherLine int
	OtherCol  int
}

var _ error = (*DuplicateBlockError)(nil)

func (e *DuplicateBlockError) Error() string {
	return fmt.Sprintf("%s/%s:%d:%d: duplicate block at %s/%s:%d:%d",
		e.Source, e.File, e.Line, e.Col, e.Source, e.File, e.OtherLine, e.OtherCol)
}

// ============================================================================
// MixinRedeclaredError
// ======================================================================================

type MixinRedeclaredError struct {
	Source string
	File   string
	Line   int
	Col    int

	OtherSource string
	OtherFile   string
	OtherLine   int
	OtherCol    int

	Name string
}

var _ error = (*MixinRedeclaredError)(nil)

func (e *MixinRedeclaredError) Error() string {
	return fmt.Sprintf("%s/%s:%d:%d: mixin `%s` redeclared at %s/%s:%d:%d",
		e.Source, e.File, e.Line, e.Col,
		e.Name,
		e.OtherSource, e.OtherFile, e.OtherLine, e.OtherCol)
}

// ============================================================================
// UnexportedExternalMixinError
// ======================================================================================

type UnexportedExternalMixinError struct {
	Source string
	File   string
	Line   int
	Col    int

	Namespace string
	Name      string
}

var _ error = (*UnexportedExternalMixinError)(nil)

func (e *UnexportedExternalMixinError) Error() string {
	return fmt.Sprintf("%s/%s:%d:%d: cannot access unexported external mixin `%s.%s`",
		e.Source, e.File, e.Line, e.Col, e.Namespace, e.Name)
}

// ============================================================================
// IllegalAndError
// ======================================================================================

type IllegalAndError struct {
	Source string
	File   string
	Line   int
	Col    int
}

var _ error = (*IllegalAndError)(nil)

func (e *IllegalAndError) Error() string {
	return fmt.Sprintf("%s/%s:%d:%d: cannot use `&` in blocks or after writing to an element's body",
		e.Source, e.File, e.Line, e.Col)
}

// ============================================================================
// SelfClosingContentError
// ======================================================================================

type SelfClosingContentError struct {
	Source string
	File   string
	Line   int
	Col    int
}

var _ error = (*SelfClosingContentError)(nil)

func (e *SelfClosingContentError) Error() string {
	return fmt.Sprintf("%s/%s:%d:%d: self-closing elements and void elements cannot have any content",
		e.Source, e.File, e.Line, e.Col)
}

// ============================================================================
// MixinContentError
// ======================================================================================

type MixinContentError struct {
	Source string
	File   string
	Line   int
	Col    int
}

var _ error = (*MixinContentError)(nil)

func (e *MixinContentError) Error() string {
	return fmt.Sprintf("%s/%s:%d:%d: cannot use text items inside mixin calls",
		e.Source, e.File, e.Line, e.Col)
}
