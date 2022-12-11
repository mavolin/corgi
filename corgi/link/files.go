package link

import (
	"strings"

	"github.com/mavolin/corgi/corgi/file"
	"github.com/mavolin/corgi/corgi/file/minify"
	"github.com/mavolin/corgi/corgi/parse"
	"github.com/mavolin/corgi/corgi/resource"
)

// fileLinker links use, linkExtend, and include directives.
type fileLinker struct {
	f               *file.File
	resourceSources []resource.Source
}

// newFileLinker creates a new fileLinker that links the given file.
func newFileLinker(f *file.File, resourceSources ...resource.Source) *fileLinker {
	return &fileLinker{f: f, resourceSources: resourceSources}
}

// link performs the linking.
func (l *fileLinker) link() error {
	if err := l.linkExtend(); err != nil {
		return err
	}

	if err := l.linkUses(); err != nil {
		return err
	}

	if err := l.linkIncludes(); err != nil {
		return err
	}

	return nil
}

// ============================================================================
// Extend
// ======================================================================================

func (l *fileLinker) linkExtend() error {
	if l.f.Extend == nil {
		return nil
	}

	err := l._linkExtend()
	if err != nil {
		return &Error{
			Source: l.f.Source,
			File:   l.f.Name,
			Line:   l.f.Extend.Line,
			Col:    l.f.Extend.Col,
			Cause:  err,
		}
	}

	return nil
}

// ugly, but best way to wrap returned errors as LinkErrors in linkExtend.
func (l *fileLinker) _linkExtend() error {
	rf, err := resource.ReadCorgiFile(l.f.Extend.Path, l.resourceSources...)
	if err != nil {
		return err
	}

	p := parse.New(parse.ModeExtend, parse.ContextRegular, rf.Source.Name(), rf.Name, rf.Contents)
	pf, err := p.Parse()
	if err != nil {
		return err
	}

	minify.Minify(pf)

	pfLinker := New(pf, parse.ModeExtend)
	pfLinker.AddResourceSource(rf.Source)

	if err = pfLinker.Link(); err != nil {
		return err
	}

	l.f.Extend.File = *pf
	return nil
}

// ============================================================================
// Use
// ======================================================================================

func (l *fileLinker) linkUses() error {
	for i, use := range l.f.Uses {
		if err := l.linkUse(&l.f.Uses[i]); err != nil {
			return &Error{
				Source: l.f.Source,
				File:   l.f.Name,
				Line:   use.Line,
				Col:    use.Col,
				Cause:  err,
			}
		}
	}

	return nil
}

func (l *fileLinker) linkUse(use *file.Use) error {
	rFiles, err := resource.ReadCorgiLib(use.Path, l.resourceSources...)
	if err != nil {
		return err
	}

	parsedFiles := make([]file.File, len(rFiles))

	for i, libFile := range rFiles {
		p := parse.New(parse.ModeUse, parse.ContextRegular, libFile.Source.Name(), libFile.Name, libFile.Contents)
		pf, err := p.Parse()
		if err != nil {
			return err
		}

		minify.Minify(pf)

		parsedFiles[i] = *pf
	}

	for i, pf := range parsedFiles {
		pf := pf
		pfLinker := New(&pf, parse.ModeUse)
		pfLinker.resourceFiles = make([]file.File, len(rFiles)-1)

		copy(pfLinker.resourceFiles, parsedFiles[:i])
		copy(pfLinker.resourceFiles[i:], parsedFiles[i+1:])

		pfLinker.AddResourceSource(rFiles[i].Source)

		if err = pfLinker.Link(); err != nil {
			return err
		}

		parsedFiles[i] = pf
	}

	use.Files = parsedFiles
	return nil
}

// ============================================================================
// Include
// ======================================================================================

func (l *fileLinker) linkIncludes() error {
	return l.linkIncludesScope(l.f.Scope, parse.ContextRegular)
}

func (l *fileLinker) linkIncludesScope(s file.Scope, pctx parse.Context) error {
	return file.WalkError(s, func(imtPtr *file.ScopeItem) (bool, error) {
		switch itm := (*imtPtr).(type) {
		case file.Include:
			if err := l.linkInclude(&itm, pctx); err != nil {
				return false, &Error{
					Source: l.f.Source,
					File:   l.f.Name,
					Line:   itm.Line,
					Col:    itm.Col,
					Cause:  err,
				}
			}

			*imtPtr = itm
			return false, nil
		case file.Block:
			if err := l.linkIncludesScope(itm.Body, parse.ContextRegular); err != nil {
				return false, err
			}

			*imtPtr = itm
			return false, nil
		case file.If:
			if err := l.linkIncludesIf(&itm, pctx); err != nil {
				return false, err
			}

			*imtPtr = itm
			return false, nil
		case file.IfBlock:
			if err := l.linkIncludesIfBlock(&itm, pctx); err != nil {
				return false, err
			}

			*imtPtr = itm
			return false, nil
		case file.Switch:
			if err := l.linkIncludesSwitch(&itm, pctx); err != nil {
				return false, err
			}

			*imtPtr = itm
			return false, nil
		case file.Mixin:
			if err := l.linkIncludesScope(itm.Body, parse.ContextMixinDefinition); err != nil {
				return false, err
			}

			*imtPtr = itm
			return false, nil
		case file.MixinCall:
			if err := l.linkIncludesScope(itm.Body, pctx); err != nil {
				return false, err
			}

			*imtPtr = itm
			return false, nil
		default:
			return true, nil
		}
	})
}

func (l *fileLinker) linkIncludesIf(if_ *file.If, pctx parse.Context) error {
	subContext := pctx
	if pctx == parse.ContextMixinCall {
		subContext = parse.ContextMixinCallConditional
	}

	if err := l.linkIncludesScope(if_.Then, subContext); err != nil {
		return err
	}

	for _, ei := range if_.ElseIfs {
		if err := l.linkIncludesScope(ei.Then, subContext); err != nil {
			return err
		}
	}

	if if_.Else != nil {
		if err := l.linkIncludesScope(if_.Else.Then, subContext); err != nil {
			return err
		}
	}

	return nil
}

func (l *fileLinker) linkIncludesIfBlock(ifBlock *file.IfBlock, pctx parse.Context) error {
	subContext := pctx
	if pctx == parse.ContextMixinCall {
		subContext = parse.ContextMixinCallConditional
	}

	if err := l.linkIncludesScope(ifBlock.Then, subContext); err != nil {
		return err
	}

	if ifBlock.Else != nil {
		if err := l.linkIncludesScope(ifBlock.Else.Then, subContext); err != nil {
			return err
		}
	}

	return nil
}

func (l *fileLinker) linkIncludesSwitch(sw *file.Switch, pctx parse.Context) error {
	subContext := pctx
	if pctx == parse.ContextMixinCall {
		subContext = parse.ContextMixinCallConditional
	}

	for _, c := range sw.Cases {
		if err := l.linkIncludesScope(c.Then, subContext); err != nil {
			return err
		}
	}

	if sw.Default != nil {
		if err := l.linkIncludesScope(sw.Default.Then, subContext); err != nil {
			return err
		}
	}

	return nil
}

func (l *fileLinker) linkInclude(incl *file.Include, pctx parse.Context) error {
	rf, err := resource.ReadFile(incl.Path, l.resourceSources...)
	if err != nil {
		return err
	}

	if !strings.HasSuffix(rf.Name, resource.Extension) {
		incl.Include = file.OtherInclude{Contents: rf.Contents}
		return nil
	}

	p := parse.New(parse.ModeInclude, pctx, rf.Source.Name(), rf.Name, rf.Contents)
	pf, err := p.Parse()
	if err != nil {
		return err
	}

	pfLinker := New(pf, parse.ModeInclude)
	pfLinker.resourceSources = l.resourceSources

	if err = pfLinker.Link(); err != nil {
		return err
	}

	incl.Include = file.CorgiInclude{File: *pf}
	return nil
}
