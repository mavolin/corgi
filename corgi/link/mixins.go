package link

import (
	"unicode"

	"github.com/mavolin/corgi/corgi/file"
	"github.com/mavolin/corgi/corgi/parse"
)

func (l *Linker) checkMixins() error {
	if err := l.checkMixinRedeclared(); err != nil {
		return err
	}

	return l.checkMixinsScope(l.f.Scope, true)
}

func (l *Linker) checkMixinsScope(s file.Scope, topLevel bool) error {
	for _, itm := range s {
		switch itm := itm.(type) {
		case file.Block:
			if err := l.checkMixinsScope(itm.Body, false); err != nil {
				return err
			}
		case file.Element:
			if err := l.checkMixinsScope(itm.Body, false); err != nil {
				return err
			}
		case file.If:
			if err := l.checkMixinsScope(itm.Then, false); err != nil {
				return err
			}

			for _, ei := range itm.ElseIfs {
				if err := l.checkMixinsScope(ei.Then, false); err != nil {
					return err
				}
			}

			if itm.Else != nil {
				if err := l.checkMixinsScope(itm.Else.Then, false); err != nil {
					return err
				}
			}
		case file.IfBlock:
			if err := l.checkMixinsScope(itm.Then, false); err != nil {
				return err
			}

			if itm.Else != nil {
				if err := l.checkMixinsScope(itm.Else.Then, false); err != nil {
					return err
				}
			}
		case file.Switch:
			for _, c := range itm.Cases {
				if err := l.checkMixinsScope(c.Then, false); err != nil {
					return err
				}
			}

			if itm.Default != nil {
				if err := l.checkMixinsScope(itm.Default.Then, false); err != nil {
					return err
				}
			}
		case file.For:
			if err := l.checkMixinsScope(itm.Body, false); err != nil {
				return err
			}
		case file.While:
			if err := l.checkMixinsScope(itm.Body, false); err != nil {
				return err
			}
		case file.Mixin:
			if err := l.checkMixinParamsRedeclared(itm); err != nil {
				return err
			}

			if itm.Name == "init" {
				if !topLevel || l.mode != parse.ModeUse {
					return &InitError{
						Source: l.f.Source,
						File:   l.f.Name,
						Line:   itm.Line,
						Col:    itm.Col,
					}
				}

				if err := l.checkInitMixin(itm); err != nil {
					return err
				}
			}

			if err := l.checkMixinsScope(itm.Body, false); err != nil {
				return err
			}
		case file.MixinCall:
			if err := l.checkMixinsScope(itm.Body, false); err != nil {
				return err
			}
		}
	}

	return nil
}

func (l *Linker) checkMixinParamsRedeclared(m file.Mixin) error {
	params := make(map[string]filePos, len(m.Params))

	for _, param := range m.Params {
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

	return nil
}

func (l *Linker) checkInitMixin(m file.Mixin) error {
	if len(m.Params) > 0 {
		return &InitParamsError{
			Source: l.f.Source,
			File:   l.f.Name,
			Line:   m.Line,
			Col:    m.Col,
		}
	}

	for _, itm := range m.Body {
		_, isCode := itm.(file.Code)
		if !isCode {
			return &InitBodyError{
				Source: l.f.Source,
				File:   l.f.Name,
				Line:   m.Line,
				Col:    m.Col,
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

// checkMixinRedeclared checks if a mixin is redeclared anywhere within in this
// file, the top-level of a used file with namespace '.', or the top-level of a
// resource file.
func (l *Linker) checkMixinRedeclared() error {
	mixins := make(map[string]filePos)
	unexportedMixins := make(map[string]filePos)

	err := l.checkMixinRedeclaredFile(l.f, unexportedMixins, mixins)
	if err != nil {
		return err
	}

	for _, rfile := range l.rFiles {
		rfile := rfile
		err := l.checkMixinRedeclaredFile(&rfile, unexportedMixins, mixins)
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

	// child scopes
	return l.checkMixinRedeclaredScope(l.f.Scope, false)
}

func (l *Linker) checkMixinRedeclaredFile(f *file.File, unexportedMixins, mixins map[string]filePos) error {
	for _, itm := range f.Scope {
		mixin, ok := itm.(file.Mixin)
		if !ok {
			continue
		}

		if mixin.Name == "init" {
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
			if !skip && itm.Name != "init" {
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
