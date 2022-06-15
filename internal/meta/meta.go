// Package meta contains metadata about the compiler.
package meta

// Version is the version of the binary.
//
// This should be set during compilation using
// `-ldflags "-X github.com/mavolin/corgi/internal/meta.Version=1.2.3"`.
var Version = "develop"
