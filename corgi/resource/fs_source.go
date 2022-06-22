package resource

import (
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

// FSSource is a Source that reads files from a fs.FS.
type FSSource struct {
	name    string
	filesys fs.FS
}

func NewFSSource(name string, filesys fs.FS) *FSSource {
	return &FSSource{name: name, filesys: filesys}
}

var _ Source = (*FSSource)(nil)

func (s *FSSource) ReadCorgiFile(name string) (*File, error) {
	if !strings.HasSuffix(name, Extension) {
		name += Extension
	}

	stat, err := fs.Stat(s.filesys, name)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}

		return nil, err
	}

	if stat.IsDir() {
		return nil, nil
	}

	data, err := fs.ReadFile(s.filesys, name)
	if err != nil {
		return nil, err
	}

	return &File{
		Name:     name,
		Source:   s,
		Contents: string(data),
	}, nil
}

func (s *FSSource) ReadCorgiLib(name string) ([]File, error) {
	if strings.HasSuffix(name, Extension) {
		f, err := s.ReadCorgiFile(name)
		if err != nil {
			return nil, err
		}

		if f == nil {
			return nil, nil
		}

		return []File{*f}, nil
	}

	stat, err := fs.Stat(s.filesys, name)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			if !strings.HasSuffix(name, Extension) {
				return s.ReadCorgiLib(name + Extension)
			}

			return nil, nil
		}

		return nil, err
	}

	if !stat.IsDir() {
		f, err := s.ReadCorgiFile(name)
		if err != nil {
			return nil, err
		}

		if f == nil {
			return nil, nil
		}

		return []File{*f}, nil
	}

	dir, err := fs.ReadDir(s.filesys, name)
	if err != nil {
		return nil, err
	}

	files := make([]File, 0, len(dir))

	for _, entry := range dir {
		if entry.IsDir() {
			continue
		}

		if !strings.HasSuffix(entry.Name(), Extension) {
			continue
		}

		fileName := filepath.Join(name, entry.Name())

		f, err := s.ReadCorgiFile(fileName)
		if err != nil {
			return nil, err
		}

		if f != nil {
			files = append(files, *f)
		}
	}

	if len(files) == 0 {
		return nil, nil
	}

	return files, nil
}

func (s *FSSource) ReadFile(name string) (*File, error) {
	stat, err := fs.Stat(s.filesys, name)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			if !strings.HasSuffix(name, Extension) {
				return s.ReadCorgiFile(name + Extension)
			}

			return nil, nil
		}

		return nil, err
	}

	if stat.IsDir() {
		if !strings.HasSuffix(name, Extension) {
			return s.ReadCorgiFile(name + Extension)
		}

		return nil, nil
	}

	data, err := fs.ReadFile(s.filesys, name)
	if err != nil {
		return nil, err
	}

	return &File{
		Name:     name,
		Source:   s,
		Contents: string(data),
	}, nil
}

func (s *FSSource) Name() string {
	return s.name
}
