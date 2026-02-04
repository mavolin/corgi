// Package meta contains metadata about the compiler.
package meta

import (
	"runtime/debug"
	"strings"
)

// CLI indicates that the binary being run is the CLI defined in package cmd.
//
// We use this flag to give more concrete internal error messages:
// Internal errors indicate that something is wrong with this module.
// Most internal errors should never occur regardless of context.
// However, for some errors it is ambiguous whether they occurred because of
// a bug in this module, or due to incorrect usage by a consumer of the library.
//
// For the CLI, it is clear that regardless of whether the error arose from a
// bug or by incorrect usage, code in this module is at fault.
// As such, we want an issue to be filed with this module's issue tracker,
// regardless.
//
// However, when consumed as a library, we want to indicate to the consumer
// that the error might have been caused by them, and not necessarily due to a
// bug.
// In such a case, instead of always asking for an issue to be filed, we can
// explain how else the error might have occurred, so the consumer can
// ascertain whether the bug lies with them or with this module.
var CLI = false

const Module = "github.com/mavolin/corgi/v2"

// Version is the version of the binary.
var Version = func() string {
	const develVersion = "devel"

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
