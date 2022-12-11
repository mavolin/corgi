package imports

import "fmt"

// CollisionError is the error returned if a namespace collision between two
// imports is detected.
type CollisionError struct {
	// Source is the source of the file that attempted to import the package,
	// but could not because of the conflict in the other file.
	Source string
	// File is the name of the file that attempted to import the package,
	// but could not because of the conflict in the other file.
	File string
	Line int
	Col  int

	// OtherSource is the source of the file that also has an import using the
	// same namespace.
	OtherSource string
	// OtherFile is name of the file that also has an import using the same
	// namespace.
	OtherFile string
	OtherLine int
	OtherCol  int

	// Namespace is the conflicting namespace.
	Namespace string
}

var _ error = (*CollisionError)(nil)

func (e *CollisionError) Error() string {
	return fmt.Sprintf("%s/%s:%d:%d: import namespace `%s` already in use in %s/%s:%d:%d for different import",
		e.Source, e.File, e.Line, e.Col,
		e.Namespace,
		e.OtherSource, e.OtherFile, e.OtherLine, e.OtherCol)
}
