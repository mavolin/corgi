package mixin

import (
	"github.com/mavolin/corgi/corgi/file"
	"github.com/mavolin/corgi/corgi/parse"
)

// Checker checks that mixins are not redeclared, all init mixins are correctly
// defined, and no params or args are used twice.
type Checker struct {
	mode parse.Mode
	f    file.File
}

func NewChecker(mode parse.Mode, f file.File) *Checker {
	return &Checker{mode: mode, f: f}
}

// Check checks that...
//
// 1. ... no mixins are redeclared, i.e. that there are no mixins with the same
// name on the same level.
//
// 2. ... all init mixins are correctly defined, i.e. that they accept no
// arguments, are only present in used and extend files, and are placed at the
// top-level.
//
// 3. ... no params are defined twice.
func (c *Checker) Check() error {
	return c.checkScope(c.f.Scope, true)
}

type filePos struct {
	source string
	file   string
	line   int
	col    int
}

func (c *Checker) checkScope(s file.Scope, topLevel bool) error {
	mixinNames := make(map[file.Ident]filePos)

	for _, itm := range s {
		switch itm := itm.(type) {
		case file.Block:
			if err := c.checkScope(itm.Body, false); err != nil {
				return err
			}
		case file.Element:
			if err := c.checkScope(itm.Body, false); err != nil {
				return err
			}
		case file.Include:
			corgiIncl, ok := itm.Include.(file.CorgiInclude)
			if ok {
				if err := c.checkScope(corgiIncl.File.Scope, false); err != nil {
					return err
				}
			}
		case file.If:
			if err := c.checkScope(itm.Then, false); err != nil {
				return err
			}

			for _, ei := range itm.ElseIfs {
				if err := c.checkScope(ei.Then, false); err != nil {
					return err
				}
			}

			if itm.Else != nil {
				if err := c.checkScope(itm.Else.Then, false); err != nil {
					return err
				}
			}
		case file.Switch:
			for _, case_ := range itm.Cases {
				if err := c.checkScope(case_.Then, false); err != nil {
					return err
				}
			}

			if itm.Default != nil {
				if err := c.checkScope(itm.Default.Then, false); err != nil {
					return err
				}
			}
		case file.For:
			if err := c.checkScope(itm.Body, false); err != nil {
				return err
			}
		case file.While:
			if err := c.checkScope(itm.Body, false); err != nil {
				return err
			}
		case file.Mixin:
			if itm.Name == "init" {
				if !topLevel {
					return &InitParamsError{
						Source: c.f.Source,
						File:   c.f.Name,
						Line:   itm.Line,
						Col:    itm.Col,
					}
				}

				if err := c.checkInitMixin(itm); err != nil {
					return err
				}
			} else {
				if other, ok := mixinNames[itm.Name]; ok {
					return &RedeclaredError{
						Name:        string(itm.Name),
						Source:      c.f.Source,
						File:        c.f.Name,
						Line:        itm.Line,
						Col:         itm.Col,
						OtherSource: other.source,
						OtherFile:   other.file,
						OtherLine:   other.line,
						OtherCol:    other.col,
					}
				}

				mixinNames[itm.Name] = filePos{
					source: c.f.Source,
					file:   c.f.Name,
					line:   itm.Line,
					col:    itm.Col,
				}

				if err := c.checkDuplicateParams(itm); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (c *Checker) checkInitMixin(m file.Mixin) error {
	if len(m.Params) > 0 {
		return &InitParamsError{
			Source: c.f.Source,
			File:   c.f.Name,
			Line:   m.Line,
			Col:    m.Col,
		}
	}

	if c.mode != parse.ModeUse && c.mode != parse.ModeExtend {
		return &InitFileTypeError{
			Source: c.f.Source,
			File:   c.f.Name,
			Line:   m.Line,
			Col:    m.Col,
		}
	}

	return file.WalkError(m.Body, func(itmPtr *file.ScopeItem) (bool, error) {
		switch (*itmPtr).(type) {
		case file.If, file.Switch, file.For, file.While, file.Code:
			return true, nil
		default:
			return false, &InitBodyError{
				Source: c.f.Source,
				File:   c.f.Name,
				Line:   m.Line,
				Col:    m.Col,
			}
		}
	})
}

func (c *Checker) checkDuplicateParams(m file.Mixin) error {
	if len(m.Params) <= 1 {
		return nil
	}

	for i, param := range m.Params[:len(m.Params)-1] {
		for _, cmp := range m.Params[i+1:] {
			if param.Name == cmp.Name {
				return &DuplicateParamError{
					Name:      string(param.Name),
					Source:    c.f.Source,
					File:      c.f.Name,
					Line:      param.Line,
					Col:       param.Col,
					OtherLine: cmp.Line,
					OtherCol:  cmp.Col,
				}
			}
		}
	}

	return nil
}
