package file

// IsExported reports whether the given identifier, assumed to be valid, would
// be exported.
func IsExported(s Identifier) bool {
	if len(s) == 0 {
		return false
	}
	return 'A' <= s[0] && s[0] <= 'Z'
}
