package resource

// NotFoundError is returned when a resource file could not be found.
//
// It is never returned by the Source directly.
type NotFoundError struct {
	Name string
}

func (e *NotFoundError) Error() string {
	return "could not find resource file '" + e.Name + "' in any of the resource directories"
}
