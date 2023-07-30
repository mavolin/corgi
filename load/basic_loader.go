package load

import (
	"errors"

	"github.com/mavolin/corgi/file"
)

// BasicLoader is a [Loader] implementation that allows configuration of every
// step of the loading process.
type BasicLoader struct {
	// MainReader reads and returns the main file located under the passed path.
	//
	// A return of (nil, nil) is valid and indicates the file wasn't found.
	MainReader func(path string) (*File, error)
	// TemplateReader reads and returns the template file located under the
	// passed module path.
	//
	// A return of (nil, nil) is valid and indicates the file wasn't found.
	TemplateReader func(extendingFile *file.File, path string) (*File, error)
	// IncludeReader reads and returns the file located under the passed path.
	//
	// It gets passed the file that includes it to help resolve path correctly.
	//
	// A return of (nil, nil) is valid and indicates the file wasn't found.
	IncludeReader func(includingFile *file.File, path string) (*File, error)
	// LibraryReader reads and returns the files of the library available under
	// the passed use path.
	//
	// It gets passed the file that uses it to help resolve version correctly.
	// If usingFile is nil, the library is loaded standalone and path should be
	// interpreted as an absolute path instead of a module path.
	//
	// A return of (nil, nil) is valid and indicates the library wasn't found.
	LibraryReader func(usingFile *file.File, path string) (*Library, error)
	// DirLibraryLoader loads the library in the directory of f, a
	// main, include, or template file, which is not yet linked or validated.
	//
	// A return of (nil, nil) is valid and indicates that there are no library
	// files in f's directory.
	DirLibraryLoader func(f *file.File) (*file.Library, error)

	// Parser is the function used to parse a file.
	Parser func(in []byte) (*file.File, error)
	// PreLinkValidator is the validator called after parsing and before
	// linking.
	PreLinkValidator func(*file.File) error
	// Linker links the given file recursively.
	//
	// Commonly, it utilizes this Loader to load the linkLibraries of f, f's
	// includes, and the template of f, if it has one.
	Linker func(*file.File) error

	// FileValidator is the validator function called if a main file is loaded.
	//
	// It is expected that the function recursively validates all files the
	// passed file depends on.
	MainValidator func(*file.File) error
	// LibraryValidator is the validator called if a library is loaded with
	// link set to false.
	//
	// It is expected that the function recursively validates all files the
	// passed library depends on.
	LibraryValidator func(*file.Library) error
}

var _ Loader = BasicLoader{}

type File struct {
	Name         string
	Module       string
	PathInModule string

	AbsolutePath string

	// IsCorgi indicates whether this is a corgi file, or not.
	//
	// Only relevant for include files.
	IsCorgi bool
	Raw     []byte
}

type Library struct {
	Module       string
	PathInModule string

	AbsolutePath string

	Files       []File
	Precompiled *file.Library // if set, no other fields need to be set
}

func (b BasicLoader) LoadLibrary(usingFile *file.File, usePath string) (*file.Library, error) {
	libw, err := b.LibraryReader(usingFile, usePath)
	if err != nil {
		return nil, err
	}
	if libw == nil {
		return nil, nil
	}

	if libw.Precompiled != nil {
		return libw.Precompiled, nil
	}

	lib := &file.Library{
		Module:       libw.Module,
		PathInModule: libw.PathInModule,
		AbsolutePath: libw.AbsolutePath,
		Precompiled:  false,
		Files:        make([]*file.File, 0, len(libw.Files)),
	}

	for _, fw := range libw.Files {
		f, err := b.Parser(fw.Raw)
		if f == nil {
			f = new(file.File)
		}
		f.Type = file.TypeLibraryFile
		f.Name = fw.Name
		f.Module = fw.Module
		f.PathInModule = fw.PathInModule
		f.AbsolutePath = fw.AbsolutePath
		f.Library = lib
		f.Raw = string(fw.Raw)
		if err != nil {
			return lib, err
		}

		if err = b.PreLinkValidator(f); err != nil {
			return lib, err
		}

		if err = b.Linker(f); err != nil {
			return lib, err
		}

		lib.Files = append(lib.Files, f)
	}

	if usingFile == nil {
		if err := b.LibraryValidator(lib); err != nil {
			return nil, err
		}
	}

	return lib, nil
}

func (b BasicLoader) LoadInclude(includingFile *file.File, name string) (file.IncludeFile, error) {
	fw, err := b.IncludeReader(includingFile, name)
	if err != nil {
		return nil, err
	}
	if fw == nil {
		return nil, nil
	}

	if !fw.IsCorgi {
		return file.OtherInclude{Contents: string(fw.Raw)}, nil
	}

	f, err := b.Parser(fw.Raw)
	if f == nil {
		f = new(file.File)
	}
	f.Type = file.TypeInclude
	f.Name = fw.Name
	f.Module = fw.Module
	f.PathInModule = fw.PathInModule
	f.AbsolutePath = fw.AbsolutePath
	f.Raw = string(fw.Raw)

	var dirLibErr error
	f.DirLibrary, dirLibErr = b.DirLibraryLoader(f)

	cincl := file.CorgiInclude{File: f}
	if err != nil {
		return cincl, errors.Join(err, dirLibErr)
	}

	if err = b.PreLinkValidator(f); err != nil {
		return cincl, errors.Join(err, dirLibErr)
	}

	if err = b.Linker(f); err != nil {
		return cincl, errors.Join(err, dirLibErr)
	}

	return cincl, nil
}

func (b BasicLoader) LoadTemplate(extendingFile *file.File, extendPath string) (*file.File, error) {
	fw, err := b.TemplateReader(extendingFile, extendPath)
	if err != nil {
		return &file.File{
			Name:         fw.Name,
			Module:       fw.Module,
			PathInModule: fw.PathInModule,
			AbsolutePath: fw.AbsolutePath,
			Raw:          string(fw.Raw),
		}, err
	}
	if fw == nil {
		return nil, nil
	}

	f, err := b.Parser(fw.Raw)
	if f == nil {
		f = new(file.File)
	}
	f.Type = file.TypeTemplate
	f.Name = fw.Name
	f.Module = fw.Module
	f.PathInModule = fw.PathInModule
	f.AbsolutePath = fw.AbsolutePath
	f.Raw = string(fw.Raw)

	var dirLibErr error
	f.DirLibrary, dirLibErr = b.DirLibraryLoader(f)

	if err != nil {
		return f, errors.Join(err, dirLibErr)
	}

	if err = b.PreLinkValidator(f); err != nil {
		return f, errors.Join(err, dirLibErr)
	}

	if err = b.Linker(f); err != nil {
		return f, errors.Join(err, dirLibErr)
	}

	return f, nil
}

func (b BasicLoader) LoadMain(path string) (*file.File, error) {
	fw, err := b.MainReader(path)
	if err != nil {
		return &file.File{
			Name:         fw.Name,
			Module:       fw.Module,
			PathInModule: fw.PathInModule,
			AbsolutePath: fw.AbsolutePath,
			Raw:          string(fw.Raw),
		}, err
	}
	if fw == nil {
		return nil, nil
	}

	f, err := b.Parser(fw.Raw)
	if f == nil {
		f = new(file.File)
	}
	f.Type = file.TypeMain
	f.Name = fw.Name
	f.Module = fw.Module
	f.PathInModule = fw.PathInModule
	f.AbsolutePath = fw.AbsolutePath
	f.Raw = string(fw.Raw)

	var dirLibErr error
	f.DirLibrary, dirLibErr = b.DirLibraryLoader(f)

	if err != nil {
		return f, errors.Join(err, dirLibErr)
	}

	if err = b.PreLinkValidator(f); err != nil {
		return f, errors.Join(err, dirLibErr)
	}

	if err = b.Linker(f); err != nil {
		return f, errors.Join(err, dirLibErr)
	}

	if err = b.MainValidator(f); err != nil {
		return f, errors.Join(err, dirLibErr)
	}

	return f, dirLibErr
}
