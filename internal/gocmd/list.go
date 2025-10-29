package gocmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type (
	ListAllModulesResult struct {
		MainModule   ListModuleResult
		Dependencies []ListModuleResult
	}

	ListModuleResult struct {
		Path     string            // module path
		Query    string            // version query corresponding to this version
		Version  string            // module version
		Versions []string          // available module versions
		Replace  *ListModuleResult // replaced by this module
		Time     *time.Time        // time version was created
		// Update     *ListModuleResult // available update (with -u)
		Main      bool   // is this the main module?
		Indirect  bool   // module is only indirectly needed by main module
		Dir       string // directory holding local copy of files, if any
		GoMod     string // path to go.mod file describing module, if any
		GoVersion string // go version used in module
		// Retracted  []string          // retraction information, if any (with -retracted or -u)
		// Deprecated string            // deprecation message, if any (with -u)
		Error    *ListModuleError // error loading module
		Sum      string           // checksum for path, version (as in go.sum)
		GoModSum string           // checksum for go.mod (as in go.sum)
		Origin   any              // provenance of module
		// Reuse      bool              // reuse of old module info is safe
	}

	ListModuleError struct {
		Err string // the error itself
	}
)

func (cmd *Cmd) ListAllModules(ctx context.Context, dir string) (*ListAllModulesResult, error) {
	list := cmd.cmd(ctx, "list", "-m", "-json", "all")
	list.Dir = dir

	stdout, err := list.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("go list -m all: getting stdout pipe: %w", err)
	}
	stderr := bytes.NewBuffer(make([]byte, 0, 2048))
	list.Stderr = stderr

	if err := list.Start(); err != nil {
		return nil, fmt.Errorf("go list -m all: starting command: %w", err)
	}

	dec := json.NewDecoder(stdout)

	var result ListAllModulesResult

	for dec.More() {
		var mod ListModuleResult
		if err = dec.Decode(&mod); err != nil {
			if err := list.Wait(); err != nil {
				return nil, formatError("go list -m all", err, stderr.Bytes())
			}
			return nil, fmt.Errorf("go list -m all: parsing dependency module: %w", err)
		}
		if mod.Main {
			result.MainModule = mod
		} else {
			result.Dependencies = append(result.Dependencies, mod)
		}
	}

	if err = list.Wait(); err != nil {
		return nil, formatError("go list -m all", err, stderr.Bytes())
	}

	return &result, nil
}
