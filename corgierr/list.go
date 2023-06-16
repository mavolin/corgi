package corgierr

import "strings"

// List represents a collection of [Error] objects.
type List []*Error //nolint:errname

// Error calls [Error.Error] for each item in the list, separating them by
// newlines.
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

// Pretty calls [Error.Pretty] on each item in the list, separating each error
// by two newlines (effectively leaving a single blank line in-between errors).
func (l List) Pretty(o PrettyOptions) string {
	var sb strings.Builder
	for i, err := range l {
		if i > 0 {
			sb.WriteString("\n\n")
		}

		sb.WriteString(err.Pretty(o))
	}

	return sb.String()
}
