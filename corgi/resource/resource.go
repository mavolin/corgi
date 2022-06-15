// Package resource provides abstractions for accessing resources.
package resource

// Extension is the file extension for corgi files.
const Extension = ".corgi"

type (
	// Source is an interface that allows access to resources providing
	// extended, included, and used files.
	Source interface {
		// ReadCorgiFile returns the requested Corgi file.
		//
		// Name's file extension is optional and might not be present.
		//
		// If no file is found, ReadCorgiFile should return nil, nil.
		ReadCorgiFile(name string) (*File, error)
		// ReadCorgiLib returns the library files that are available under the
		// given name.
		// That name might either point to a directory of corgi files, or to a
		// single file.
		//
		// When accessing a file, the file extension is optional and might not
		// be present.
		//
		// If both a file and a directory with the same name exists,
		// directories should take precedence, e.g. if name is foo and a
		// directory foo and a file foo.corgi exists, the dir should be chosen.
		//
		// If no file is found, ReadCorgiLib should return nil, nil.
		ReadCorgiLib(name string) ([]File, error)
		// ReadFile reads the given file from the resource source.
		//
		// The file extension is optional and might not be present.
		// If name has no extension and a file with the given name and a corgi
		// file with the name + extension exists, then the file without
		// extension should be read.
		//
		// If no file is found, ReadFile should return nil, nil.
		ReadFile(name string) (*File, error)
	}

	// File represents a read resource file.
	File struct {
		// Name is the path to the file, relative to the resource directory it
		// is located in.
		//
		// It includes the file's file extension.
		Name string
		// Source is the name of the source.
		// It should be unique to the source.
		// Ideally, it should be the path to the source root, without a
		// trailing slash.
		//
		// It must be the same for all files returned by the source.
		Source string
		// Contents are the contents of the file.
		Contents string
	}
)
