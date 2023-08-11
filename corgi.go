// Package corgi provides parsing for corgi files.
package corgi

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/exp/slog"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"

	"github.com/mavolin/corgi/file"
	"github.com/mavolin/corgi/file/precomp"
	"github.com/mavolin/corgi/file/typeinfer"
	"github.com/mavolin/corgi/internal/gocmd"
	"github.com/mavolin/corgi/internal/gomod"
	"github.com/mavolin/corgi/link"
	"github.com/mavolin/corgi/load"
	"github.com/mavolin/corgi/parse"
	"github.com/mavolin/corgi/std"
	"github.com/mavolin/corgi/validate"
)

const (
	PrecompFileName = "lib.precorgi"
	Ext             = ".corgi"
	LibExt          = ".corgil"
)

var ErrNotExists = errors.New("file does not exist")

// loader is responsible for loading, i.e. reading, parsing, validating, and
// linking corgi files, like the CLI does.
//
// It keeps state about the module data of loaded files, more concretely
// results of calling `go mod edit -json` used to resolve import paths.
//
// It is concurrently safe, but only meant to be used once.
type loader struct {
	// mainMod is the go.mod of the main or library file being loaded.
	mainMod       *modfile.File
	mainModSysAbs string

	noPrecompile bool

	loader load.Loader
	linker *link.Linker
	cmd    *gocmd.Cmd
	log    *slog.Logger
}

type LoadOptions struct {
	// GoExecPath is the path to the go binary to used.
	//
	// If not set, $GOROOT/bin/go will be used.
	GoExecPath string

	// NoPrecompile forces the loader to always read corgi files instead of
	// loading a precompiled file.
	NoPrecompile bool

	// Logger is used to log the individual steps of the logging process.
	//
	// If left as nil, nothing will be logged
	Logger *slog.Logger
}

var nopLog = slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError + 1}))

func newLoader(o LoadOptions) (*loader, error) {
	var l loader
	l.noPrecompile = o.NoPrecompile
	bl := &load.BasicLoader{
		MainReader:     l.readMain,
		TemplateReader: l.readTemplate,
		IncludeReader:  l.readInclude,
		LibraryReader:  l.readLibrary,
		Parser: func(in []byte) (*file.File, error) {
			log := l.log.WithGroup("parser")

			log.Info("parsing")
			f, err := parse.Parse(in)
			if err != nil {
				log.Error("parse failed", slog.Any("err", err))
			}
			log.Info("parsed")
			return f, err
		},
		PreLinkValidator: func(f *file.File) error {
			log := l.log.
				WithGroup("pre_link_validator").
				With(slog.String("mod", f.Module), slog.String("path_in_mod", f.PathInModule),
					slog.String("abs", f.AbsolutePath))

			log.Info("validating use namespaces")
			err := validate.PreLink(f)
			log.Info("validated use namespaces", slog.Any("err", err))
			return err
		},
		Linker: func(f *file.File) error {
			log := l.log.
				WithGroup("linker").
				With(slog.String("mod", f.Module), slog.String("path_in_mod", f.PathInModule),
					slog.String("abs", f.AbsolutePath))

			log.Info("inferring types of mixin params")
			typeinfer.Scope(f.Scope)
			log.Info("inferred types")
			log.Info("linking")
			err := l.linker.LinkFile(f)
			log.Info("linked file", slog.Any("err", err))
			return err
		},
		LibraryLinker: func(lib *file.Library) error {
			log := l.log.
				WithGroup("lib_linker").
				With(slog.String("mod", lib.Module), slog.String("path_in_mod", lib.PathInModule),
					slog.String("abs", lib.AbsolutePath))

			log.Info("inferring types of mixin params")
			for _, f := range lib.Files {
				typeinfer.Scope(f.Scope)
			}
			log.Info("inferred types")
			log.Info("linking")
			err := l.linker.LinkLibrary(lib)
			log.Info("linked file", slog.Any("err", err))
			return err
		},
		MainValidator: func(f *file.File) error {
			log := l.log.
				WithGroup("main_validator").
				With(slog.String("mod", f.Module), slog.String("path_in_mod", f.PathInModule),
					slog.String("abs", f.AbsolutePath))

			log.Info("validating file")
			err := validate.File(f)
			log.Info("validated file", slog.Any("err", err))
			return err
		},
		LibraryValidator: func(lib *file.Library) error {
			log := l.log.
				WithGroup("library_validator").
				With(slog.String("mod", lib.Module), slog.String("path_in_mod", lib.PathInModule),
					slog.String("abs", lib.AbsolutePath))

			log.Info("validating library")
			err := validate.Library(lib)
			log.Info("validated library", slog.Any("err", err))
			return err
		},
	}
	cl := load.Cache(bl)
	bl.DirLibraryLoader = func(f *file.File) (*file.Library, error) {
		if f.Module == "" {
			return nil, nil
		}

		return cl.LoadDirLibrary(f, func() (*file.Library, error) {
			lib, err := bl.LoadLibrary(f, path.Join(f.Module, path.Dir(f.PathInModule)))
			if err != nil {
				if errors.Is(err, load.ErrEmptyLib) {
					return nil, nil
				}

				return lib, err
			}

			return lib, nil
		})
	}

	l.loader = cl
	l.linker = link.New(l.loader)

	if o.GoExecPath == "" {
		if goroot := os.Getenv("GOROOT"); goroot != "" {
			o.GoExecPath = filepath.Join(goroot, "bin", "go")
		} else {
			return nil, errors.New("corgi.LoadOptions: GoExecPath not set and $GOROOT is empty")
		}
	}
	l.cmd = gocmd.NewCmd(o.GoExecPath)

	l.log = o.Logger
	if l.log == nil {
		l.log = nopLog
	}

	return &l, nil
}

func LoadMain(sysPath string, o LoadOptions) (*file.File, error) {
	l, err := newLoader(o)
	if err != nil {
		return nil, err
	}

	f, err := l.loader.LoadMain(sysPath)
	if err != nil {
		return f, err
	}

	if f == nil && err == nil {
		return nil, ErrNotExists
	}

	return f, nil
}

// LoadMainData parses and links the passed main file's raw data.
//
// Callers should set the Name of the returned file, so that error messages
// are correctly printed.
func LoadMainData(in []byte, o LoadOptions) (*file.File, error) {
	l, err := newLoader(o)
	if err != nil {
		return nil, err
	}

	f, err := parse.Parse(in)
	if f == nil && err == nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	f.Type = file.TypeMain

	if err := validate.PreLink(f); err != nil {
		return f, err
	}

	if err := l.linker.LinkFile(f); err != nil {
		return f, err
	}

	return f, validate.File(f)
}

// LoadLibrary parses and links the library located at the passed file system
// path.
func LoadLibrary(sysPath string, o LoadOptions) (*file.Library, error) {
	l, err := newLoader(o)
	if err != nil {
		return nil, err
	}

	lib, err := l.loader.LoadLibrary(nil, sysPath)
	if err != nil {
		return lib, err
	}

	if lib == nil && err == nil {
		return nil, ErrNotExists
	}

	return lib, nil
}

func (l *loader) readMain(sysPath string) (*load.File, error) {
	log := l.log.WithGroup("main_reader").With(slog.String("path", sysPath))

	p, err := l.resolvePaths(log, sysPath)
	if err != nil {
		return nil, err
	}

	log = log.With(slog.String("abs", p.sysAbs))

	data, err := l.readFile(log, sysPath)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}

	if err := l.readMainMod(log, filepath.Dir(sysPath)); err != nil {
		return nil, err
	}

	f := &load.File{
		Name:         p.base,
		AbsolutePath: p.sysAbs,
		IsCorgi:      true,
		Raw:          data,
	}
	if l.mainMod != nil {
		f.Module = l.mainMod.Module.Mod.Path
		f.PathInModule = pathInMod(l.mainModSysAbs, p.sysAbs)
	}
	return f, nil
}

func (l *loader) readTemplate(_ *file.File, extendPath string) (*load.File, error) {
	log := l.log.WithGroup("template_reader").With(slog.String("extend_path", extendPath))

	log.Info("reading file")
	log.Info("locating parent module")

	mod, err := l.locateModule(extendPath)
	if err != nil {
		return nil, err
	}
	if mod == nil {
		return nil, nil
	}

	sysAbs := filepath.Join(mod.sysAbsPath, filepath.FromSlash(mod.pathInMod))
	log = log.With(slog.String("module", mod.path), slog.String("path_in_mod", mod.pathInMod), slog.String("abs", mod.sysAbsPath))
	log.Info("located parent module", slog.String("module", mod.path))

	f, err := os.ReadFile(sysAbs)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Info("file doesn't exist in module")
			return nil, nil
		}

		log.Error("failed to load file", slog.Any("err", err))
		return nil, fmt.Errorf("%s: %w", extendPath, err)
	}

	log.Info("loaded file")

	return &load.File{
		Name:         filepath.Base(sysAbs),
		Module:       mod.path,
		PathInModule: mod.pathInMod,
		AbsolutePath: sysAbs,
		IsCorgi:      true,
		Raw:          f,
	}, nil
}

func (l *loader) readInclude(includingFile *file.File, slashPath string) (*load.File, error) {
	slashAbs := filepath.Join(path.Dir(includingFile.AbsolutePath), filepath.FromSlash(slashPath))
	sysAbs := filepath.FromSlash(slashAbs)

	log := l.log.
		WithGroup("include_reader").
		With(slog.String("include_path", slashPath), slog.String("abs", sysAbs),
			slog.String("including_file", includingFile.AbsolutePath))

	f, err := l.readFile(log, sysAbs)
	if err != nil {
		return nil, err
	}
	if f == nil {
		return nil, nil
	}

	return &load.File{
		Name:         path.Base(slashAbs),
		AbsolutePath: slashAbs,
		IsCorgi:      path.Ext(slashPath) == ".corgi",
		Raw:          f,
	}, nil
}

func (l *loader) readLibrary(usingFile *file.File, p string) (*load.Library, error) {
	if usingFile == nil {
		return l.readStandaloneLibrary(p)
	}

	return l.readDepLibrary(usingFile, p)
}

func (l *loader) readStandaloneLibrary(sysDir string) (*load.Library, error) {
	log := l.log.
		WithGroup("standalone_library_reader").
		With(slog.String("path", sysDir))

	dir, err := l.resolvePaths(log, sysDir)
	if err != nil {
		return nil, err
	}

	log = log.With(slog.String("abs", dir.sysAbs))

	if err := l.readMainMod(log, sysDir); err != nil {
		return nil, err
	}

	log.Info("reading standalone library")

	var modulePath, pathInModule string
	if l.mainMod != nil {
		modulePath = l.mainMod.Module.Mod.Path
		pathInModule, _ = filepath.Rel(l.mainModSysAbs, dir.sysAbs)
	}

	return l.readLibraryDir(log, sysDir, modulePath, pathInModule)
}

func (l *loader) readDepLibrary(_ *file.File, usePath string) (*load.Library, error) {
	log := l.log.
		WithGroup("dep_library_reader").
		With(slog.String("path", usePath))

	mod, err := l.locateModule(usePath)
	if err != nil {
		return nil, err
	}
	if mod == nil {
		if stdlib := std.Lib[usePath]; stdlib != nil {
			return &load.Library{Precompiled: stdlib}, nil
		}

		return nil, nil
	}

	sysAbs := filepath.Join(mod.sysAbsPath, filepath.FromSlash(mod.pathInMod))
	log = log.With(slog.String("abs", sysAbs))

	return l.readLibraryDir(log, sysAbs, mod.path, mod.pathInMod)
}

func (l *loader) readLibraryDir(log *slog.Logger, sysDir string, modulePath, pathInModule string) (*load.Library, error) {
	log.Info("loading dir information")

	dir, err := l.resolvePaths(log, sysDir)
	if err != nil {
		return nil, err
	}

	files, err := os.ReadDir(sysDir)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			log.Info("library does not exist")
			return nil, nil
		}

		log.Info("failed to load dir information", slog.Any("err", err))
		return nil, nil
	}

	log.Info("loaded dir information", slog.Int("num_files", len(files)))

	if len(files) == 0 {
		log.Info("dir is empty")
		return nil, nil
	}

	log.Info("checking if library is precompiled")

	if l.noPrecompile {
		log.Info("loader was configured to ignore precompiled library files, skipping check")
	} else {
		for _, entry := range files {
			name := entry.Name()

			log := log.With(slog.String("file", name))
			log.Debug("scanning file")

			if entry.Type() == os.ModeDir {
				log.Debug("skipping: file is dir")
				continue
			} else if name != PrecompFileName {
				log.Debug("skipping: name doesn't match: " + PrecompFileName)
				continue
			}

			log.Info("found file with matching name, reading")

			f, err := os.Open(filepath.Join(sysDir, name))
			if err != nil {
				log.Error("failed to open precompiled library file", slog.Any("err", err))
				return nil, fmt.Errorf("%s: failed to open precompiled library: %w", sysDir, err)
			}
			//goland:noinspection GoDeferInLoop
			defer f.Close()

			log.Info("decoding precompiled library file")

			lib, err := precomp.Decode(f)
			if err != nil {
				log.Error("failed to decode precompiled library file", slog.Any("err", err))
				return nil, fmt.Errorf("%s: failed to decode precompiled library: %w", sysDir, err)
			}

			log.Info("decoded precompiled library file, returning it")

			lib.AbsolutePath = dir.sysAbs
			lib.Module = modulePath
			lib.PathInModule = pathInModule
			return &load.Library{Precompiled: lib}, nil
		}

		log.Info("found no precompiled library file, compiling by hand")
	}

	lib := load.Library{
		AbsolutePath: dir.sysAbs,
		Files:        make([]load.File, 0, len(files)),
	}
	if l.mainMod != nil {
		lib.Module = l.mainMod.Module.Mod.Path
		lib.PathInModule, _ = filepath.Rel(l.mainModSysAbs, dir.sysAbs)
	}

	log.Info("looking for corgi lib files")

	for _, entry := range files {
		name := entry.Name()

		log := log.With(slog.String("file", name))
		log.Debug("scanning file")

		if entry.Type() == os.ModeDir {
			log.Debug("skipping: file is dir")
			continue
		} else if !strings.HasSuffix(name, LibExt) {
			log.Debug("skipping: extension doesn't match: " + LibExt)
			continue
		}

		log.Info("found corgi lib file, reading")

		p := filepath.Join(sysDir, name)
		readFile, err := os.ReadFile(p)
		if err != nil {
			log.Error("failed to open corgi lib file", slog.Any("err", err))
			return nil, fmt.Errorf("%s: failed to read library file: %w", p, err)
		}

		log.Info("read corgi lib file successfully")

		f := load.File{
			Name:         name,
			Module:       modulePath,
			AbsolutePath: filepath.Join(dir.sysAbs, name),
			IsCorgi:      true,
			Raw:          readFile,
		}
		if modulePath != "" {
			f.PathInModule = filepath.Join(f.PathInModule, name)
		}

		lib.Files = append(lib.Files, f)
	}

	log.Info("read directory, returning with library",
		slog.Int("size", len(lib.Files)), slog.Int("skipped", len(files)-len(lib.Files)))

	return &lib, nil
}

// ============================================================================
// Utils
// ======================================================================================

func (l *loader) readFile(log *slog.Logger, sysPath string) ([]byte, error) {
	log.Info("reading file")

	f, err := os.ReadFile(sysPath)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			log.Error("file not found")
			return nil, nil
		}

		log.Error("failed to read file", slog.Any("err", err))
		return nil, fmt.Errorf("%s: %w", sysPath, err)
	}

	log.Info("file read")
	return f, nil
}

type paths struct {
	base string

	sysRel   string
	sysAbs   string
	slashRel string
	slashAbs string
}

func (l *loader) resolvePaths(log *slog.Logger, sysRel string) (paths, error) {
	p := paths{sysRel: sysRel, slashRel: filepath.ToSlash(sysRel)}

	var err error
	p.sysAbs, err = filepath.Abs(filepath.FromSlash(p.sysRel))
	if err != nil {
		log.Error("failed to resolve absolute path", slog.Any("err", err))
		return paths{}, fmt.Errorf("%s: %w", sysRel, err)
	}
	p.slashAbs = filepath.ToSlash(p.sysAbs)
	p.base = filepath.Base(p.sysAbs)

	return p, nil
}

type mod struct {
	path       string
	pathInMod  string
	sysAbsPath string
}

func (l *loader) locateModule(of string) (*mod, error) {
	log := l.log.WithGroup("locate_module").With(slog.String("of", of))
	log.Info("locating module")

	if l.mainMod == nil {
		log.Info("main file has no go.mod, downloading latest instead of using tagged version")
		return l.downloadModule(log, of, "latest")
	}

	if strings.HasPrefix(of, l.mainMod.Module.Mod.Path) {
		log.Info("file is in main module, using workdir instead of module cache", slog.String("module", l.mainMod.Module.Mod.Path))
		return &mod{
			path:       l.mainMod.Module.Mod.Path,
			pathInMod:  pathInMod(l.mainMod.Module.Mod.Path, of),
			sysAbsPath: l.mainModSysAbs,
		}, nil
	}

	log.Info("looking for module in main file's go.mod")

	var dep module.Version

	log.Info("looking for replace directives")

	for _, replace := range l.mainMod.Replace {
		log := log.With(slog.String("old", replace.Old.String()), slog.String("new", replace.New.String()))
		log.Debug("scanning replace directive")
		if strings.HasPrefix(of, replace.Old.Path) {
			if path.IsAbs(replace.New.Path) {
				log.Info("found and respecting replace directive")
				return &mod{
					path:       replace.Old.Path,
					pathInMod:  pathInMod(replace.Old.Path, of),
					sysAbsPath: replace.New.Path,
				}, nil
			}

			dep = replace.New
			goto foundModule
		}
	}

	for _, require := range l.mainMod.Require {
		log := log.With(slog.String("require", require.Mod.String()))
		log.Debug("scanning require directive")

		if strings.HasPrefix(of, require.Mod.Path) {
			log.Info("found matching require directive")
			dep = require.Mod
			break
		}
	}

	if dep.Path == "" {
		log.Info("module not in main file's go.mod (dirty go.mod? (go mod tidy?)), downloading latest using `go mod download`")
		return l.downloadModule(log, of, "latest")
	}

	log = log.With(slog.String("module", dep.Path))

	log.Info("locating in module cache")

foundModule:
	sysModCache := l.cmd.EnvGOMODCACHE()
	if sysModCache == "" {
		log.Error("unable to locate go mod cache (`go env GOMODCACHE` == \"\")")
		return nil, errors.New("unable to locate go mod cache")
	}

	log.Info("looking up module in go module cache")

	sysModuleAbs := filepath.Join(sysModCache, filepath.FromSlash(dep.Path)) + "@" + dep.Version
	f, err := os.Open(sysModuleAbs)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			log.Info("module version not cached, downloading")
			return l.downloadModule(log, dep.Path, dep.Version)
		}

		log.Error("failed to check if module is cached", slog.Any("err", err))

		return nil, fmt.Errorf("failed to check if module is cached: %w", err)
	}

	log.Info("module version cached, returning path to cached module")

	_ = f.Close()
	return &mod{
		path:       dep.Path,
		pathInMod:  pathInMod(dep.Path, of),
		sysAbsPath: sysModuleAbs,
	}, nil
}

func (l *loader) downloadModule(log *slog.Logger, of, version string) (*mod, error) {
	log.Info("downloading module", slog.String("of", of), slog.String("version", version))
	return l._downloadModule(log, of, "", version)
}

func (l *loader) _downloadModule(log *slog.Logger, modulePath, pathInModule, version string) (*mod, error) {
	query := modulePath + "@" + version

	log.Debug("running `go mod download " + query + "`")

	m, err := l.cmd.DownloadMod(query)
	if err != nil {
		if strings.Contains(err.Error(), "malformed module path") || strings.Contains(err.Error(), "unrecognized import path") {
			log.Info("module does not exist")
			return nil, nil
		} else if strings.Contains(err.Error(), "no matching versions") {
			pathInModule = path.Join(path.Base(modulePath), pathInModule)
			modulePath = path.Dir(modulePath)
			if modulePath == "." {
				log.Info("module does not exist")
				return nil, nil
			}

			log.Debug("possibly a module path with appended dir, retrying with parent dir")
			return l._downloadModule(log, modulePath, pathInModule, version)
		}

		log.Error("failed to download module", slog.Any("err", err))
		return nil, err
	}

	log.Info("downloaded module", slog.String("module_path", m.Path), slog.String("abs", m.Dir))

	return &mod{
		path:       m.Path,
		pathInMod:  pathInModule,
		sysAbsPath: m.Dir,
	}, nil
}

func pathInMod(modulePath string, fullPath string) string {
	if len(fullPath) <= len(modulePath) {
		return ""
	}

	fullPath = fullPath[len(modulePath):]
	for len(fullPath) > 0 && fullPath[0] == '/' {
		fullPath = fullPath[1:]
	}

	return fullPath
}

func (l *loader) readMainMod(log *slog.Logger, sysDir string) error {
	log.Info("reading main module")

	mod, absPath, err := gomod.Find(sysDir)
	if err != nil {
		log.Error("failed to read main module", slog.Any("err", err))
		return err
	}
	if mod == nil {
		log.Info("file not in module")
	} else {
		log.Info("read main module", slog.String("path", filepath.FromSlash(mod.Module.Mod.Path)))
		l.mainMod = mod
		l.mainModSysAbs = filepath.Dir(absPath)
	}

	return nil
}
