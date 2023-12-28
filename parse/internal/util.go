package internal

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/fileerr"
	anno2 "github.com/mavolin/corgi/internal/anno"
)

// pos returns the position of the current state as a [file.Position].
func pos(c *current) file.Position {
	return file.Position{
		Line: c.pos.line,
		Col:  c.pos.col,
	}
}

func slice(iface any) []any {
	if iface == nil {
		return nil
	}

	return iface.([]any)
}

func sliceOf[T any](ifacesI any) []T {
	ifaces := slice(ifacesI)
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

func getTuple[T any](iface any, index int) T {
	s := iface.([]any)
	if index < 0 {
		index = len(s) + index
	}

	return s[index].(T)
}

func optGetTuple[T any](iface any, index int) T {
	s, ok := iface.([]any)
	if !ok {
		var zero T
		return zero
	}

	if index < 0 {
		index = len(s) + index
	}

	if index < len(s) {
		return s[index].(T)
	}

	var zero T
	return zero
}

func optGetTuplePtr[T any](iface any, index int) *T {
	s, ok := iface.([]any)
	if !ok {
		return nil
	}

	if index < 0 {
		index = len(s) + index
	}

	if index < len(s) {
		t := s[index].(T)
		return &t
	}

	return nil
}

func getTuples[T any](tuplesI any, index int) []T {
	tuples := slice(tuplesI)
	if len(tuples) == 0 {
		return nil
	}

	s := make([]T, 0, len(tuples))
	for _, tuple := range tuples {
		tupleSlice := slice(tuple)
		if len(tupleSlice) == 0 {
			continue
		}

		index := index
		if index < 0 {
			index = len(tupleSlice) + index
		}

		if t, ok := tupleSlice[index].(T); ok {
			s = append(s, t)
		}
	}

	return s
}

func collectList[T any](firstI any, restI any, restTuplePos int) []T {
	restIs := slice(restI)

	list := make([]T, len(restIs)+1)
	list[0] = firstI.(T)
	for i, el := range restIs {
		list[i+1] = getTuple[T](el, restTuplePos)
	}

	return list
}

func optCast[T any](iface any) T {
	if casted, ok := iface.(T); ok {
		return casted
	}

	var zero T
	return zero
}

func ptr[T any](t T) *T {
	return &t
}

func optCastPtr[T any](iface any) *T {
	if casted, ok := iface.(T); ok {
		return &casted
	}

	return (*T)(nil)
}

func firstRune(iface any) rune {
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
func combineGoCode(exprsI any) file.GoCode {
	exprIs := slice(exprsI)
	exprs := combineGoCodeSlice(exprIs)
	exprs = exprs[:len(exprs):len(exprs)]
	return file.GoCode{Expressions: exprs}
}

func combineGoCodeSlice(exprIs []any) []file.GoCodeItem {
	exprs := make([]file.GoCodeItem, 0, 16)
	var prevGoCode *file.RawGoCode

	for _, eI := range exprIs {
		if eI == nil {
			continue
		}

		switch expr := eI.(type) {
		case []any:
			subExprs := combineGoCodeSlice(expr)
			if prevGoCode != nil {
				if len(subExprs) > 0 {
					if c, ok := subExprs[0].(file.RawGoCode); ok {
						prevGoCode.Code += c.Code
						subExprs = subExprs[1:]
					}
				}
				if len(subExprs) > 0 { // if there are still subExprs left
					exprs = append(exprs, *prevGoCode)
					prevGoCode = nil
				}
			}

			if len(subExprs) > 0 {
				if c, ok := subExprs[len(subExprs)-1].(file.RawGoCode); ok {
					prevGoCode = &c
					subExprs = subExprs[:len(subExprs)-1]
				}
			}

			exprs = append(exprs, subExprs...)
		case file.RawGoCode:
			if prevGoCode == nil {
				prevGoCode = &expr
			} else {
				prevGoCode.Code += expr.Code
			}
		case file.String:
			if prevGoCode != nil {
				exprs = append(exprs, *prevGoCode)
				prevGoCode = nil
			}

			exprs = append(exprs, expr)
		case file.BlockFunction:
			if prevGoCode != nil {
				exprs = append(exprs, *prevGoCode)
				prevGoCode = nil
			}

			exprs = append(exprs, expr)
		default:
			panic(fmt.Sprintf("parser: GoCode: invalid expression item %T (you shouldn't see this error, please open an issue)", expr))
		}
	}

	if prevGoCode != nil {
		exprs = append(exprs, *prevGoCode)
	}

	return exprs
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
