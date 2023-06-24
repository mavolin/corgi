package loader

import "github.com/mavolin/corgi/file"

// BasicLoader is a [Loader] implementation that allows configuration of every
// step of the loading process.
type BasicLoader struct {
	// MainReader reads and returns the main file located under the passed
	// module path.
	//
	// A return of (nil, nil) is valid and indicates the file wasn't found.
	MainReader func(path string) (*File, error)
	// TemplateReader reads and returns the template file located under the
	// passed module path.
	//
	// A return of (nil, nil) is valid and indicates the file wasn't found.
	TemplateReader func(path string) (*File, error)
	// IncludeReader reads and returns the file located under the passed path.
	//
	// It gets passed the file that includes it to help resolve path correctly.
	//
	// A return of (nil, nil) is valid and indicates the file wasn't found.
	IncludeReader func(includingFile *file.File, path string) (*File, error)
	// PrecompiledLibraryReader attempts to load the precompiled library
	// available under the passed use path.
	//
	// If there is no such library or the library is not precompiled, the
	// function should return (nil, nil).
	//
	// It is guaranteed that before every call to LibraryReader,
	// PrecompiledLibraryReader will be called, unless PrecompiledLibraryReader
	// is nil.
	//
	// Naturally, a precompiled library won't be parsed, pre-link validated, or
	// linked.
	// However, the LibraryValidator will be called on the library.
	PrecompiledLibraryReader func(path string) (*file.Library, error)
	// LibraryReader reads and returns the files of the library available under
	// the passed use path
	//
	// It will only be called, if PrecompiledLibraryReader returned (nil, nil).
	//
	// A return of (nil, nil) is valid and indicates the library wasn't found.
	LibraryReader func(path string) (*Library, error)

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
	Name       string
	Module     string
	ModulePath string

	AbsolutePath string

	// IsCorgi indicates whether this is a corgi file, or not.
	//
	// Only relevant for include files.
	IsCorgi bool
	Raw     []byte
}

type Library struct {
	Module     string
	ModulePath string

	AbsolutePath string

	Files []File
}

func (b BasicLoader) LoadLibrary(usePath string, link bool) (*file.Library, error) {
	lib, err := b.PrecompiledLibraryReader(usePath)
	if err != nil {
		return lib, err
	}
	if lib != nil {
		if link {
			if err := b.LibraryValidator(lib); err != nil {
				return lib, err
			}
		}

		return lib, nil
	}

	libw, err := b.LibraryReader(usePath)
	if err != nil {
		return nil, err
	}
	if libw == nil {
		return nil, nil
	}

	lib = &file.Library{
		Module:       libw.Module,
		ModulePath:   libw.ModulePath,
		AbsolutePath: libw.AbsolutePath,
		Precompiled:  false,
		Files:        make([]file.File, 0, len(libw.Files)),
	}

	for _, fw := range libw.Files {
		f, err := b.Parser(fw.Raw)
		if f == nil {
			f = new(file.File)
		}
		f.Name = fw.Name
		f.Module = fw.Module
		f.ModulePath = fw.ModulePath
		f.AbsolutePath = fw.AbsolutePath
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

		lib.Files = append(lib.Files, *f)
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
	f.Name = fw.Name
	f.Module = fw.Module
	f.ModulePath = fw.ModulePath
	f.AbsolutePath = fw.AbsolutePath
	f.Raw = string(fw.Raw)
	cincl := file.CorgiInclude{File: f}
	if err != nil {
		return cincl, err
	}

	if err = b.PreLinkValidator(f); err != nil {
		return cincl, err
	}

	if err = b.Linker(f); err != nil {
		return cincl, err
	}

	return cincl, nil
}

func (b BasicLoader) LoadTemplate(extendPath string) (*file.File, error) {
	return b.load(extendPath, b.TemplateReader)
}

func (b BasicLoader) LoadMain(path string) (*file.File, error) {
	f, err := b.load(path, b.MainReader)
	if err != nil {
		return f, err
	}
	if f == nil {
		return nil, nil
	}

	if err = b.MainValidator(f); err != nil {
		return f, err
	}

	return f, nil
}

func (b BasicLoader) load(path string, reader func(string) (*File, error)) (*file.File, error) {
	fw, err := reader(path)
	if err != nil {
		return &file.File{
			Name:         fw.Name,
			Module:       fw.Module,
			ModulePath:   fw.ModulePath,
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
	f.Name = fw.Name
	f.Module = fw.Module
	f.ModulePath = fw.ModulePath
	f.AbsolutePath = fw.AbsolutePath
	f.Raw = string(fw.Raw)
	if err != nil {
		return f, err
	}

	if err = b.PreLinkValidator(f); err != nil {
		return f, err
	}

	if err = b.Linker(f); err != nil {
		return f, err
	}

	return f, nil
}
