package loader

import (
	"sync"

	"github.com/mavolin/corgi/file"
)

type CachingLoader struct {
	l Loader

	mut               sync.Mutex
	mainFiles         map[string]*cachedFile
	templates         map[string]*cachedFile
	includes          map[string]*cachedFile
	linkedLibraries   map[string]*cachedFile
	unlinkedLibraries map[string]*cachedFile
}

var _ Loader = (*CachingLoader)(nil)

type cachedFile struct {
	f    *file.File
	incl file.IncludeFile
	lib  *file.Library
	link bool
	err  error
	done <-chan struct{}
}

// Cache wraps the passed [Loader] and returns a loader that caches results to
// the Loader's methods.
//
// It also ensures that loads for the same path are only made once.
func Cache(l Loader) *CachingLoader {
	return &CachingLoader{
		l:                 l,
		mainFiles:         make(map[string]*cachedFile),
		templates:         make(map[string]*cachedFile),
		includes:          make(map[string]*cachedFile),
		linkedLibraries:   make(map[string]*cachedFile),
		unlinkedLibraries: make(map[string]*cachedFile),
	}
}

func (l *CachingLoader) LoadLibrary(usePath string, link bool) (*file.Library, error) {
	if link {
		cached := l.load(l.linkedLibraries, func(cached *cachedFile) {
			cached.lib, cached.err = l.l.LoadLibrary(usePath, true)
		}, usePath)
		return cached.lib, cached.err
	}

	cached := l.load(l.unlinkedLibraries, func(cached *cachedFile) {
		cached.lib, cached.err = l.l.LoadLibrary(usePath, false)
	}, usePath)
	return cached.lib, cached.err
}

func (l *CachingLoader) LoadInclude(includingFile *file.File, path string) (file.IncludeFile, error) {
	cached := l.load(l.includes, func(cached *cachedFile) {
		cached.incl, cached.err = l.l.LoadInclude(includingFile, path)
	}, path)
	return cached.incl, cached.err
}

func (l *CachingLoader) LoadTemplate(extendPath string) (*file.File, error) {
	cached := l.load(l.templates, func(cached *cachedFile) {
		cached.f, cached.err = l.l.LoadTemplate(extendPath)
	}, extendPath)
	return cached.f, cached.err
}

func (l *CachingLoader) LoadMain(path string) (*file.File, error) {
	cached := l.load(l.mainFiles, func(cached *cachedFile) {
		cached.f, cached.err = l.l.LoadTemplate(path)
	}, path)
	return cached.f, cached.err
}

func (l *CachingLoader) load(m map[string]*cachedFile, loader func(*cachedFile), path string) *cachedFile {
	l.mut.Lock()
	if cached := m[path]; cached != nil {
		l.mut.Unlock()
		<-cached.done
		return cached
	}

	done := make(chan struct{})
	cached := cachedFile{done: done}
	m[path] = &cached
	l.mut.Unlock()

	loader(&cached)
	close(done)
	return &cached
}
