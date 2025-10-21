package load

import (
	"context"

	"github.com/mavolin/corgi/v2/file"
	"github.com/mavolin/corgi/v2/file/diagnostic"
	"github.com/mavolin/corgi/v2/internal/cache"
)

type (
	Memoizer struct {
		c *cache.Map[file.CorgiImportPath, *computeResult]
	}

	computeResult struct {
		Package     *file.Package
		Diagnostics diagnostic.List
		Error       error
	}
)

var _ Cache = (*Memoizer)(nil)

// NewMemoizer creates a cache that memoizes the results of calls to Import so
// that subsequent calls with the same path will always return the same result.
//
// The cache of the returned loader is never cleared.
//
// Since the path given to Import contains no information about package version
// (or rather the version of the module providing it, except possibly the
// major version), special care must be placed to ensure that the cache is only
// used for imports with the same version.
// The easiest way to ensure that is to use the cache only for packages that
// originate from the same module.
func NewMemoizer() *Memoizer {
	return &Memoizer{
		c: cache.NewMap[file.CorgiImportPath, *computeResult](),
	}
}

func (l *Memoizer) Import(ctx context.Context, path file.CorgiImportPath, compute ComputeFunc) (*file.Package, diagnostic.List, error) {
	result := l.c.Get(path, func() *computeResult {
		p, d, err := compute(ctx)
		return &computeResult{p, d, err}
	})
	return result.Package, result.Diagnostics, result.Error
}
