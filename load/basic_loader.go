package load

import (
	"errors"
	"path"
	"path/filepath"
	"sync"

	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/link"
	"github.com/mavolin/corgi/parse"
	"github.com/mavolin/corgi/validate"
)

// BasicLoader is a [Loader] implementation that allows configuration of every
// step of the loading process.
type BasicLoader struct {
	// FileReader reads the file at the passed path and returns it.
	FileReader func(path string) (*File, error)
	// PackageReader reads all corgi files in the directory at the passed path
	// and returns them.
	//
	// A return of a package with no files is valid and indicates the package
	// contains no corgi files.
	PackageReader func(path string) (*Package, error)
	// ImportReader reads the package specified by the passed import path and
	// returns it.
	//
	// A return of a package with no files is valid and indicates the package
	// contains no corgi files.
	ImportReader func(path string) (*Package, error)

	linker     *link.Linker
	linkerOnce sync.Once
}

var _ Loader = (*BasicLoader)(nil)

type File struct {
	Name         string
	Module       string
	PathInModule string
	AbsolutePath string

	Raw []byte
}

type Package struct {
	Module       string
	PathInModule string
	AbsolutePath string

	Files []File
}

func (l *BasicLoader) LoadPackage(path string) (*file.Package, error) {
	rawPkg, pkgErr := l.PackageReader(path)
	p, loadErr := l.loadPackage(rawPkg)
	return p, errors.Join(pkgErr, loadErr)
}

func (l *BasicLoader) LoadFiles(paths ...string) (*file.Package, error) {
	errs := make([]error, 0, len(paths)+1)

	files := make([]File, 0, len(paths))
	for _, filePath := range paths {
		f, err := l.FileReader(filePath)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		files = append(files, *f)
	}

	rawPkg := &Package{Files: files}
	if len(files) > 0 {
		f := files[0]
		rawPkg.Module = f.Module
		rawPkg.PathInModule = path.Dir(f.PathInModule)
		rawPkg.AbsolutePath = filepath.Dir(f.AbsolutePath)
	}

	p, err := l.loadPackage(rawPkg)
	if err != nil {
		errs = append(errs, err)
	}
	return p, errors.Join(errs...)
}

func (l *BasicLoader) LoadImport(path string) (*file.Package, error) {
	rawPkg, pkgErr := l.ImportReader(path)
	p, loadErr := l.loadPackage(rawPkg)
	return p, errors.Join(pkgErr, loadErr)
}

func (l *BasicLoader) ParseFile(path string) (*file.File, error) {
	rawFile, err := l.FileReader(path)
	if err != nil {
		return nil, err
	}

	f, err := parse.Parse(rawFile.Raw)
	if f == nil {
		f = new(file.File)
	}
	f.Name = rawFile.Name
	f.Module = rawFile.Module
	f.PathInModule = rawFile.PathInModule
	f.AbsolutePath = rawFile.AbsolutePath
	return f, err
}

func (l *BasicLoader) loadPackage(rawPkg *Package) (*file.Package, error) {
	if rawPkg == nil {
		return nil, nil
	}

	p := &file.Package{
		Module:       rawPkg.Module,
		PathInModule: rawPkg.PathInModule,
		AbsolutePath: rawPkg.AbsolutePath,
		Files:        make([]*file.File, len(rawPkg.Files)),
	}

	parseErrs := make([]error, 0, len(rawPkg.Files))

	for i, rawFile := range rawPkg.Files {
		f, err := parse.Parse(rawFile.Raw)

		if f == nil {
			f = new(file.File)
		}
		f.Name = rawFile.Name
		f.Module = rawFile.Module
		f.PathInModule = rawFile.PathInModule
		f.AbsolutePath = rawFile.AbsolutePath
		f.Package = p
		p.Files[i] = f

		if err != nil {
			parseErrs = append(parseErrs, err)
			continue
		}
	}

	p.Info = file.AnalyzePackage(p)

	linkErr := l.link(p)
	valErr := validate.Package(p)

	// linking errors are probably caused by parser errors, so
	// return parser errors first
	if len(parseErrs) > 0 {
		if valErr != nil {
			parseErrs = append(parseErrs, valErr)
		}
		return p, errors.Join(parseErrs...)
	}

	return p, errors.Join(linkErr, valErr)
}

func (l *BasicLoader) link(p *file.Package) error {
	l.linkerOnce.Do(func() {
		l.linker = link.New(l)
	})

	return l.linker.Link(p)
}
