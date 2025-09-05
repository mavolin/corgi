package parse

import (
	"os"
	"path/filepath"
	"testing"
)

// FuzzParser tests that the parser doesn't crash or hang on arbitrary input.
func FuzzParser(f *testing.F) {
	// Read all .corgi files in the project
	corgiFiles, err := recursivelyReadAll("../../", "*.corgi")
	if err != nil {
		f.Fatalf("failed to read .corgi files: %v", err)
	}

	for _, file := range corgiFiles {
		data, err := os.ReadFile(file)
		if err != nil {
			f.Fatalf("failed to read file %s: %v", file, err)
		}
		f.Add(string(data))
	}

	f.Add("")
	f.Add("package")
	f.Add("package main")
	f.Add("package main\n\nimport \"std\"")
	f.Add("comp Woof() {}")
	f.Add("comp Woof() {\n\tdiv [ Hello World! ]\n}")
	f.Add(":Bark(woof: 12)")
	f.Add("span")
	f.Add("div [ Hello World! ]")

	f.Fuzz(func(_ *testing.T, data string) {
		// We're just checking that Parse doesn't panic, so we ignore the return values
		_, _ = Parse(data, Options{})
	})
}

func recursivelyReadAll(dir string, pattern string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
			return err
		} else if !matched {
			return nil
		}
		files = append(files, path)
		return nil
	})
	return files, err
}
