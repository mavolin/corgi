// Package charref provides utilities for handling HTML character references.
//
// (Legacy) characters without a semicolon are not supported.
package charref

//go:generate go run github.com/mavolin/corgi/v2/tools/codegen/charrefexport

// Is returns true, if name is a named character reference.
//
// name must be stripped of the leading '&' and the trailing ';'.
func Is(name string) bool {
	return Chars(name) != ""
}
