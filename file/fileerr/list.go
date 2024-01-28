package fileerr

import (
	"sort"
	"strings"
)

// Join joins the given errors into a single error.
// All errors must be either a pointer to an [Error] or a [List].
// Other errors are ignored.
//
// Returns either a [List] or nil.
func Join(errs ...error) error {
	var n int
	for _, err := range errs {
		//goland:noinspection GoTypeAssertionOnErrors
		switch err := err.(type) {
		case *Error:
			n++
		case List:
			n += len(err)
		}
	}

	if n == 0 {
		return nil
	}

	sum := make(List, 0, n)
	for _, err := range errs {
		//goland:noinspection GoTypeAssertionOnErrors
		switch err := err.(type) {
		case *Error:
			sum = append(sum, err)
		case List:
			sum = append(sum, err...)
		}
	}

	return sum
}

// List is a list of [Error] objects.
type List []*Error

var _ error = List(nil)

func (l List) Error() string {
	var sb strings.Builder
	for i, err := range l {
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(err.Error())
	}
	return sb.String()
}

// Pretty calls [Error.Pretty] on each non-nil [Error] in the list, separating each error
// by two newlines (effectively leaving a single blank line in-between errors).
//
// Errors are printed in sorted order.
// The source slice will not be modified.
func (l List) Pretty(o PrettyOptions) string {
	errs := l
	if !sort.IsSorted(l) {
		errs = make(List, len(l))
		copy(errs, l)
		sort.Sort(errs)
	}

	var sb strings.Builder
	for _, err := range errs {
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

var _ sort.Interface = List(nil)

func (l List) Len() int { return len(l) }

func (l List) Less(i, j int) bool {
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

func (l List) Swap(i, j int) { l[i], l[j] = l[j], l[i] }
