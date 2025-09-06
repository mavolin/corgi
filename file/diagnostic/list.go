package diagnostic

import (
	"cmp"
	"errors"
	"slices"
	"strings"
)

type List []*Diagnostic

// Tidy first orders the diagnostics in the list, first by type, then by
// module, package, file, line and column.
//
// It then removes all duplicates.
// That includes diagnostics with the same pointer, but also diagnostics with
// identical message and cause, regardless of annotations or other fields.
// The rationale is that if the same error caused the same diagnostic (as
// determined by equal messages), it should only be reported once.
func (l *List) Tidy() {
	slices.SortFunc(*l, sort)

	// duplicates are always adjacent, after sorting
	*l = slices.Compact(*l)
	if len(*l) > 1 {
		for ai, a := range (*l)[:len(*l)-1] {
			if a == nil {
				continue
			}

			for bi, b := range (*l)[ai+1:] {
				if b == nil {
					continue
				}
				if a.Message == b.Message && a.Cause != nil && b.Cause != nil && (errors.Is(a.Cause, b.Cause) || errors.Is(b.Cause, a.Cause)) {
					// same message and cause, remove b
					(*l)[ai+1+bi] = nil
				}
			}
		}
	}
	*l = slices.DeleteFunc(*l, func(d *Diagnostic) bool { return d == nil })
	*l = slices.Clip(*l)
}

func (l List) Short() string {
	var sb strings.Builder
	sb.Grow(len(l) * 128)
	for i, d := range l {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(d.Short())
	}
	return sb.String()
}

func (l List) Pretty(o PrettyOptions) string {
	o.applyDefaults()

	if !slices.IsSortedFunc(l, sort) {
		clone := make(List, len(l))
		copy(clone, l)
		slices.SortStableFunc(clone, sort)
		//goland:noinspection GoAssignmentToReceiver
		l = clone
	}

	var sb strings.Builder
	sb.Grow(len(l) * 1024)
	for i, d := range l {
		if i > 0 {
			if d.Explanation != "" || len(d.Examples) > 0 || len(d.Hints) > 0 || d.Docs != "" {
				sb.WriteString("\n\n\n") // two empty lines for better visual separation
			} else {
				sb.WriteString("\n\n")
			}
		}

		d.pretty(&sb, o)
	}

	return sb.String()
}

func sort(a, b *Diagnostic) int {
	var aa, ba *Annotation
	if len(a.Primary) > 0 {
		aa = &a.Primary[0]
	}
	if len(b.Primary) > 0 {
		ba = &b.Primary[0]
	}

	ta, tb := typeOrder(a.Type), typeOrder(b.Type)
	switch {
	case ta != tb:
		return cmp.Compare(ta, tb)
	case ta == 4 && tb == 4 && a.Type != b.Type:
		return cmp.Compare(a.Type, b.Type)
	case aa == nil && ba != nil:
		return -1
	case aa != nil && ba == nil:
		return 1
	case aa != nil /* && ba != nil */ :
		hasPackage := aa.File.Package != nil && ba.File.Package != nil
		switch {
		case aa.File.Package == nil && ba.File.Package != nil:
			return -1
		case aa.File.Package != nil && ba.File.Package == nil:
			return 1
		case hasPackage && aa.File.Package.Module != ba.File.Package.Module:
			return cmp.Compare(aa.File.Package.Module, ba.File.Package.Module)
		case hasPackage && aa.File.Package != ba.File.Package:
			return cmp.Compare(aa.File.Package.PathInModule, ba.File.Package.PathInModule)
		case aa.File.Name != ba.File.Name:
			return cmp.Compare(aa.File.Name, ba.File.Name)
		case aa.Start.Line != ba.Start.Line:
			return cmp.Compare(aa.Start.Line, ba.Start.Line)
		case aa.Start.Col != ba.Start.Col:
			return cmp.Compare(aa.Start.Col, ba.Start.Col)
		}
	}
	switch {
	case a.Message != b.Message:
		return cmp.Compare(a.Message, b.Message)
	case a.Cause != b.Cause: //nolint:errorlint
		if a.Cause == nil {
			return -1
		} else if b.Cause == nil {
			return 1
		}
		return cmp.Compare(a.Cause.Error(), b.Cause.Error())
	default:
		return 0
	}
}

func typeOrder(t Type) int {
	switch t {
	case InternalError:
		return 0
	case Error:
		return 1
	case Warning:
		return 2
	case Lint:
		return 3
	default:
		return 4
	}
}
