// Package charref provides utilities for handling HTML character references.
//
// (Legacy) characters without a semicolon are not supported.
package charref

// Is returns true, if name is a named character reference.
//
// name must be stripped of the leading '&' and the trailing ';'.
func Is(name string) bool {
	return Chars(name) != ""
}
