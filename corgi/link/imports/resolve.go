package imports

import "github.com/mavolin/corgi/corgi/file"

// ============================================================================
// Extend
// ======================================================================================

func (r *Resolver) resolveExtend() error {
	if r.f.Extend == nil {
		return nil
	}

	for _, imp := range r.f.Extend.File.Imports {
		if err := r.addImport(imp); err != nil {
			return err
		}
	}

	return nil
}

// ============================================================================
// Use
// ======================================================================================

func (r *Resolver) resolveUses() error {
	for _, use := range r.f.Uses {
		for _, uf := range use.Files {
			for _, imp := range uf.Imports {
				if err := r.addImport(imp); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// ============================================================================
// Include
// ======================================================================================

func (r *Resolver) resolveIncludes() error {
	return r.resolveIncludesScope(r.f.Scope)
}

func (r *Resolver) resolveIncludesScope(s file.Scope) error {
	for _, itm := range s {
		switch itm := itm.(type) {
		case file.Include:
			ci, ok := itm.Include.(file.CorgiInclude)
			if !ok {
				break
			}

			for _, imp := range ci.File.Imports {
				if err := r.addImport(imp); err != nil {
					return err
				}
			}
		case file.Block:
			if err := r.resolveIncludesScope(itm.Body); err != nil {
				return err
			}
		case file.Element:
			if err := r.resolveIncludesScope(itm.Body); err != nil {
				return err
			}
		case file.If:
			if err := r.resolveIncludesScope(itm.Then); err != nil {
				return err
			}

			for _, ei := range itm.ElseIfs {
				if err := r.resolveIncludesScope(ei.Then); err != nil {
					return err
				}
			}

			if itm.Else != nil {
				if err := r.resolveIncludesScope(itm.Else.Then); err != nil {
					return err
				}
			}
		case file.IfBlock:
			if err := r.resolveIncludesScope(itm.Then); err != nil {
				return err
			}

			if itm.Else != nil {
				if err := r.resolveIncludesScope(itm.Else.Then); err != nil {
					return err
				}
			}
		case file.Switch:
			for _, c := range itm.Cases {
				if err := r.resolveIncludesScope(c.Then); err != nil {
					return err
				}
			}

			if itm.Default != nil {
				if err := r.resolveIncludesScope(itm.Default.Then); err != nil {
					return err
				}
			}
		case file.For:
			if err := r.resolveIncludesScope(itm.Body); err != nil {
				return err
			}
		case file.While:
			if err := r.resolveIncludesScope(itm.Body); err != nil {
				return err
			}
		case file.Mixin:
			if err := r.resolveIncludesScope(itm.Body); err != nil {
				return err
			}
		case file.MixinCall:
			if err := r.resolveIncludesScope(itm.Body); err != nil {
				return err
			}
		}
	}

	return nil
}
