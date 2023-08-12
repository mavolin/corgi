package gomod

import (
	"os"
	"path/filepath"

	"golang.org/x/mod/modfile"
)

// Find attempts to find the go.mod file governing this directory by first
// searching for it in dir and then in dir's parents.
//
// dir must be an absolute path in using the system's path separator.
//
// If it finds it, Find returns the module and the absolute path to it.
func Find(dir string) (*modfile.File, string, error) {
	var err error
	dir, err = filepath.Abs(dir)
	if err != nil {
		return nil, "", err
	}

	for {
		p := filepath.Join(dir, "go.mod")
		f, err := os.ReadFile(p)
		if err != nil {
			if len(dir) <= 1 {
				return nil, "", nil
			}

			dir = filepath.Dir(dir)
			continue
		}

		mod, err := modfile.ParseLax(p, f, nil)
		return mod, p, err
	}
}
