package load

import "github.com/mavolin/corgi/file"

// Loader is an abstraction of where the Linker loads its files from.
//
// It must be concurrently safe.
type Loader interface {
	// LoadLibrary loads the library provided under the passed module path.
	//
	// It gets passed the file that uses it.
	//
	// If usingFile is nil, the library is loaded standalone and path should be
	// interpreted as an absolute path instead of a module path.
	// This happens, for example, when pre-compiling a library.
	//
	// A return of (nil, nil) is valid and indicates that the loader was unable
	// to find a library with the given path.
	LoadLibrary(usingFile *file.File, usePath string) (*file.Library, error)
	// LoadInclude loads the file available under the slash-separated path,
	// relative to the passed includingFile.
	//
	// A return of (nil, nil) is valid and indicates that the load was unable
	// to find a file that matches.
	LoadInclude(includingFile *file.File, path string) (file.IncludeFile, error)
	// LoadTemplate loads the template file associated with the given
	// module path.
	//
	// A return of (nil, nil) is valid and indicates that the load was unable
	// to find a file that matches.
	LoadTemplate(extendingFile *file.File, extendPath string) (*file.File, error)
	// LoadMain loads a main file available under the passed system path.
	//
	// A return of (nil, nil) is valid and indicates that the load was unable
	// to find a file that matches.
	LoadMain(path string) (*file.File, error)
}
