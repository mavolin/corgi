package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
)

func NeedRegenerate() error {
	temp, err := os.MkdirTemp("", "need_regenerate")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer func() { _ = os.RemoveAll(temp) }()

	// use cp instead of os.CopyFS so we don't get problems with file permissions
	cpCmd := exec.Command("cp", "-r", "./.", temp)
	cpCmd.Stderr = os.Stderr
	if err = cpCmd.Run(); err != nil {
		return fmt.Errorf("copy to temp dir: %w", err)
	}

	genCmd := exec.Command("go", "generate", "./...")
	genCmd.Dir = temp
	genCmd.Stdout, genCmd.Stderr = os.Stdout, os.Stderr
	if err = genCmd.Run(); err != nil {
		return fmt.Errorf("go generate: %w", err)
	}

	diffCmd := exec.Command("git", "diff", "--exit-code", "--no-index", ".", temp)
	diffCmd.Stderr = os.Stderr
	diffOut, err := diffCmd.Output()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && exitErr.ExitCode() == 1 {
			fmt.Println(string(diffOut))
			return fmt.Errorf("generated files are out of date, run 'go generate ./...' and commit the changes")
		}
		return fmt.Errorf("git diff: %w", err)
	}

	return nil
}
