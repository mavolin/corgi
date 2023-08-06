package fileutil

import "github.com/mavolin/corgi/file"

func EqualLibrary(a, b *file.Library) bool {
	return a == b || (a != nil && a.Module == b.Module && a.PathInModule == b.PathInModule)
}

func EqualFile(a, b *file.File) bool {
	return a == b || (a != nil && a.Module == b.Module && a.PathInModule == b.PathInModule)
}
