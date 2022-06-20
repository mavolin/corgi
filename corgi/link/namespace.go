package link

func (l *Linker) checkNamespaceCollisions() error {
	if err := l.checkUseNamespaceCollisions(); err != nil {
		return err
	}

	return nil
}

func (l *Linker) checkUseNamespaceCollisions() error {
	for i, use := range l.f.Uses {
		if use.Namespace == "." {
			continue
		}

		for _, cmp := range l.f.Uses[i+1:] {
			if use.Namespace == cmp.Namespace {
				return &UseNamespaceError{
					Source:    l.f.Source,
					File:      l.f.Name,
					Line:      use.Line,
					OtherLine: cmp.Line,
					Namespace: string(use.Namespace),
				}
			}
		}
	}

	return nil
}
