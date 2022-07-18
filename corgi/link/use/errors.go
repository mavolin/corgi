package use

import "fmt"

type NamespaceError struct {
	Source string
	File   string
	Line   int

	OtherLine int

	Namespace string
}

var _ error = (*NamespaceError)(nil)

func (e *NamespaceError) Error() string {
	return fmt.Sprintf("%s/%s:%d: namespace collision with `use` in line %d",
		e.Source, e.File, e.Line, e.OtherLine)
}
