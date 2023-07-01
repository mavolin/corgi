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
	"sync"

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
	"github.com/mavolin/corgi/validate"
)

const (
	PrecompFileName = "lib.precorgi"
	Ext             = ".corgi"
	LibExt          = ".corgil"
)

var (
	ErrExists      = errors.New("file does not exist")
	ErrGoNotInPath = errors.New("the go command is not in the path (set the -go flag/call corgi.SetGoExecPath)")
)

// loader is responsible for loading, i.e. reading, parsing, validating, and
// linking corgi files, like the CLI does.
//
// It keeps state about the module data of loaded files, more concretely
// results of calling `go mod edit -json` used to resolve import paths.
//
// It is concurrently safe.
type loader struct {
	mut sync.Mutex
	// mods contains a mapping between directories and their associated modules
	mods map[string] /* abs of dir */ *goModule

	mainMod *goModule

	loader load.Loader
	linker *link.Linker
	cmd    *gocmd.Cmd
	log    *slog.Logger
}

type goModule struct {
	mod       *modfile.File
	err       error
	slashPath string // forward slash path to mod
	done      <-chan struct{}
}

type LoadOptions struct {
	// GoExecPath is the path to the go binary to used.
	//
	// If not set, the Go executable referenced in the system's PATH will be
	// used, as resolved once at program start.
	GoExecPath string

	// Logger is used to log the individual steps of the logging process.
	//
	// If left as nil, nothing will be logged
	Logger *slog.Logger
}

var nopLog = slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelError + 1}))

func newLoader(o LoadOptions) (*loader, error) {
	var l loader

	bl := load.BasicLoader{
		MainReader:     l.readMain,
		TemplateReader: l.readTemplate,
		IncludeReader:  l.readInclude,
		LibraryReader:  l.readLibrary,
		Parser: func(in []byte) (*file.File, error) {
			log := l.log.WithGroup("parser")

			log.Info("parsing")
			f, err := parse.Parse(in)
			log.Error("parsed filed", slog.Any("err", err))
			return f, err
		},
		PreLinkValidator: func(f *file.File) error {
			log := l.log.
				WithGroup("pre_link_validator").
				With(slog.String("mod", f.Module), slog.String("path_in_mod", f.ModulePath),
					slog.String("abs", filepath.FromSlash(f.AbsolutePath)))

			log.Info("validating use namespaces")
			err := validate.PreLink(f)
			log.Info("validated use namespaces", slog.Any("err", err))
			return err
		},
		Linker: func(f *file.File) error {
			log := l.log.
				WithGroup("linker").
				With(slog.String("mod", f.Module), slog.String("path_in_mod", f.ModulePath),
					slog.String("abs", filepath.FromSlash(f.AbsolutePath)))

			log.Info("inferring types of mixin params")
			typeinfer.Scope(f.Scope)
			log.Info("inferred types")
			log.Info("linking")
			err := l.linker.LinkFile(f)
			log.Info("linked file", slog.Any("err", err))
			return err
		},
		MainValidator: func(f *file.File) error {
			log := l.log.
				WithGroup("main_validator").
				With(slog.String("mod", f.Module), slog.String("path_in_mod", f.ModulePath),
					slog.String("abs", filepath.FromSlash(f.AbsolutePath)))

			log.Info("validating file")
			err := validate.File(f)
			log.Info("validated file", slog.Any("err", err))
			return err
		},
		LibraryValidator: func(lib *file.Library) error {
			log := l.log.
				WithGroup("library_validator").
				With(slog.String("mod", lib.Module), slog.String("path_in_mod", lib.ModulePath),
					slog.String("abs", filepath.FromSlash(lib.AbsolutePath)))

			log.Info("validating library")
			err := validate.Library(lib)
			log.Info("validated library", slog.Any("err", err))
			return err
		},
	}
	cl := load.Cache(bl)
	bl.DirLibraryLoader = func(f *file.File) (*file.Library, error) {
		return cl.LoadDirLibrary(f, func() (*file.Library, error) {
			return bl.LoadLibrary(f, path.Join(f.Module, path.Base(f.ModulePath)))
		})
	}

	l.loader = cl

	l.linker = link.New(l.loader)

	l.cmd = gocmd.NewCmd(o.GoExecPath)
	if l.cmd == nil {
		return nil, ErrGoNotInPath
	}

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
	if f == nil && err == nil {
		return nil, ErrExists
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
		return nil, err
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
	if lib == nil && err == nil {
		return nil, ErrExists
	}

	return lib, nil
}

func (l *loader) readMain(sysPath string) (*load.File, error) {
	log := l.log.WithGroup("main_reader").With(slog.String("path", sysPath))

	log.Info("reading file")

	sysAbs, err := filepath.Abs(filepath.FromSlash(sysPath))
	slashAbs := filepath.ToSlash(sysAbs)
	if err != nil {
		log.Error("failed to resolve absolute path", slog.Any("err", err))
		return nil, fmt.Errorf("%s: %w", sysPath, err)
	}

	log = log.With(slog.String("abs", sysAbs))

	f, err := os.ReadFile(sysAbs)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			log.Error("file not found")
			return nil, nil
		}

		log.Error("failed to read file", slog.Any("err", err))
		return nil, fmt.Errorf("%s: %w", sysPath, err)
	}

	log.Info("file read", slog.String("abs", sysAbs))

	return &load.File{
		Name:         filepath.Base(sysAbs),
		AbsolutePath: slashAbs,
		IsCorgi:      true,
		Raw:          f,
	}, nil
}

func (l *loader) readTemplate(_ *file.File, slashPath string) (*load.File, error) {
	log := l.log.WithGroup("template_reader").With(slog.String("path", slashPath))

	log.Info("reading file")
	log.Info("locating parent module")

	mod, _ := l.goMod(path.Dir(slashPath))

	if mod == nil {
		log = log.With(slog.String("mod", mod.mod.Module.Mod.Path))
		log.Info("located parent module")
	} else {
		log.Info("file is not inside a module")
	}

	if mod != nil && l.mainMod != nil && mod.slashPath == l.mainMod.slashPath {
		slashPathInMod := slashPath[len(mod.mod.Module.Mod.Path)+1:]
		slashAbs := path.Join(mod.slashPath, slashPathInMod)
		sysAbs := filepath.FromSlash(slashAbs)

		log = log.With(slog.String("path_in_mod", slashPathInMod), slog.String("abs", sysAbs))

		log.Info("file is in same module as main, reading directly")

		f, err := os.ReadFile(filepath.FromSlash(slashAbs))
		if err != nil {
			if errors.Is(err, os.ErrExist) {
				log.Error("file does not exist")
				return nil, nil
			}

			log.Error("failed to load file", slog.Any("err", err))
			return nil, fmt.Errorf("%s: %w", slashPath, err)
		}

		log.Info("file read from main module")
		return &load.File{
			Name:         path.Base(slashAbs),
			Module:       mod.mod.Module.Mod.Path,
			ModulePath:   slashPathInMod,
			AbsolutePath: slashAbs,
			IsCorgi:      true,
			Raw:          f,
		}, nil
	}

	log.Info("file is in different module, locating cache or downloading")
	slashModAbs, slashModPath, err := l.locateModule(mod, slashPath)
	if err != nil {
		log.Error("failed to locate or download module", slog.Any("err", err))
		return nil, err
	}

	slashPathInMod := slashPath[len(slashModPath)+1:]
	slashAbs := path.Join(slashModAbs, slashPathInMod)
	sysAbs := filepath.FromSlash(slashAbs)

	log = log.With(slog.String("path_in_mod", slashPathInMod), slog.String("abs", sysAbs))

	log.Info("located module")

	f, err := os.ReadFile(sysAbs)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			log.Info("file doesn't exist in module")
			return nil, nil
		}

		log.Error("failed to load file", slog.Any("err", err))
		return nil, fmt.Errorf("%s: %w", slashPath, err)
	}

	log.Info("loaded file")

	return &load.File{
		Name:         filepath.Base(sysAbs),
		Module:       slashModPath,
		ModulePath:   slashPathInMod,
		AbsolutePath: slashAbs,
		IsCorgi:      true,
		Raw:          f,
	}, nil
}

func (l *loader) readInclude(includingFile *file.File, slashPath string) (*load.File, error) {
	slashAbs := path.Join(path.Dir(includingFile.AbsolutePath), slashPath)
	sysAbs := filepath.FromSlash(slashAbs)

	log := l.log.
		WithGroup("include_reader").
		With(slog.String("path", slashPath), slog.String("abs", sysAbs),
			slog.String("including_file", filepath.FromSlash(includingFile.AbsolutePath)))

	log.Info("reading file")

	f, err := os.ReadFile(sysAbs)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			log.Info("file not found")
			return nil, nil
		}

		log.Info("failed to read file", slog.Any("err", err))
		return nil, fmt.Errorf("%s: %w", slashPath, err)
	}

	log.Info("loaded file")
	return &load.File{
		Name:         path.Base(slashAbs),
		AbsolutePath: slashAbs,
		IsCorgi:      true,
		Raw:          f,
	}, nil
}

func (l *loader) readLibrary(usingFile *file.File, slashPath string) (*load.Library, error) {
	log := l.log.
		WithGroup("library_reader").
		With(slog.String("path", slashPath))

	if usingFile == nil {
		log.Info("reading standalone library")
		return l.loadLibrary(slashPath, "", "")
	}

	log = log.With(slog.String("using_file", filepath.ToSlash(usingFile.AbsolutePath)))
	log.Info("reading dependency library")

	log.Info("locating parent module")
	mod, _ := l.goMod(slashPath)

	if mod == nil {
		log.Info("located parent module", slog.String("mod", mod.mod.Module.Mod.Path))
	} else {
		log.Info("file is outside of module")
	}

	if mod != nil && l.mainMod != nil && mod.slashPath == l.mainMod.slashPath {
		log.Info("lib is in same module as main, loading directly")

		slashPathInMod := slashPath[len(mod.mod.Module.Mod.Path)+1:]
		slashAbs := path.Join(mod.slashPath, slashPathInMod)

		return l.loadLibrary(slashAbs, mod.mod.Module.Mod.Path, slashPathInMod)
	}

	log.Info("lib is in different module, locating module")

	modAbs, modPath, err := l.locateModule(mod, slashPath)
	if err != nil {
		log.Error("failed to locate or download module", slog.Any("err", err))
		return nil, err
	}

	slashPathInMod := slashPath[len(modPath)+1:]
	slashAbs := path.Join(modAbs, slashPathInMod)
	sysAbs := filepath.FromSlash(slashAbs)

	log = log.With(slog.String("path_in_mod", slashPathInMod), slog.String("abs", sysAbs))

	log.Info("located module")
	return l.loadLibrary(sysAbs, modPath, slashPath[len(modAbs)+1:])
}

func (l *loader) loadLibrary(slashAbs, slashModPath, slashPathInMod string) (*load.Library, error) {
	sysAbs := filepath.FromSlash(slashAbs)

	log := l.log.
		WithGroup("library_reader").
		With(slog.String("abs", sysAbs),
			slog.String("mod", slashModPath), slog.String("path_in_mod", slashPathInMod))

	log.Info("loading dir information")

	files, err := os.ReadDir(sysAbs)
	if err != nil {
		log.Info("failed to load dir information", slog.Any("err", err))
		return nil, nil
	}

	log.Info("loaded dir information")

	absSlash := filepath.ToSlash(slashAbs)

	for _, entry := range files {
		name := entry.Name()

		log := log.With(slog.String("file", name))
		log.Info("looking for precompiled library file")

		if entry.Type() != os.ModeDir {
			log.Info("skipping: file is dir")
		} else if name != PrecompFileName {
			log.Info("skipping: name doesn't match: " + PrecompFileName)
			continue
		}

		log.Info("found file with matching name, reading")

		f, err := os.Open(filepath.Join(slashAbs, name))
		if err != nil {
			log.Error("failed to open precompiled library file", slog.Any("err", err))
			return nil, fmt.Errorf("%s: failed to open precompiled library: %w", sysAbs, err)
		}
		//goland:noinspection GoDeferInLoop
		defer f.Close()

		log.Info("decoding precompiled library file")

		lib, err := precomp.Decode(f)
		if err != nil {
			log.Error("failed to decode precompiled library file", slog.Any("err", err))
			return nil, fmt.Errorf("%s: failed to decode precompiled library: %w", sysAbs, err)
		}

		log.Info("decoded precompiled library file")

		lib.AbsolutePath = absSlash
		lib.Module = slashModPath
		lib.ModulePath = slashPathInMod
		return &load.Library{Precompiled: lib}, nil
	}

	log.Info("found no precompiled library file, compiling by hand")

	lib := load.Library{
		Module:       slashModPath,
		ModulePath:   slashPathInMod,
		AbsolutePath: filepath.ToSlash(slashAbs),
		Files:        make([]load.File, 0, len(files)),
	}

	for _, entry := range files {
		name := entry.Name()

		log := log.With(slog.String("file", name))
		log.Info("looking for corgi lib files")

		if entry.Type() != os.ModeDir {
			log.Info("skipping: file is dir")
		} else if !strings.HasSuffix(name, LibExt) {
			log.Info("skipping: extension doesn't match: " + LibExt)
			continue
		}

		log.Info("found corgi lib file, reading")

		readFile, err := os.ReadFile(filepath.Join(sysAbs, name))
		if err != nil {
			log.Error("failed to open corgi lib file", slog.Any("err", err))
			return nil, fmt.Errorf("%s: failed to read library file: %w", filepath.Join(sysAbs, name), err)
		}

		log.Info("read corgi lib file successfully")

		lib.Files = append(lib.Files, load.File{
			Name:         name,
			Module:       slashModPath,
			ModulePath:   slashPathInMod,
			AbsolutePath: path.Join(absSlash, name),
			IsCorgi:      true,
			Raw:          readFile,
		})
	}

	log.Info("read directory, returning with library",
		slog.Int("size", len(lib.Files)), slog.Int("skipped", len(files)-len(lib.Files)))

	return &lib, nil
}

func (l *loader) locateModule(mod *goModule, slashPath string) (slashAbs string, slashModPath string, err error) {
	log := l.log.WithGroup("locate_module").With(slog.String("path", slashPath))

	// we can't resolve cache path w/o knowing the module, the only option left
	// is to directly download
	if mod == nil {
		log.Info("no module information for file/dir, downloading latest directly using `go mod download`")
		return l.downloadModule(slashPath + "@latest")
	}

	log = log.With(slog.String("mod", mod.mod.Module.Mod.Path))

	// module is not mentioned in the main module, download it directly
	if l.mainMod == nil {
		log.Info("main file has no go.mod, downloading latest instead of using tagged version")
		return l.downloadModule(slashPath + "@latest")
	}

	log.Info("looking for module in main file's go.mod")

	var dep module.Version

	for _, replace := range l.mainMod.mod.Replace {
		if strings.HasPrefix(slashPath, replace.Old.Path) {
			if path.IsAbs(replace.New.Path) {
				log.Info("found and respecting replace directive", slog.String("replace_with", replace.New.Path))
				return replace.New.Path, replace.Old.Path, nil
			}

			dep = replace.New
			goto foundModule
		}
	}

	for _, require := range mod.mod.Require {
		if strings.HasPrefix(slashPath, require.Mod.Path) {
			log = log.With(slog.String("mod_ver", require.Mod.Version))
			log.Info("found require directive")
			dep = require.Mod
			break
		}
	}

	if dep.Path == "" {
		log.Info("module not main file's go.mod (go mod tidy?), downloading latest using `go mod download`")
		return l.downloadModule(slashPath + "@latest")
	}

foundModule:
	sysModCache := l.cmd.Env_GOMODCACHE()
	if sysModCache == "" {
		log.Error("unable to locate go mod cache (`go env GOMODCACHE` == \"\")")
		return "", "", errors.New("unable to locate go mod cache")
	}

	log.Info("looking up module in go module cache")

	sysModulePath := filepath.Join(sysModCache, filepath.FromSlash(dep.Path)) + "@" + dep.Version
	f, err := os.Open(sysModulePath)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			log.Info("module version not cached, downloading")
			return l.downloadModule(dep.Path + "@" + dep.Version)
		}

		log.Info("failed to check if module is cached")

		return "", "", fmt.Errorf("failed to check if module is cached: %w", err)
	}

	log.Info("module version cached, returning cached file")

	_ = f.Close()
	return filepath.ToSlash(sysModulePath), dep.Path, nil
}

func (l *loader) downloadModule(slashPath string) (slashAbs string, slashModPath string, err error) {
	mod, err := l.cmd.DownloadMod(slashPath)
	if err != nil {
		return "", "", err
	}
	if mod.Error != "" {
		return "", "", errors.New(mod.Error)
	}

	return filepath.ToSlash(mod.GoMod), mod.Path, nil
}

func (l *loader) goMod(p string) (*goModule, error) {
	l.mut.Lock()

	mod := l.mods[p]
	if mod != nil {
		l.mut.Unlock()
		<-mod.done
		return mod, mod.err
	}

	done := make(chan struct{})
	mod = &goModule{done: done}
	l.mods[p] = mod
	l.mut.Unlock()

	// gomod.Find takes about 220µs, so it's still worth to cache
	mod.mod, mod.slashPath, mod.err = gomod.Find(p)
	mod.slashPath = filepath.ToSlash(mod.slashPath)
	close(done)

	return mod, mod.err
}
