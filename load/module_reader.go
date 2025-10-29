package load

import (
	"cmp"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/internal/assert"
	"github.com/mavolin/corgi/v2/internal/gocmd"
	"github.com/mavolin/corgi/v2/internal/meta"
)

// Stdlib is the path to the corgi standard library.
const Stdlib file.GoImportPath = meta.Module + "/std"

type (
	// ModuleReader is a [Reader] that works in a specific Go module and downloads
	// dependencies using a Go executable.
	ModuleReader struct {
		logger *slog.Logger
		// sorted by GoPath descending, so that longer prefixes are matched
		// first
		symbolicImportsByGoPath    []SymbolicImportMapping
		symbolicImportsByCorgiPath []SymbolicImportMapping

		mainModule gocmd.ListModuleResult
		// sorted by Path descending, so that longer prefixes are matched first
		modules []gocmd.ListModuleResult
	}
	SymbolicImportMapping struct {
		CorgiPath file.CorgiImportPath
		GoPath    file.GoImportPath
	}
)

var _ Reader = (*ModuleReader)(nil)

type ModuleReaderOptions struct {
	// GoExecPath is a path to a go binary.
	//
	// Default: `go` from $PATH
	GoExecPath string

	// StdLibPath is the path to the corgi stdlib.
	//
	// Default: github.com/mavolin/corgi/v2/std
	StdLibPath file.GoImportPath

	// SymbolicImports is a mapping of symbolic paths to their real import
	// paths.
	// The stdlib import is always mapped to "corgi".
	// ModuleReader takes ownership of the slice and the caller must not modify
	// it after calling NewModuleReader.
	//
	// Default: nil
	SymbolicImports []SymbolicImportMapping

	// Logger produces logs for reading operations.
	//
	// Default: no logging
	Logger *slog.Logger
}

func (o *ModuleReaderOptions) applyDefaults() error {
	if o.Logger == nil {
		o.Logger = slog.New(slog.DiscardHandler)
	}

	var err error
	if o.GoExecPath == "" {
		o.GoExecPath, err = exec.LookPath("go")
		if err != nil {
			return fmt.Errorf("ModuleReader: GoExecPath: looking up `go` in $PATH: %w", err)
		}
		o.Logger.Info("No GoExecPath set, using `go` from $PATH", slog.String("resolved_path", o.GoExecPath))
	}

	if o.StdLibPath == "" {
		o.StdLibPath = Stdlib
	}

	if o.SymbolicImports == nil {
		o.SymbolicImports = make([]SymbolicImportMapping, 0, 1)
	}
	for _, m := range o.SymbolicImports {
		if err = m.CorgiPath.CheckValid(); err != nil {
			return fmt.Errorf("ModuleReader: SymbolicImports: invalid corgi import path %q: %w", m.CorgiPath, err)
		} else if err = m.GoPath.CheckValid(); err != nil {
			return fmt.Errorf("ModuleReader: SymbolicImports: invalid go import path %q: %w", m.GoPath, err)
		}
	}
	o.SymbolicImports = append(o.SymbolicImports, SymbolicImportMapping{
		CorgiPath: "corgi",
		GoPath:    o.StdLibPath,
	})

	return nil
}

// NewModuleReader creates a new [Reader] that loads packages and dependencies
// of dir, where dir is part of a Go module.
//
// The modules dependencies are loaded once and must not change during the
// lifetime of the returned ModuleReader.
func NewModuleReader(ctx context.Context, dir filesystemPath, o ModuleReaderOptions) (*ModuleReader, error) {
	if err := o.applyDefaults(); err != nil {
		return nil, err
	}

	r := ModuleReader{
		logger:                     o.Logger,
		symbolicImportsByGoPath:    o.SymbolicImports,
		symbolicImportsByCorgiPath: make([]SymbolicImportMapping, len(o.SymbolicImports)),
	}
	copy(r.symbolicImportsByCorgiPath, o.SymbolicImports)
	slices.SortFunc(r.symbolicImportsByGoPath, func(a, b SymbolicImportMapping) int {
		return -cmp.Compare(a.GoPath, b.GoPath)
	})
	slices.SortFunc(r.symbolicImportsByCorgiPath, func(a, b SymbolicImportMapping) int {
		return -cmp.Compare(a.CorgiPath, b.CorgiPath)
	})

	cmd := gocmd.New(o.GoExecPath)
	result, err := cmd.ListAllModules(ctx, dir)
	if err != nil {
		return nil, fmt.Errorf("NewModuleReader: %w", err)
	}

	r.mainModule = result.MainModule
	r.modules = append(result.Dependencies, result.MainModule) //nolint:gocritic
	slices.SortFunc(r.modules, func(a, b gocmd.ListModuleResult) int {
		return -cmp.Compare(a.Path, b.Path)
	})

	return &r, nil
}

// LocalImportPath returns the import path of the passed directory, which must
// be part of the module represented by r.
func (r *ModuleReader) LocalImportPath(dir filesystemPath) (file.CorgiImportPath, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("absolute path of %q: %w", dir, err)
	}

	if !strings.HasPrefix(abs, r.mainModule.Dir) {
		return "", fmt.Errorf("%q is not part of module %q (located in %q)",
			dir, r.mainModule.Path, r.mainModule.Dir)
	}

	rel, err := filepath.Rel(r.mainModule.Dir, abs)
	if err != nil {
		return "", fmt.Errorf("path in module: %w", err)
	}

	imp := file.GoImportPath(path.Join(r.mainModule.Path, filepath.ToSlash(rel)))
	assert.NoError(imp.CheckValid(), "constructed invalid import path: "+string(imp))

	symbolicPath := r.symbolicImport(imp)
	if symbolicPath != "" {
		return symbolicPath, nil
	}

	return file.CorgiImportPath(imp), nil
}

func (r *ModuleReader) ReadImport(_ context.Context, p file.CorgiImportPath) (*Package, error) {
	if err := p.CheckValid(); err != nil {
		return nil, fmt.Errorf("invalid import path %q: %w", p, err)
	}

	resolved := r.resolveGoPath(p)
	if resolved != file.GoImportPath(p) {
		r.logger.Debug("Adjusted symbolic import",
			slog.String("old", string(p)),
			slog.String("adjusted", string(resolved)))
	}

	mod := r.findModule(resolved)
	if mod == nil {
		return nil, fmt.Errorf("no required module provides package %s", p)
	}

	packagePath := file.PackagePath(strings.TrimPrefix(string(resolved[len(mod.Path):]), "/"))
	assert.NoError(file.GoImportPath(packagePath).CheckValid(), "constructed invalid package path: "+string(packagePath))
	return r.readImport(*mod, packagePath)
}

func (r *ModuleReader) readImport(mod gocmd.ListModuleResult, pkgPath file.PackagePath) (*Package, error) {
	fullPath := path.Join(mod.Dir, filepath.FromSlash(string(pkgPath)))

	logger := r.logger.With(
		slog.String("module", mod.Path),
		slog.String("module_dir", mod.Dir),
		slog.String("package_path", string(pkgPath)))
	logger.Debug("Reading package")

	dir, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, fmt.Errorf("reading directory %q: %w", fullPath, err)
	}

	pkg := &Package{
		Module:       file.ModulePath(mod.Path),
		PathInModule: pkgPath,
		Files:        make([]File, 0, len(dir)),
	}

	for _, entry := range dir {
		switch {
		case strings.HasPrefix(entry.Name(), ".") || strings.HasPrefix(entry.Name(), "_"):
			continue
		case !strings.HasSuffix(entry.Name(), Ext):
			continue
		case entry.Type() != 0:
			if !entry.IsDir() {
				r.logger.Debug("Skipping non-regular file",
					slog.String("file", entry.Name()),
					slog.String("file_type", entry.Type().String()))
			}
			continue
		}

		filePath := path.Join(fullPath, entry.Name())
		raw, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("reading file %q: %w", filePath, err)
		}
		pkg.Files = append(pkg.Files, File{
			Name: file.Name(entry.Name()),
			Raw:  string(raw),
		})
	}

	return pkg, nil
}

func (r *ModuleReader) resolveGoPath(p file.CorgiImportPath) file.GoImportPath {
	for _, mapping := range r.symbolicImportsByCorgiPath {
		if !strings.HasPrefix(string(p), string(mapping.CorgiPath)) {
			continue
		}
		gp := mapping.GoPath + file.GoImportPath(p[len(mapping.CorgiPath):])
		assert.NoError(gp.CheckValid(), "constructed invalid go import path: "+string(gp))
		return gp
	}
	return file.GoImportPath(p)
}

func (r *ModuleReader) findModule(p file.GoImportPath) *gocmd.ListModuleResult {
	for _, mod := range r.modules {
		if strings.HasPrefix(string(p), mod.Path) {
			return &mod
		}
	}
	return nil
}

func (r *ModuleReader) symbolicImport(imp file.GoImportPath) file.CorgiImportPath {
	for _, mapping := range r.symbolicImportsByGoPath {
		if strings.HasPrefix(string(imp), string(mapping.GoPath)) {
			ip := mapping.CorgiPath + file.CorgiImportPath(imp[len(mapping.GoPath):])
			assert.NoError(ip.CheckValid(), "constructed invalid corgi import path: "+string(ip))
			return ip
		}
	}
	return ""
}
