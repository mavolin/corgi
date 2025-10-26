// Package safe provides types that hold data that is trusted to be used inside
// HTML documents.
// Additionally, it provides functions to create safe instances of these types
// using compile-time constants or other safe approaches.
//
// To create instances of these types, use the Trusted* functions.
//
// This package follows some of the designs from safehtml:
// https://pkg.go.dev/github.com/google/safehtml
package safe

// Replacement is the string to be used as replacement for an unsafe value.
const Replacement = "ZreplacementZ"

type constant string
