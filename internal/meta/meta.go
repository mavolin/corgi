// Package meta contains metadata about the compiler.
package meta

import (
	"runtime/debug"
)

// Version is the version of the binary.
//
// This should be set during compilation using
// `-ldflags "-X github.com/mavolin/corgi/internal/meta.Version=v1.2.3"`.
var Version = DevelopVersion

// DevelopVersion is the version string used for development builds.
const DevelopVersion = "devel"

// Commit is the commit hash of the binary.
//
// This should be set during compilation using
// `-ldflags "-X github.com/mavolin/corgi/internal/meta.Commit=abc123"`.
var Commit = UnknownCommit

// UnknownCommit is the placeholder used for Commit if there is no information
// about the current commit.
const UnknownCommit = "unknown commit"

func init() {
	// If corgi was installed using 'go install' (as opposed to downloaded
	// from GitHub Releases), we have no release information.
	// In that case we can read version and commit from the build info.

	i, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}

	if Version == DevelopVersion && i.Main.Version != "" && i.Main.Version != "(devel)" {
		Version = i.Main.Version
	}

	if Commit == UnknownCommit {
		for _, s := range i.Settings {
			if s.Key != "vcs.revision" {
				continue
			}

			if s.Value != "" {
				Commit = s.Value
			}

			break
		}
	}
}
