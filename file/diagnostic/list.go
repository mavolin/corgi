package diagnostic

import (
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
	switch {
	case a.Type == InternalError && b.Type != InternalError:
		return -1
	case b.Type == InternalError && a.Type != InternalError:
		return 1
	default:
		return 0
	}
}
