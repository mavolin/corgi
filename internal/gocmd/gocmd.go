package gocmd

import (
	"os/exec"
	"sync"
)

type Cmd struct {
	path string

	goModCacheOnce sync.Once
	goModCache     string
}

// NewCmd creates a new go command using the passed path as the path to the Go
// executable.
func NewCmd(goExecPath string) *Cmd {
	return &Cmd{path: goExecPath}
}

func (c *Cmd) command(subcmd string, args ...string) ([]byte, error) {
	return (&exec.Cmd{Path: c.path, Args: append([]string{c.path, subcmd}, args...)}).Output()
}
