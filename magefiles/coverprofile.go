package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/mavolin/corgi/v2/internal/meta"
)

type path = string

// Coverprofile generates a coverage profile for the entire module, excluding
// generated files.
func Coverprofile(outPath string) error {
	isGeneratedFile, err := makeGeneratedFileTable()
	if err != nil {
		return err
	}

	cmd := exec.Command("go", "test", "-coverprofile", outPath, "./...")
	cmd.Stderr = os.Stderr
	if err = cmd.Run(); err != nil {
		return fmt.Errorf("generating coverprofile: %w", err)
	}

	profile, err := os.ReadFile(outPath)
	if err != nil {
		return fmt.Errorf("reading coverprofile: %w", err)
	}

	strippedProfile := stripGeneratedFiles(profile, isGeneratedFile)
	if err = os.WriteFile(outPath, strippedProfile, 0o600); err != nil {
		return fmt.Errorf("writing coverprofile: %w", err)
	}
	return nil
}

func stripGeneratedFiles(profile []byte, isGeneratedFile map[path]bool) []byte {
	out := make([]byte, 0, len(profile)+len("\n"))

	lines := strings.Split(string(profile), "\n")
	for _, line := range lines {
		p, _, ok := strings.Cut(strings.TrimPrefix(line, meta.Module+"/"), ":")
		if ok && isGeneratedFile[p] {
			continue
		}
		out = append(out, line...)
		out = append(out, '\n')
	}
	return out[:len(out)-len("\n")]
}

func makeGeneratedFileTable() (map[path]bool, error) {
	commentRegexp := regexp.MustCompile(`^\w*// Code generated .* DO NOT EDIT\.$`)

	ps := make(map[path]bool)
	err := filepath.WalkDir(".", func(p path, d os.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("%s: %w", p, err)
		} else if !d.Type().IsRegular() {
			return nil
		}

		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}

		if commentRegexp.Match(data) {
			ps[p] = true
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("finding generated files: %w", err)
	}
	return ps, nil
}
