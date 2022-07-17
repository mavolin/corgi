package parse

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/mavolin/corgi/corgi/lex"
)

var (
	// ErrNoFunc is the error returned during parsing, if a file that is being
	// parsed in ModeMain has no function definition.
	ErrNoFunc = errors.New("main files must define a func")

	// ErrMultipleProlog is the error multiple lex.Doctype items are
	// encountered, that define a prolog.
	ErrMultipleProlog = errors.New("you may only specify one XML prolog")
	// ErrMultipleDoctype is the error multiple lex.Doctype items are
	// encountered, that define a doctype (not a prolog).
	ErrMultipleDoctype = errors.New("you may only specify one doctype")
	// ErrMultipleExtend is the error returned if multiple lex.Extend items are
	// encountered.
	ErrMultipleExtend = errors.New("you may only specify one file to extend")
	// ErrMultipleFunc is the error returned if multiple lex.Func items are
	// encountered.
	ErrMultipleFunc = errors.New("you may only specify one func")

	// ErrExtendPlacement is the error returned if lex.Extend does not appear as
	// first item in a file.
	ErrExtendPlacement = errors.New("extend must be the first item in the file")
	// ErrImportPlacement is the error returned if imports are placed wrong.
	ErrImportPlacement = errors.New("imports must be declared directly after the extend statement")
	ErrUsePlacement    = errors.New("use statements must be declared directly after imports")

	ErrExtendDoctype = errors.New("files extending other files may not define a doctype or prolog")
	ErrExtendFunc    = errors.New("extended files may not define a func")

	// ErrUseExtends is the error returned if a file, that is being parsed in
	// ModeUse, has an extend statement.
	ErrUseExtends = errors.New("used files cannot extend other files")
	// ErrUseFunc is the error returned if a file, that is being parsed in
	// ModeUse, has a func statement.
	ErrUseFunc = errors.New("used files cannot define a func")
	// ErrUseDoctype is the error returned if a file, that is being parsed in
	// ModeUse, has a doctype statement.
	ErrUseDoctype = errors.New("used files cannot define a doctype")

	// ErrIncludeExtends is the error returned if a file, that is being parsed
	// in ModeInclude, has an extend statement.
	ErrIncludeExtends = errors.New("included files cannot extend other files")

	// ErrTernaryCondition is the error returned if a nil check is used in a
	// ternary expression.
	ErrTernaryCondition = errors.New("cannot use nil check or ternary expression as ternary condition")
	ErrNilCheckDefault  = errors.New("cannot use nil check or ternary expression as nil-check default")

	ErrIndexExpression    = errors.New("index expression has unclosed brackets")
	ErrFuncCallExpression = errors.New("func call expression has unclosed brackets")

	ErrCaseExpression         = errors.New("case expressions must resolve to regular Go expressions")
	ErrWhileExpression        = errors.New("while conditions must resolve to regular Go expressions")
	ErrMixinDefaultExpression = errors.New("mixin defaults must resolve to regular Go expressions")
)

// ============================================================================
// UnexpectedItemError
// ======================================================================================

// UnexpectedItemError is the error returned when an unknown item is encountered.
type UnexpectedItemError struct {
	Found    lex.ItemType
	Expected []lex.ItemType
}

var _ error = (*UnexpectedItemError)(nil)

func (e *UnexpectedItemError) Error() string {
	if len(e.Expected) == 0 {
		return fmt.Sprintf("unexpected item %s", e.Found.String())
	} else if len(e.Expected) == 1 {
		return fmt.Sprintf("unexpected item %s: expected %s", e.Found.String(), e.Expected[0].String())
	}

	var b strings.Builder

	for i := 0; i < len(e.Expected)-1; i++ {
		b.WriteString(e.Expected[i].String())
		b.WriteString(", ")
	}

	b.WriteString("or ")
	b.WriteString(e.Expected[len(e.Expected)-1].String())

	return fmt.Sprintf("unexpected item %s: expected %s", e.Found.String(), b.String())
}
