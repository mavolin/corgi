package diagnostic

import "strings"

type List []*Diagnostic

func (l List) ToError() error {
	// todo: implement
	panic("implement me")
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
