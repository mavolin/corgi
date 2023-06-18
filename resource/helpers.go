package resource

// ReadCorgiFile attempts to read the corgi file with the given name, as
// described in Source.ReadCorgiFile.
//
// If none of the sources return a file, ReadCorgiFile returns a *NotFoundError.
func ReadCorgiFile(name string, sources ...Source) (*File, error) {
	for _, source := range sources {
		f, err := source.ReadCorgiFile(name)
		if err != nil {
			return nil, err
		}

		if f != nil {
			return f, nil
		}
	}

	return nil, &NotFoundError{Name: name}
}

// ReadCorgiLib attempts to read the corgi library with the given name, as
// described in Source.ReadCorgiLib.
//
// It returns the first non-nil list of Files.
// It does not combine files from different sources.
//
// If none of the sources return any files, ReadCorgiLib returns a
// *NotFoundError.
func ReadCorgiLib(name string, sources ...Source) ([]File, error) {
	for _, source := range sources {
		files, err := source.ReadCorgiLib(name)
		if err != nil {
			return nil, err
		}

		if len(files) > 0 {
			return files, nil
		}
	}

	return nil, &NotFoundError{Name: name}
}

// ReadFile attempts to read the file with the given name from the given
// sources, as described in Source.ReadFile.
//
// If none of the sources return a file, ReadFile returns a *NotFoundError.
func ReadFile(name string, sources ...Source) (*File, error) {
	for _, source := range sources {
		f, err := source.ReadFile(name)
		if err != nil {
			return nil, err
		}

		if f != nil {
			return f, nil
		}
	}

	return nil, &NotFoundError{Name: name}
}
