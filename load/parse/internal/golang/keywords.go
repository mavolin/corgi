package golang

// IsKeyword returns whether the given string is a Go keyword.
func IsKeyword(s string) bool { // https://go.dev/ref/spec#Keywords
	switch s {
	case "break":
	case "case":
	case "chan":
	case "const":
	case "continue":
	case "default":
	case "defer":
	case "else":
	case "fallthrough":
	case "for":
	case "func":
	case "go":
	case "goto":
	case "if":
	case "import":
	case "interface":
	case "map":
	case "package":
	case "range":
	case "return":
	case "select":
	case "struct":
	case "switch":
	case "type":
	case "var":
	default:
		return false
	}
	return true
}
