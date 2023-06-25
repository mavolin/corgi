package gocmd

import (
	"os/exec"
	"strings"
	"sync"
)

var path string

func init() {
	var err error
	path, err = exec.LookPath("go")
	if err != nil {
		path = ""
		return
	}

	ver, err := (&exec.Cmd{Path: path, Args: []string{"version"}}).Output()
	if err != nil {
		path = ""
		return
	}

	if !strings.HasPrefix(string(ver), "go version") {
		path = ""
	}
}

type Cmd struct {
	path string

	goModCacheOnce sync.Once
	goModCache     string
}

// NewCmd creates a new go command using the passed path as the path to the Go
// executable.
//
// If goExecPath is empty, the Go executable in the system's PATH is used.
// If there is no Go executable in the PATH, NewCmd returns nil.
func NewCmd(goExecPath string) *Cmd {
	if goExecPath == "" {
		if path == "" {
			return nil
		}

		goExecPath = path
	}

	return &Cmd{path: goExecPath}
}

func (c *Cmd) command(subcmd string, args ...string) ([]byte, error) {
	if path == "" {
		return nil, nil
	}

	return (&exec.Cmd{Path: c.path, Args: append([]string{subcmd}, args...)}).Output()
}

func (c *Cmd) commandIn(workdir string, subcmd string, args ...string) ([]byte, error) {
	if path == "" {
		return nil, nil
	}

	return (&exec.Cmd{Path: c.path, Args: append([]string{subcmd}, args...), Dir: workdir}).Output()
}
