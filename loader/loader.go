package loader

import "github.com/mavolin/corgi/file"

// Loader is an abstraction of where the Linker loads its files from.
//
// It must be concurrently safe.
type Loader interface {
	// LoadLibrary loads the library provided under the passed use path.
	//
	// The link parameter indicates whether this library is loaded as part of
	// linking.
	//
	// It returns any errors it encounters.
	//
	// A return of (nil, nil) is valid and indicates that the loader was unable
	// to find a library with the given path.
	LoadLibrary(usePath string, link bool) (*file.Library, error)
	// LoadInclude loads an include file.
	//
	// It returns any errors it encounters.
	//
	// A return of (nil, nil) is valid and indicates that the loader was unable
	// to find a file that matches.
	LoadInclude(includingFile *file.File, path string) (file.IncludeFile, error)
	// LoadTemplate loads the template file associated with the given
	// extendPath.
	//
	// It returns any errors it encounters.
	//
	// A return of (nil, nil) is valid and indicates that the loader was unable
	// to find a file that matches.
	LoadTemplate(extendPath string) (*file.File, error)
	// LoadMain loads a main file.
	// It gets passed an absolute, forward-slash-separated path, of the
	// directory in which the including file is located.
	//
	// It returns any errors it encounters.
	//
	// A return of (nil, nil) is valid and indicates that the loader was unable
	// to find a file that matches.
	LoadMain(path string) (*file.File, error)
}
