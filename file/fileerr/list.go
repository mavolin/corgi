package fileerr

import (
	"sort"
	"strings"
)

// As collects all [Error] objects represented by err into a slice.
// All remaining errors are returned as a slice of errors
func As(err error) ([]*Error, []error) {
	if err == nil {
		return nil, nil
	}

	orig := err

	for {
		ferr, ok := err.(*Error) //nolint:errorlint
		if ok {
			return []*Error{ferr}, nil
		}

		as, ok := err.(interface{ As(any) bool })
		var e *Error
		if ok && as.As(&e) {
			return []*Error{e}, nil
		}

		switch x := err.(type) {
		case interface{ Unwrap() error }:
			err = x.Unwrap()
			if err == nil {
				return nil, nil
			}
		case interface{ Unwrap() []error }:
			var ferrs []*Error
			var errs []error

			for _, err := range x.Unwrap() {
				ferr, stErr := As(err)
				ferrs = append(ferrs, ferr...)
				errs = append(errs, stErr...)
			}

			if len(ferrs) == 0 {
				return nil, []error{orig}
			}

			return ferrs, errs
		default:
			return nil, []error{orig}
		}
	}
}

// Pretty calls [Error.Pretty] on each non-nil [Error] in the list, separating each error
// by two newlines (effectively leaving a single blank line in-between errors).
//
// Errors are sorted before being printed.
//
// Non [Error] errors are printed last using their Error method.
func Pretty(err error, o PrettyOptions) string {
	ferrs, errs := As(err)
	Sort(ferrs)

	var sb strings.Builder
	for _, err := range ferrs {
		if err == nil {
			continue
		}
		sb.WriteString(err.Pretty(o))
		sb.WriteString("\n\n")
	}

	for _, err := range errs {
		if err == nil {
			continue
		}
		sb.WriteString(err.Error())
		sb.WriteString("\n\n")
	}

	s := sb.String()
	if len(s) > 0 {
		s = s[:len(s)-2]
	}
	return s
}

// Sort sorts the list of errors by file name, line number, and column number.
func Sort(errs []*Error) {
	sort.Sort(list(errs))
}

type list []*Error

var _ sort.Interface = list(nil)

func (l list) Len() int { return len(l) }

func (l list) Less(i, j int) bool {
	a, b := l[i].ErrorAnnotation, l[j].ErrorAnnotation

	aMod := a.File.Module + "/" + a.File.PathInModule
	bMod := b.File.Module + "/" + b.File.PathInModule
	if aMod != bMod {
		return aMod < bMod
	} else if a.Line != b.Line {
		return a.Line < b.Line
	}

	return a.Start < b.Start
}

func (l list) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
