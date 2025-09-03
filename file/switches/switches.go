// Package switches provides functions for exhaustive switching over the sum
// types in package ast.
//
// The switching functions are generated using go generate.
//
// Since the sum types may evolve over time, getting extended with new types,
// this package is expressly exempt from the usual stability guarantee.
// More precisely, a minor version increment allows a breaking change in the
// signatures of the switching functions.
// This, of course, is intended breakage to warn users relying on an exhaustive
// switch that they need to handle a new type.
// All other compatibility guarantees remain untouched.
// Patch version increments do not allow breaking changes.
package switches

//go:generate go test -run=Generate -tags=generate
