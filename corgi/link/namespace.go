package link

import (
	"unicode"

	"github.com/mavolin/corgi/corgi/file"
)

func (l *Linker) checkNamespaceCollisions() error {
	if err := l.checkUseNamespaceCollisions(); err != nil {
		return err
	}

	if err := l.checkMixinRedeclared(); err != nil {
		return err
	}

	if err := l.checkMixinRedeclaredScope(l.f.Scope, true); err != nil {
		return err
	}

	return l.checkMixinParamsRedeclaredScope(l.f.Scope)
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

type filePos struct {
	source string
	file   string
	line   int
	col    int
}

// checkMixinRedeclared checks if a mixin is redeclared anywhere within the
// top-level of this file, the top-level of a used file with namespace '.',
// or the top-level of a resource file.
func (l *Linker) checkMixinRedeclared() error {
	mixins := make(map[string]filePos)
	unexportedMixins := make(map[string]filePos)

	err := l.checkFileMixinRedeclared(l.f, unexportedMixins, mixins)
	if err != nil {
		return err
	}

	for _, rfile := range l.rFiles {
		rfile := rfile
		err := l.checkFileMixinRedeclared(&rfile, unexportedMixins, mixins)
		if err != nil {
			return err
		}
	}

	for _, use := range l.f.Uses {
		if use.Namespace != "." {
			continue
		}

		for _, uf := range use.Files {
			for _, itm := range uf.Scope {
				mixin, ok := itm.(file.Mixin)
				if !ok {
					continue
				}

				if !unicode.IsLower(rune(mixin.Name[0])) {
					pos, ok := mixins[string(mixin.Name)]
					if ok {
						return &MixinRedeclaredError{
							Source:      pos.source,
							File:        pos.file,
							Line:        pos.line,
							Col:         pos.col,
							OtherSource: l.f.Source,
							OtherFile:   l.f.Name,
							OtherLine:   mixin.Line,
							OtherCol:    mixin.Col,
							Name:        string(mixin.Name),
						}
					}

					mixins[string(mixin.Name)] = filePos{
						source: uf.Source,
						file:   uf.Name,
						line:   mixin.Line,
						col:    mixin.Col,
					}
				}
			}
		}
	}

	return nil
}

func (l *Linker) checkFileMixinRedeclared(f *file.File, unexportedMixins, mixins map[string]filePos) error {
	for _, itm := range f.Scope {
		mixin, ok := itm.(file.Mixin)
		if !ok {
			continue
		}

		if unicode.IsLower(rune(mixin.Name[0])) {
			pos, ok := unexportedMixins[string(mixin.Name)]
			if ok {
				return &MixinRedeclaredError{
					Source:      pos.source,
					File:        pos.file,
					Line:        pos.line,
					Col:         pos.col,
					OtherSource: f.Source,
					OtherFile:   f.Name,
					OtherLine:   mixin.Line,
					OtherCol:    mixin.Col,
					Name:        string(mixin.Name),
				}
			}

			unexportedMixins[string(mixin.Name)] = filePos{
				source: f.Source,
				file:   f.Name,
				line:   mixin.Line,
				col:    mixin.Col,
			}
		} else {
			pos, ok := mixins[string(mixin.Name)]
			if ok {
				return &MixinRedeclaredError{
					Source:      pos.source,
					File:        pos.file,
					Line:        pos.line,
					Col:         pos.col,
					OtherSource: f.Source,
					OtherFile:   f.Name,
					OtherLine:   mixin.Line,
					OtherCol:    mixin.Col,
					Name:        string(mixin.Name),
				}
			}

			mixins[string(mixin.Name)] = filePos{
				source: f.Source,
				file:   f.Name,
				line:   mixin.Line,
				col:    mixin.Col,
			}
		}
	}

	return nil
}

// checkMixinRedeclared checks if the passed scope contains redeclared mixins.
func (l *Linker) checkMixinRedeclaredScope(s file.Scope, skip bool) error {
	mixins := make(map[string]filePos)

	for _, itm := range s {
		switch itm := itm.(type) {
		case file.Block:
			if err := l.checkMixinRedeclaredScope(itm.Body, false); err != nil {
				return err
			}
		case file.Element:
			if err := l.checkMixinRedeclaredScope(itm.Body, false); err != nil {
				return err
			}
		case file.If:
			if err := l.checkMixinRedeclaredScope(itm.Then, false); err != nil {
				return err
			}

			for _, ei := range itm.ElseIfs {
				if err := l.checkMixinRedeclaredScope(ei.Then, false); err != nil {
					return err
				}
			}

			if itm.Else != nil {
				if err := l.checkMixinRedeclaredScope(itm.Else.Then, false); err != nil {
					return err
				}
			}
		case file.IfBlock:
			if err := l.checkMixinRedeclaredScope(itm.Then, false); err != nil {
				return err
			}

			if itm.Else != nil {
				if err := l.checkMixinRedeclaredScope(itm.Else.Then, false); err != nil {
					return err
				}
			}
		case file.Switch:
			for _, c := range itm.Cases {
				if err := l.checkMixinRedeclaredScope(c.Then, false); err != nil {
					return err
				}
			}

			if itm.Default != nil {
				if err := l.checkMixinRedeclaredScope(itm.Default.Then, false); err != nil {
					return err
				}
			}
		case file.For:
			if err := l.checkMixinRedeclaredScope(itm.Body, false); err != nil {
				return err
			}
		case file.While:
			if err := l.checkMixinRedeclaredScope(itm.Body, false); err != nil {
				return err
			}
		case file.Mixin:
			if !skip {
				pos, ok := mixins[string(itm.Name)]
				if ok {
					return &MixinRedeclaredError{
						Source:      pos.source,
						File:        pos.file,
						Line:        pos.line,
						Col:         pos.col,
						OtherSource: l.f.Source,
						OtherFile:   l.f.Name,
						OtherLine:   itm.Line,
						OtherCol:    itm.Col,
						Name:        string(itm.Name),
					}
				}

				mixins[string(itm.Name)] = filePos{
					source: l.f.Source,
					file:   l.f.Name,
					line:   itm.Line,
					col:    itm.Col,
				}
			}

			if err := l.checkMixinRedeclaredScope(itm.Body, false); err != nil {
				return err
			}
		case file.MixinCall:
			if err := l.checkMixinRedeclaredScope(itm.Body, false); err != nil {
				return err
			}
		}
	}

	return nil
}

func (l *Linker) checkMixinParamsRedeclaredScope(s file.Scope) error {
	for _, itm := range s {
		switch itm := itm.(type) {
		case file.Block:
			if err := l.checkMixinParamsRedeclaredScope(itm.Body); err != nil {
				return err
			}
		case file.Element:
			if err := l.checkMixinParamsRedeclaredScope(itm.Body); err != nil {
				return err
			}
		case file.If:
			if err := l.checkMixinParamsRedeclaredScope(itm.Then); err != nil {
				return err
			}

			for _, ei := range itm.ElseIfs {
				if err := l.checkMixinParamsRedeclaredScope(ei.Then); err != nil {
					return err
				}
			}

			if itm.Else != nil {
				if err := l.checkMixinParamsRedeclaredScope(itm.Else.Then); err != nil {
					return err
				}
			}
		case file.IfBlock:
			if err := l.checkMixinParamsRedeclaredScope(itm.Then); err != nil {
				return err
			}

			if itm.Else != nil {
				if err := l.checkMixinParamsRedeclaredScope(itm.Else.Then); err != nil {
					return err
				}
			}
		case file.Switch:
			for _, c := range itm.Cases {
				if err := l.checkMixinParamsRedeclaredScope(c.Then); err != nil {
					return err
				}
			}

			if itm.Default != nil {
				if err := l.checkMixinParamsRedeclaredScope(itm.Default.Then); err != nil {
					return err
				}
			}
		case file.For:
			if err := l.checkMixinParamsRedeclaredScope(itm.Body); err != nil {
				return err
			}
		case file.While:
			if err := l.checkMixinParamsRedeclaredScope(itm.Body); err != nil {
				return err
			}
		case file.Mixin:
			params := make(map[string]filePos, len(itm.Params))

			for _, param := range itm.Params {
				first, ok := params[string(param.Name)]
				if ok {
					return &DuplicateParamError{
						Source:    l.f.Source,
						File:      l.f.Name,
						Line:      first.line,
						Col:       first.col,
						OtherLine: param.Line,
						OtherCol:  param.Col,
						Name:      string(param.Name),
					}
				}

				params[string(param.Name)] = filePos{
					line: param.Line,
					col:  param.Col,
				}
			}

			if err := l.checkMixinParamsRedeclaredScope(itm.Body); err != nil {
				return err
			}
		case file.MixinCall:
			if err := l.checkMixinParamsRedeclaredScope(itm.Body); err != nil {
				return err
			}
		}
	}

	return nil
}
