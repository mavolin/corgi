// Package meta contains metadata about the compiler.
package meta

import (
	"runtime/debug"
	"strings"
)

// Version is the version of the binary.
//
// This should be set during compilation using
// `-ldflags "-X github.com/mavolin/corgi/internal/meta.Version=1.2.3"`.
var Version = DevelopVersion

// DevelopVersion is the version string used for development builds.
const DevelopVersion = "devel"

func init() {
	// If corgi was installed using 'go install' (as opposed to downloaded
	// from GitHub Releases), we have no release information.
	// In that case we can read version and commit from the build info.
	if Version != DevelopVersion {
		return
	}

	i, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}

	if i.Main.Version != "" && i.Main.Version != "(devel)" {
		Version = strings.TrimPrefix(i.Main.Version, "v")
		return
	}

	for _, s := range i.Settings {
		if s.Key == "vcs.revision" {
			if s.Value != "" {
				Version += "-" + s.Value
			}
			return
		}
	}
}
