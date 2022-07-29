package mixin

import "fmt"

// ============================================================================
// InitCallError
// ======================================================================================

type InitCallError struct {
	Source string
	File   string
	Line   int
	Col    int
}

var _ error = (*InitCallError)(nil)

func (e *InitCallError) Error() string {
	return fmt.Sprintf("%s/%s:%d:%d: cannot call special mixin init",
		e.Source, e.File, e.Line, e.Col)
}

// ============================================================================
// InitPlacementError
// ======================================================================================

type InitPlacementError struct {
	Source string
	File   string
	Line   int
	Col    int
}

var _ error = (*InitPlacementError)(nil)

func (e *InitPlacementError) Error() string {
	return fmt.Sprintf("%s/%s:%d:%d: init mixins can only be placed at the top-level of non-main files",
		e.Source, e.File, e.Line, e.Col)
}

// ============================================================================
// InitParamsError
// ======================================================================================

type InitParamsError struct {
	Source string
	File   string
	Line   int
	Col    int
}

var _ error = (*InitParamsError)(nil)

func (e *InitParamsError) Error() string {
	return fmt.Sprintf("%s/%s:%d:%d: special mixin init cannot have parameters",
		e.Source, e.File, e.Line, e.Col)
}

// ============================================================================
// InitParamsError
// ======================================================================================

type InitBodyError struct {
	Source string
	File   string
	Line   int
	Col    int
}

var _ error = (*InitBodyError)(nil)

func (e *InitBodyError) Error() string {
	return fmt.Sprintf("%s/%s:%d:%d: special mixin init's body can only contain code",
		e.Source, e.File, e.Line, e.Col)
}

// ============================================================================
// InitFileTypeError
// ======================================================================================

type InitFileTypeError struct {
	Source string
	File   string
	Line   int
	Col    int
}

var _ error = (*InitFileTypeError)(nil)

func (e *InitFileTypeError) Error() string {
	return fmt.Sprintf("%s/%s:%d:%d: special mixin init can only be used in used and extended files",
		e.Source, e.File, e.Line, e.Col)
}

// ============================================================================
// ResourceNotFoundError
// ======================================================================================

// ResourceNotFoundError is the error returned if a mixin is used that has a
// namespace not found in the file's list of uses.
type ResourceNotFoundError struct {
	Namespace string
	Name      string

	Source string
	File   string
	Line   int
	Col    int
}

var _ error = (*ResourceNotFoundError)(nil)

func (r *ResourceNotFoundError) Error() string {
	return fmt.Sprintf(
		"%s/%s:%d:%d: no resource named `%s` found in list of uses, but is required for mixin `%s.%s`",
		r.Source, r.File, r.Line, r.Col, r.Namespace, r.Namespace, r.Name)
}

// ============================================================================
// NotFoundError
// ======================================================================================

// NotFoundError is returned when a resource does not provide a mixin with the
// given name.
type NotFoundError struct {
	UseSource string
	UseName   string
	Name      string

	Source string
	File   string
	Line   int
	Col    int
}

var _ error = (*NotFoundError)(nil)

func (e *NotFoundError) Error() string {
	if e.UseSource == "" || e.UseName == "" {
		return fmt.Sprintf("%s/%s:%d:%d: unknown mixin `%s`",
			e.Source, e.File, e.Line, e.Col, e.Name)
	}

	return fmt.Sprintf(
		"%s/%s:%d:%d: resource `%s/%s` provides no mixin called `%s`",
		e.Source, e.File, e.Line, e.Col, e.UseSource, e.UseName, e.Name)
}

// ============================================================================
// UnknownParamError
// ======================================================================================

type UnknownParamError struct {
	Name string

	Source string
	File   string
	Line   int
	Col    int
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
	Name string

	Source string
	File   string
	Line   int
	Col    int
}

var _ error = (*MissingParamError)(nil)

func (e *MissingParamError) Error() string {
	return fmt.Sprintf("%s/%s:%d:%d: missing required parameter `%s`",
		e.Source, e.File, e.Line, e.Col, e.Name)
}

// ============================================================================
// DuplicateParamError
// ======================================================================================

type DuplicateParamError struct {
	Name string

	Source string
	File   string
	Line   int
	Col    int

	OtherLine int
	OtherCol  int
}

var _ error = (*DuplicateParamError)(nil)

func (e *DuplicateParamError) Error() string {
	return fmt.Sprintf("%s/%s:%d:%d: duplicate parameter `%s` at %s/%s:%d:%d",
		e.Source, e.File, e.Line, e.Col, e.Name, e.Source, e.File, e.OtherLine, e.OtherCol)
}

// ============================================================================
// MissingNilCheckDefaultError
// ======================================================================================

type MissingNilCheckDefaultError struct {
	Name string

	Source string
	File   string
	Line   int
	Col    int
}

var _ error = (*MissingNilCheckDefaultError)(nil)

func (e *MissingNilCheckDefaultError) Error() string {
	return fmt.Sprintf("%s/%s:%d:%d: required parameter `%s` must have nil check default",
		e.Source, e.File, e.Line, e.Col, e.Name)
}

// ============================================================================
// DuplicateArgError
// ======================================================================================

type DuplicateArgError struct {
	Name string

	Source string
	File   string
	Line   int

	Col       int
	OtherLine int
	OtherCol  int
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
	// Name is the optional name of the block.
	Name string

	Source string
	File   string
	Line   int
	Col    int

	OtherLine int
	OtherCol  int
}

var _ error = (*DuplicateBlockError)(nil)

func (e *DuplicateBlockError) Error() string {
	if e.Name != "" {
		return fmt.Sprintf("%s/%s:%d:%d: duplicate block `%s` at %s/%s:%d:%d",
			e.Source, e.File, e.Line, e.Col, e.Name, e.Source, e.File, e.OtherLine, e.OtherCol)
	}

	return fmt.Sprintf("%s/%s:%d:%d: duplicate block at %s/%s:%d:%d",
		e.Source, e.File, e.Line, e.Col, e.Source, e.File, e.OtherLine, e.OtherCol)
}

// ============================================================================
// NestedBlockError
// ======================================================================================

type NestedBlockError struct {
	Source string
	File   string
	Line   int
	Col    int
}

var _ error = (*NestedBlockError)(nil)

func (e *NestedBlockError) Error() string {
	return fmt.Sprintf("%s/%s:%d:%d: blocks must be placed at the top-level of the mixin call body",
		e.Source, e.File, e.Line, e.Col)
}

// ============================================================================
// DuplicateBlockError
// ======================================================================================

type UnknownBlockError struct {
	// Name is the optional name of the block.
	Name string

	Source string
	File   string
	Line   int
	Col    int
}

var _ error = (*UnknownBlockError)(nil)

func (e *UnknownBlockError) Error() string {
	if e.Name != "" {
		return fmt.Sprintf("%s/%s:%d:%d: unknown block `%s` not defined by mixin",
			e.Source, e.File, e.Line, e.Col, e.Name)
	}

	return fmt.Sprintf("%s/%s:%d:%d: unknown block not defined by mixin", e.Source, e.File, e.Line, e.Col)
}

// ============================================================================
// RedeclaredError
// ======================================================================================

type RedeclaredError struct {
	Name string

	Source string
	File   string
	Line   int
	Col    int

	OtherSource string
	OtherFile   string
	OtherLine   int
	OtherCol    int
}

var _ error = (*RedeclaredError)(nil)

func (e *RedeclaredError) Error() string {
	return fmt.Sprintf("%s/%s:%d:%d: mixin `%s` redeclared at %s/%s:%d:%d",
		e.Source, e.File, e.Line, e.Col,
		e.Name,
		e.OtherSource, e.OtherFile, e.OtherLine, e.OtherCol)
}

// ============================================================================
// UnexportedExternalMixinError
// ======================================================================================

type UnexportedExternalMixinError struct {
	Namespace string
	Name      string

	Source string
	File   string
	Line   int
	Col    int
}

var _ error = (*UnexportedExternalMixinError)(nil)

func (e *UnexportedExternalMixinError) Error() string {
	return fmt.Sprintf("%s/%s:%d:%d: cannot access unexported external mixin `%s.%s`",
		e.Source, e.File, e.Line, e.Col, e.Namespace, e.Name)
}

// ============================================================================
// CallBodyError
// ======================================================================================

type CallBodyError struct {
	Source string
	File   string
	Line   int
	Col    int
}

var _ error = (*CallBodyError)(nil)

func (e *CallBodyError) Error() string {
	return fmt.Sprintf("%s/%s:%d:%d: cannot use text items inside mixin call bodies",
		e.Source, e.File, e.Line, e.Col)
}
