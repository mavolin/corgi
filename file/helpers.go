package file

type ContainingElement struct {
	Component *Component
	Element   *ElementReference
}

// IsExported reports whether the given identifier, assumed to be valid, would
// be exported.
func IsExported(s string) bool {
	if len(s) == 0 {
		return false
	}
	return 'A' <= s[0] && s[0] <= 'Z'
}
