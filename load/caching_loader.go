package load

import (
	"path"
	"sync"

	"github.com/mavolin/corgi/file"
)

type CachingLoader struct {
	l Loader

	mut  sync.Mutex
	pkgs map[string]*cachedPackage
}

var _ Loader = (*CachingLoader)(nil)

type cachedPackage struct {
	pkg  *file.Package
	err  error
	done <-chan struct{}
}

// Cache wraps the passed [Loader] in a loader that memoizes the results of
// calls to LoadPackage and LoadImport so that subsequent calls with the same
// path will always return the same result.
//
// This is useful to prevent redundant work when loading a file where multiple
// dependencies import the same package.
//
// Since the Cache will never be refreshed and is never cleared, it should only
// be used if files never change or for single-use cases.
func Cache(l Loader) *CachingLoader {
	return &CachingLoader{
		l:    l,
		pkgs: make(map[string]*cachedPackage),
	}
}

func (l *CachingLoader) LoadPackage(path string) (*file.Package, error) {
	return l.load(l.l.LoadPackage, path)
}

func (l *CachingLoader) LoadFiles(paths ...string) (*file.Package, error) {
	return l.l.LoadFiles(paths...)
}

func (l *CachingLoader) LoadImport(path string) (*file.Package, error) {
	return l.load(l.l.LoadImport, path)
}

func (l *CachingLoader) load(loader func(path string) (*file.Package, error), p string) (*file.Package, error) {
	l.mut.Lock()
	if cached := l.pkgs[p]; cached != nil {
		l.mut.Unlock()
		<-cached.done
		return cached.pkg, cached.err
	}

	done := make(chan struct{})
	cached := &cachedPackage{done: done}
	l.pkgs[p] = cached
	l.mut.Unlock()

	cached.pkg, cached.err = loader(p)
	close(done)

	l.mut.Lock()
	l.pkgs[cached.pkg.AbsolutePath] = cached
	l.pkgs[path.Join(cached.pkg.Module, cached.pkg.PathInModule)] = cached
	l.mut.Unlock()

	return cached.pkg, cached.err
}
