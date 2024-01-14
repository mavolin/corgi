package load

import (
	"github.com/mavolin/corgi/file"
)

// Loader is an abstraction of where the Linker loads its files from.
//
// It must be concurrently safe.
type Loader interface {
	// LoadPackage loads all corgi files in the directory at the passed system
	// path and combines them into a single package.
	//
	// LoadPackage may return a package with no files, indicating that the
	// package was found, but it contained no corgi files.
	LoadPackage(path string) (*file.Package, error)
	// LoadFiles loads all named files and combines them into a single package.
	LoadFiles(paths ...string) (*file.Package, error)
	// LoadImport loads the package specified by the passed import path.
	//
	// LoadImport may return a package with no files, indicating that the
	// package was found, but it contained no corgi files.
	LoadImport(path string) (*file.Package, error)
}
