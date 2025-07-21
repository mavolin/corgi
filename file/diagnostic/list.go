package diagnostic

import (
	"cmp"
	"slices"
	"strings"
)

type List []*Diagnostic

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
			sb.WriteString("\n\n\n")
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

	switch {
	case a.Type == InternalError && b.Type != InternalError:
		return -1
	case a.Type != InternalError && b.Type == InternalError:
		return 1
	case aa == nil && ba != nil:
		return -1
	case aa != nil && ba == nil:
		return 1
	case aa == nil /* && ba == nil */ :
		return 0
	case aa.File == ba.File && aa.Start.Line == ba.Start.Line:
		return cmp.Compare(aa.Start.Col, ba.Start.Col)
	case aa.File == ba.File:
		return cmp.Compare(aa.Start.Line, ba.Start.Line)
	case aa.File.Package == ba.File.Package:
		return cmp.Compare(aa.File.Name, ba.File.Name)
	case aa.File.Package.Module == ba.File.Package.Module:
		return cmp.Compare(aa.File.Package.PathInModule, ba.File.Package.PathInModule)
	default:
		return cmp.Compare(aa.File.Package.Module, ba.File.Package.Module)
	}
}
