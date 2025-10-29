package load

import (
	"context"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/diagnostic"
)

// Directory uses a [ModuleReader] to load the corgi package in the given
// directory.
// It must be governed by a Go module file.
//
// See [Load] for more information on the loading process or the two kinds of
// errors returned by this function.
//
// dir must use the native path separator of the operating system.
func Directory(ctx context.Context, dir filesystemPath, o Options) (*file.Package, diagnostic.List, error) {
	o.applyDefaults()

	r, err := NewModuleReader(ctx, dir, ModuleReaderOptions{
		Logger: o.Logger.WithGroup("reader"),
	})
	if err != nil {
		return nil, nil, err
	}

	imp, err := r.LocalImportPath(dir)
	if err != nil {
		return nil, nil, err
	}

	return Load(ctx, r, imp, o)
}
