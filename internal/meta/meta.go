// Package meta contains metadata about the compiler.
package meta

import (
	"runtime/debug"
	"strings"
)

var CLI = false

// develVersion is the version string used for development builds.
const develVersion = "devel"

const Module = "github.com/mavolin/corgi/v2"

// Version is the version of the binary.
var Version = func() string {
	if buildInfo == nil {
		return develVersion
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

		return develVersion + "+" + commit
	}

	return develVersion
}()

var (
	commit = buildInfoSetting("vcs.revision")
	dirty  = buildInfoSetting("vcs.modified") == "true"

	buildInfo = func() *debug.BuildInfo {
		info, ok := debug.ReadBuildInfo()
		if !ok {
			return nil
		}
		return info
	}()
)

func buildInfoSetting(name string) string {
	if buildInfo == nil {
		return ""
	}

	for _, s := range buildInfo.Settings {
		if s.Key == name {
			return s.Value
		}
	}
	return ""
}
