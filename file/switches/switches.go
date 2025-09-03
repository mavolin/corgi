// Package switches provides functions for exhaustive switching over the sum
// types in package ast.
//
// The switching functions are generated using go generate.
package switches

//go:generate go test -run=Generate -tags=generate
