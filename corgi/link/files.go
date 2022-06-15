package link

import (
	"strings"

	"github.com/pkg/errors"

	"github.com/mavolin/corgi/corgi/file"
	"github.com/mavolin/corgi/corgi/file/minify"
	"github.com/mavolin/corgi/corgi/parse"
	"github.com/mavolin/corgi/corgi/resource"
)

// linkFile links all extended, used, and included files in the given linkFile
// recursively.
func (l *Linker) linkFile() error {
	if l.f.Extend != nil {
		err := l.fileExtend()
		if err != nil {
			return errors.Wrapf(err, "%s/%s:%d:%d", l.f.Source, l.f.Name, l.f.Extend.Line, l.f.Extend.Col)
		}
	}

	for i, use := range l.f.Uses {
		use := use
		err := l.fileUse(&use)
		if err != nil {
			return errors.Wrapf(err, "%s/%s:%d:%d", l.f.Source, l.f.Name, use.Line, use.Col)
		}

		l.f.Uses[i] = use
	}

	return l.fileInclude(l.f.Scope, parse.ContextRegular)
}

func (l *Linker) fileExtend() error {
	rf, err := resource.ReadCorgiFile(l.f.Extend.Path, l.rSources...)
	if err != nil {
		return err
	}

	p := parse.New(parse.ModeExtend, parse.ContextRegular, rf.Source, rf.Name, rf.Contents)
	pf, err := p.Parse()
	if err != nil {
		return err
	}

	minify.Minify(pf)

	pfLinker := New(pf)
	if err = pfLinker.Link(); err != nil {
		return err
	}

	l.f.Extend.File = *pf
	return nil
}

func (l *Linker) fileUse(use *file.Use) error {
	rFiles, err := resource.ReadCorgiLib(use.Path, l.rSources...)
	if err != nil {
		return err
	}

	parsedFiles := make([]file.File, len(rFiles))

	for i, libFile := range rFiles {
		p := parse.New(parse.ModeUse, parse.ContextRegular, libFile.Source, libFile.Name, libFile.Contents)
		pf, err := p.Parse()
		if err != nil {
			return err
		}

		minify.Minify(pf)

		parsedFiles[i] = *pf
	}

	for i, pf := range parsedFiles {
		pf := pf
		pfLinker := New(&pf)
		pfLinker.rFiles = append(parsedFiles[:i], parsedFiles[i+1:]...) //nolint:gocritic

		if err = pfLinker.Link(); err != nil {
			return err
		}

		parsedFiles[i] = pf
	}

	use.Files = parsedFiles
	return nil
}

func (l *Linker) fileInclude(s file.Scope, context parse.Context) error {
	for i, itm := range s {
		switch itm := itm.(type) {
		case file.Include:
			rf, err := resource.ReadFile(itm.Path, l.rSources...)
			if err != nil {
				return errors.Wrapf(err, "%s/%s:%d:%d", l.f.Source, l.f.Name, itm.Line, itm.Col)
			}

			if strings.HasSuffix(itm.Path, resource.Extension) {
				p := parse.New(parse.ModeInclude, context, rf.Source, rf.Name, rf.Contents)
				pf, err := p.Parse()
				if err != nil {
					return errors.Wrapf(err, "%s/%s:%d:%d", l.f.Source, l.f.Name, itm.Line, itm.Col)
				}

				pfLinker := New(pf)
				if err = pfLinker.Link(); err != nil {
					return errors.Wrapf(err, "%s/%s:%d:%d", l.f.Source, l.f.Name, itm.Line, itm.Col)
				}

				itm.Include = file.CorgiInclude{File: *pf}
			} else {
				itm.Include = file.RawInclude{Text: rf.Contents}
			}

			s[i] = itm
		case file.Block:
			if err := l.fileInclude(itm.Body, parse.ContextRegular); err != nil {
				return err
			}
		case file.Element:
			if err := l.fileInclude(itm.Body, context); err != nil {
				return err
			}
		case file.If:
			subContext := context
			if context == parse.ContextMixinCall {
				subContext = parse.ContextMixinCallConditional
			}

			if err := l.fileInclude(itm.Then, subContext); err != nil {
				return err
			}

			for _, ei := range itm.ElseIfs {
				if err := l.fileInclude(ei.Then, subContext); err != nil {
					return err
				}
			}

			if itm.Else != nil {
				if err := l.fileInclude(itm.Else.Then, subContext); err != nil {
					return err
				}
			}
		case file.IfBlock:
			subContext := context
			if context == parse.ContextMixinCall {
				subContext = parse.ContextMixinCallConditional
			}

			if err := l.fileInclude(itm.Then, subContext); err != nil {
				return err
			}

			if itm.Else != nil {
				if err := l.fileInclude(itm.Else.Then, subContext); err != nil {
					return err
				}
			}
		case file.Switch:
			subContext := context
			if context == parse.ContextMixinCall {
				subContext = parse.ContextMixinCallConditional
			}

			for _, c := range itm.Cases {
				if err := l.fileInclude(c.Then, subContext); err != nil {
					return err
				}
			}

			if itm.Default != nil {
				if err := l.fileInclude(itm.Default.Then, subContext); err != nil {
					return err
				}
			}
		case file.For:
			if err := l.fileInclude(itm.Body, context); err != nil {
				return err
			}
		case file.While:
			if err := l.fileInclude(itm.Body, context); err != nil {
				return err
			}
		case file.Mixin:
			if err := l.fileInclude(itm.Body, parse.ContextMixinDefinition); err != nil {
				return err
			}
		case file.MixinCall:
			if err := l.fileInclude(itm.Body, parse.ContextMixinCall); err != nil {
				return err
			}
		}
	}

	return nil
}
