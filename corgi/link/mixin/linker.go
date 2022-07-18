package mixin

import (
	"unicode"

	"github.com/mavolin/corgi/corgi/file"
	"github.com/mavolin/corgi/pkg/stack"
)

// CallLinker links mixin calls to their mixin definitions.
type CallLinker struct {
	f *file.File
	// resourceFiles are the files that are in the same directory as the
	// current file.
	resourceFiles []file.File
	// scopes contains the file.Scopes that the current mixin call is in.
	//
	// It is used to resolve the mixin the call references.
	scopes stack.Stack[file.Scope]
}

// NewCallLinker creates a new CallLinker that links the mixins calls in the
// given file.
func NewCallLinker(f *file.File, resourceFiles ...file.File) *CallLinker {
	return &CallLinker{
		f:             f,
		resourceFiles: resourceFiles,
		scopes:        stack.New[file.Scope](100),
	}
}

// Link links all mixins calls.
func (l *CallLinker) Link() error {
	return l.linkScope(l.f.Scope)
}

func (l *CallLinker) linkScope(s file.Scope) error {
	l.scopes.Push(s)
	defer l.scopes.Pop()

	for i, itm := range s {
		switch itm := itm.(type) {
		case file.Block:
			if err := l.linkScope(itm.Body); err != nil {
				return err
			}
		case file.Element:
			if err := l.linkScope(itm.Body); err != nil {
				return err
			}
		case file.If:
			if err := l.linkScope(itm.Then); err != nil {
				return err
			}

			for _, ei := range itm.ElseIfs {
				if err := l.linkScope(ei.Then); err != nil {
					return err
				}
			}

			if itm.Else != nil {
				if err := l.linkScope(itm.Else.Then); err != nil {
					return err
				}
			}
		case file.IfBlock:
			if err := l.linkScope(itm.Then); err != nil {
				return err
			}

			if itm.Else != nil {
				if err := l.linkScope(itm.Else.Then); err != nil {
					return err
				}
			}
		case file.Switch:
			for _, c := range itm.Cases {
				if err := l.linkScope(c.Then); err != nil {
					return err
				}
			}

			if itm.Default != nil {
				if err := l.linkScope(itm.Default.Then); err != nil {
					return err
				}
			}
		case file.For:
			if err := l.linkScope(itm.Body); err != nil {
				return err
			}
		case file.While:
			if err := l.linkScope(itm.Body); err != nil {
				return err
			}
		case file.Mixin:
			if err := l.linkScope(itm.Body); err != nil {
				return err
			}
		case file.MixinCall:
			if err := l.linkMixinCall(&itm); err != nil {
				return err
			}

			s[i] = itm

			if err := l.linkScope(itm.Body); err != nil {
				return err
			}
		}
	}

	return nil
}

func (l *CallLinker) linkMixinCall(c *file.MixinCall) error {
	if c.Namespace != "" && unicode.IsLower(rune(c.Name[0])) {
		return &UnexportedExternalMixinError{
			Source:    l.f.Source,
			File:      l.f.Name,
			Line:      c.Line,
			Col:       c.Col,
			Namespace: string(c.Namespace),
			Name:      string(c.Name),
		}
	}

	// can't call init mixins
	if c.Name == "init" {
		return &InitCallError{
			Source: l.f.Source,
			File:   l.f.Name,
			Line:   c.Line,
			Col:    c.Col,
		}
	}

	scopes := l.scopes.Clone()

	// if no namespace, look in the current file and dot-imported used files
	if c.Namespace == "" {
		for scopes.Len() > 0 {
			s := scopes.Pop()

			for _, itm := range s {
				mixin, ok := itm.(file.Mixin)
				if !ok {
					continue
				}

				if mixin.Name == c.Name {
					c.Mixin = &mixin
					c.MixinSource = l.f.Source
					c.MixinFile = l.f.Name
					return nil
				}
			}
		}
	}

	for _, rfile := range l.resourceFiles {
		for _, itm := range rfile.Scope {
			mixin, ok := itm.(file.Mixin)
			if !ok {
				continue
			}

			if mixin.Name == c.Name {
				c.Mixin = &mixin
				c.MixinSource = rfile.Source
				c.MixinFile = rfile.Name
				return nil
			}
		}
	}

	// no namespace and unexported => don't look in used files
	if unicode.IsLower(rune(c.Name[0])) {
		return &NotFoundError{
			Source:    l.f.Source,
			File:      l.f.Name,
			Line:      c.Line,
			Col:       c.Col,
			Namespace: string(c.Namespace),
			Name:      string(c.Name),
		}
	}

Uses:
	for _, use := range l.f.Uses {
		switch {
		case use.Namespace == "." && c.Namespace == "":
			fallthrough
		case use.Namespace == c.Namespace:
			// handled below
		case use.Namespace == "_":
			continue Uses
		default:
			continue Uses
		}

		for _, uf := range use.Files {
			for _, itm := range uf.Scope {
				mixin, ok := itm.(file.Mixin)
				if !ok {
					continue
				}

				if mixin.Name == c.Name {
					c.Mixin = &mixin
					c.MixinSource = uf.Source
					c.MixinFile = uf.Name
					return nil
				}
			}
		}
	}

	return &NotFoundError{
		Source:    l.f.Source,
		File:      l.f.Name,
		Line:      c.Line,
		Col:       c.Col,
		Namespace: string(c.Namespace),
		Name:      string(c.Name),
	}
}
