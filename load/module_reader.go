package load

import (
	"context"
	"fmt"
	"log/slog"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/internal/cache"
	"github.com/mavolin/corgi/v2/internal/gocmd"
	"github.com/mavolin/corgi/v2/internal/gomod"
	"golang.org/x/mod/modfile"
)

// ModuleReader is a [Reader] that works in a specific Go module and downloads
// dependencies using a Go executable.
type ModuleReader struct {
	goCmd      *gocmd.Cmd
	logger     *slog.Logger
	stdLibPath string

	modFile      *modfile.File
	modFileAbs   filesystemPath
	modDownloads *cache.Value[[]gocmd.DownloadModule]
}

var _ Reader = (*ModuleReader)(nil)

type ModuleReaderOptions struct {
	// GoExecPath is a path to a go binary.
	//
	// Default: $GOROOT/bin/go
	GoExecPath string

	// StdLibPath is the path to the corgi stdlib.
	//
	// Default: github.com/mavolin/corgi/v2/std
	StdLibPath string

	// Logger is used to log the individual steps of the logging process.
	//
	// Default: no logging
	Logger *slog.Logger
}

// NewModuleReader creates a new ModuleReader with the passed options.
// ModuleDir is the directory or a child directory of the module, whose go.mod
// file should be used.
func NewModuleReader(moduleDir string, o ModuleReaderOptions) (*ModuleReader, error) {
	var r ModuleReader

	r.logger = o.Logger
	if o.Logger == nil {
		r.logger = slog.New(slog.DiscardHandler)
	}

	var err error
	if o.GoExecPath == "" {
		o.GoExecPath, err = exec.LookPath("go")
		if err == /* IS */ nil {
			o.Logger.Info("No GoExecPath set, using `go` from $PATH", slog.String("resolved_path", o.GoExecPath))
		} else {
			o.Logger.Warn("GoExecPath not set and no `go` in $PATH, running in local mode; won't be able to download external dependencies.")
		}
	}

	if o.StdLibPath == "" {
		o.StdLibPath = "github.com/mavolin/corgi/v2/std"
	}
	r.stdLibPath = o.StdLibPath

	r.modFile, r.modFileAbs, err = gomod.Find(moduleDir)
	if err != nil {
		o.Logger.Error("Failed to locate go.mod", slog.String("err", err.Error()))
		return nil, fmt.Errorf("locating go.mod: %w", err)
	}
	o.Logger.Info("Found go.mod",
		slog.String("module", r.modFile.Module.Mod.Path),
		slog.String("path", r.modFileAbs))

	if o.GoExecPath != "" {
		// r.goCmd = gocmd.New(o.GoExecPath)
		// r.modDownloads = cache.Preload(r.goCmd.DownloadModules)
	}

	return &r, nil
}

// LocalImportPath returns the import path of the passed directory, which must
// be part of the module.
//
// dir must use the filesystem's separator.
func (r *ModuleReader) LocalImportPath(dir filesystemPath) (file.CorgiImportPath, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", fmt.Errorf("getting absolute path of %q: %w", dir, err)
	}

	if !strings.HasPrefix(abs, r.modFileAbs) {
		return "", fmt.Errorf("%q is not part of the module (located in %q)", dir, r.modFileAbs)
	}

	rel, err := filepath.Rel(r.modFileAbs, abs)
	if err != nil {
		return "", fmt.Errorf("computing path in module: %w", err)
	}

	return file.CorgiImportPath(path.Join(r.modFile.Module.Mod.Path, filepath.ToSlash(rel))), nil
}

func (r *ModuleReader) ReadImport(ctx context.Context, p file.CorgiImportPath) (*Package, error) {
	if strings.HasPrefix(string(p), r.modFile.Module.Mod.Path) {
		return r.readLocalImport(ctx, p)
	}
	return r.readExternalImport(ctx, p)
}

// todo: ignore files prefixed with _
// todo: respect replace directives

func (r *ModuleReader) readLocalImport(ctx context.Context, p file.CorgiImportPath) (*Package, error) {
	panic("implement me")
}

func (r *ModuleReader) readExternalImport(ctx context.Context, p file.CorgiImportPath) (*Package, error) {
	if strings.HasPrefix(string(p), "corgi/") {
		resolved := file.CorgiImportPath(path.Join(r.stdLibPath, string(p[len("corgi/"):])))
		r.logger.Info("Adjusted stdlib import",
			slog.String("old", string(p)),
			slog.String("adjusted", string(resolved)))
		p = resolved
	}
	panic("implement me")
}
