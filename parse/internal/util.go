package internal

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/fileerr"
	anno2 "github.com/mavolin/corgi/internal/anno"
)

// pos returns the position of the current state as a [file.Position].
func pos(c *current) file.Position {
	return file.Position{
		Line: c.pos.line,
		Col:  c.pos.col,
	}
}

func islice(iface any) []any {
	if iface == nil {
		return nil
	}

	return iface.([]any)
}

func typedSlice[T any](ifacesI any) []T {
	ifaces := islice(ifacesI)
	if len(ifaces) == 0 {
		return nil
	}

	slice := make([]T, 0, len(ifaces))
	for _, iface := range ifaces {
		if t, ok := iface.(T); ok {
			slice = append(slice, t)
		}
	}

	return slice
}

func typedSliceFromTuples[T any](tuplesI any, index int) []T {
	tuples := islice(tuplesI)
	if len(tuples) == 0 {
		return nil
	}

	slice := make([]T, 0, len(tuples))
	for _, tuple := range tuples {
		tupleSlice := islice(tuple)
		if len(tupleSlice) == 0 {
			continue
		}

		index := index
		if index < 0 {
			index = len(tupleSlice) + index
		}

		if t, ok := tupleSlice[index].(T); ok {
			slice = append(slice, t)
		}
	}

	return slice
}

func getTuple[T any](iface any, index int) T {
	slice := iface.([]any)
	if index < 0 {
		index = len(slice) + index
	}

	return slice[index].(T)
}

func castedOrZero[T any](iface any) T {
	if casted, ok := iface.(T); ok {
		return casted
	}

	var zero T
	return zero
}

func ptrOrNil[T any](iface any) *T {
	if casted, ok := iface.(T); ok {
		return &casted
	}

	return (*T)(nil)
}

func ptr[T any](t T) *T {
	return &t
}

func char(iface any) rune {
	c, _ := utf8.DecodeRune(iface.([]byte))
	return c
}

// concat concatenates all elements captured by an expression.
//
// It assumes iface is either a []byte or a []any that, recursively,
// contains []byte or []any slices.
func concat(iface any) string {
	if iface == nil {
		return ""
	}

	var sb strings.Builder
	sb.Grow(256)
	concatBuilder(&sb, iface)
	return sb.String()
}

// concatBuilder is a helper for concat that writes the contents of iface
// to sb.
//
// If it encounters a []byte, it writes it to sb.
// If it encounters a []any, it calls concatBuilder on each element.
func concatBuilder(sb *strings.Builder, iface any) {
	if iface == nil {
		return
	}

	switch iface := iface.(type) {
	case []byte:
		sb.Write(iface)
	case []any:
		for _, v := range iface {
			concatBuilder(sb, v)
		}
	}
}

type annotation = anno2.Annotation

func anno(c *current, aw annotation) fileerr.Annotation {
	return anno2.Lines(c.globalStore["lines"].([]string), aw)
}

// ============================================================================
// expression.peg
// ======================================================================================

//nolint:unparam
func combineExpressions(exprsI any) (file.Expression, error) {
	exprIs := islice(exprsI)
	var exprs []file.ExpressionItem

	var prevGoExpr *file.GoExpression

	for _, ei := range exprIs {
		switch expr := ei.(type) {
		case []file.ExpressionItem:
			for _, e := range expr {
				switch expr := e.(type) {
				case file.GoExpression:
					if prevGoExpr == nil {
						prevGoExpr = &expr
					} else {
						prevGoExpr.Expression += expr.Expression
					}
				default:
					if prevGoExpr != nil {
						exprs = append(exprs, *prevGoExpr)
						prevGoExpr = nil
					}
					exprs = append(exprs, expr)
				}
			}
		case file.StringExpression:
			if prevGoExpr != nil {
				exprs = append(exprs, *prevGoExpr)
				prevGoExpr = nil
			}

			exprs = append(exprs, expr)
		case file.TernaryExpression:
			if prevGoExpr != nil {
				exprs = append(exprs, *prevGoExpr)
				prevGoExpr = nil
			}

			exprs = append(exprs, expr)
		case file.GoExpression:
			if prevGoExpr == nil {
				prevGoExpr = &expr
			} else {
				prevGoExpr.Expression += expr.Expression
			}
		default:
			panic(fmt.Sprintf("parser: GoExpression: invalid expression item %T:\n%#v\n\n(you shouldn't see this error, please open an issue)", expr, expr))
		}
	}

	if prevGoExpr != nil {
		exprs = append(exprs, *prevGoExpr)
	}

	return file.Expression{Expressions: exprs}, nil
}

func chainExprItmsCheck(itms []file.ChainExpressionItem) bool {
	for _, itm := range itms {
		switch itm := itm.(type) {
		case file.IndexExpression:
			if itm.CheckValue || itm.CheckIndex {
				return true
			}
		case file.DotIdentExpression:
			if itm.Check {
				return true
			}
		case file.ParenExpression:
			if itm.Check {
				return true
			}
		case file.TypeAssertionExpression:
			if itm.Check {
				return true
			}
		}
	}

	return false
}
