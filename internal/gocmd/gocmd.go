package gocmd

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
)

type Cmd struct {
	path string
}

// New creates a new go command using the passed path as the path to the Go
// executable.
func New(goExecPath string) *Cmd {
	return &Cmd{path: goExecPath}
}

func (cmd *Cmd) cmd(ctx context.Context, subcmd string, args ...string) *exec.Cmd {
	args2 := make([]string, len(args)+1)
	args2[0] = subcmd
	copy(args2[1:], args)
	return exec.CommandContext(ctx, cmd.path, args2...) // #nosec G204
}

func formatError(cmd string, err error, stderr []byte) error {
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return fmt.Errorf("%s: %w: %s", cmd, err, string(stderr[:min(len(stderr), 256)]))
	}
	return fmt.Errorf("%s: %w", cmd, err)
}
