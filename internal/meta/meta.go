// Package meta contains metadata about the compiler.
package meta

import (
	"runtime/debug"
	"strings"
)

const Module = "github.com/mavolin/corgi/v2"

// Version is the version of the binary.
var Version = func() string {
	if buildInfo == nil {
		return DevelVersion
	}

	if buildInfo.Main.Version != "" && buildInfo.Main.Version != "(devel)" {
		return strings.TrimPrefix(buildInfo.Main.Version, "v")
	} else if commit != "" {
		commit := commit
		if len(commit) > 7 {
			commit = commit[:7]
		}
		if dirty {
			commit += ".dirty"
		}

		return DevelVersion + "+" + commit
	}

	return DevelVersion
}()

var commit = func() string {
	if buildInfo == nil {
		return ""
	}

	for _, s := range buildInfo.Settings {
		if s.Key == "vcs.revision" {
			return s.Value
		}
	}
	return ""
}()

var dirty = func() bool {
	if buildInfo == nil {
		return false
	}

	for _, s := range buildInfo.Settings {
		if s.Key == "vcs.modified" && s.Value == "true" {
			return true
		}
	}
	return false
}()

var buildInfo = func() *debug.BuildInfo {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return nil
	}
	return info
}()

// DevelVersion is the version string used for development builds.
const DevelVersion = "devel"
