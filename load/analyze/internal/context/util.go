package context

func IsKeyword(w string) bool {
	switch w {
	// https://go.dev/ref/spec#Keywords
	case "break", "case", "chan", "const", "continue",
		"default", "defer", "else", "fallthrough", "for",
		"func", "go", "goto", "if", "import",
		"interface", "map", "package", "range", "return",
		"select", "struct", "switch", "type", "var":
		return true
	default:
		return false
	}
}
